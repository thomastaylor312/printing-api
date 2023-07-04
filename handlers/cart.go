package handlers

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog"
	"github.com/thomastaylor312/printing-api/store"
	"github.com/thomastaylor312/printing-api/types"
)

type CartHandlers struct {
	db     store.DataStore
	config atomic.Value
}

func NewCartHandlers(db store.DataStore, conf atomic.Value) *CartHandlers {
	return &CartHandlers{db: db, config: conf}
}

// GetCarts gets all carts
func (c *CartHandlers) GetCarts(w http.ResponseWriter, r *http.Request) {
	logger := httplog.LogEntry(r.Context()).With().Str("name", "carts").Logger()
	logger.Debug().Msg("Getting keys")

	keys, err := getKeys(c.db, "carts")
	if err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error getting cart keys: %v", err), http.StatusInternalServerError)
		return
	}

	logger.Debug().Msg("Getting data from all keys")
	allData, err := fetchByKeys[types.Cart](c.db, keys)
	if err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error getting carts: %v", err), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(allData); err != nil {
		logger.Error().Err(err).Msg("Error encoding response")
	}
}

// GetUserCart gets a user's cart. There can only ever be one cart per user.
func (c *CartHandlers) GetUserCart(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	logger := httplog.LogEntry(r.Context()).With().Str("userID", userID).Logger()

	cart, err := fetchOne[types.Cart](c.db, "carts:"+userID)
	if err != nil && !errors.Is(err, store.ErrKeyNotFound) {
		writeHttpError(r.Context(), w, fmt.Errorf("error getting cart: %v", err), http.StatusInternalServerError)
		return
	} else if err != nil {
		cart = &types.Cart{UserID: userID}
	}

	if err := json.NewEncoder(w).Encode(json.NewEncoder(w).Encode(cart)); err != nil {
		logger.Error().Err(err).Msg("Error encoding response")
	}
}

func (c *CartHandlers) PutCart(w http.ResponseWriter, r *http.Request) {
	logger := httplog.LogEntry(r.Context())
	userID := chi.URLParam(r, "userId")
	cart := types.Cart{}
	// Validate that we can decode the cart
	if err := json.NewDecoder(r.Body).Decode(&cart); err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("body is not valid JSON: %v", err), http.StatusBadRequest)
		return
	}

	for _, print := range cart.Prints {
		if err := c.normalizePrint(&print); err != nil {
			writeHttpError(r.Context(), w, err, http.StatusBadRequest)
			return
		}
	}

	rawBuf := new(bytes.Buffer)
	if err := gob.NewEncoder(rawBuf).Encode(cart); err != nil {
		// If we can't encode, that is our fault, not the user's
		writeHttpError(r.Context(), w, fmt.Errorf("error updating cart: %v", err), http.StatusInternalServerError)
		return
	}

	if err := c.db.Set("carts:"+userID, rawBuf.Bytes()); err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error updating cart: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

	if err := c.ensureCart(userID); err != nil {
		writeHttpError(r.Context(), w, err, http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(cart); err != nil {
		logger.Error().Err(err).Msg("Error writing paper response")
	}
}

// AddPrintToCart adds a single print to a user's cart
func (c *CartHandlers) AddPrintToCart(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	cartID := chi.URLParam(r, "id")
	logger := httplog.LogEntry(r.Context()).With().Str("userID", userID).Str("cartID", cartID).Logger()
	print := types.Print{}
	// Validate that we can decode the print
	if err := json.NewDecoder(r.Body).Decode(&print); err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("body is not valid JSON: %v", err), http.StatusBadRequest)
		return
	}

	logger.Debug().Msg("Validating print")

	if err := c.normalizePrint(&print); err != nil {
		writeHttpError(r.Context(), w, err, http.StatusBadRequest)
		return
	}

	logger.Debug().Msg("Getting cart")

	// Get the cart
	cart, err := fetchOne[types.Cart](c.db, "carts:"+userID)
	if err != nil && !errors.Is(err, store.ErrKeyNotFound) {
		writeHttpError(r.Context(), w, fmt.Errorf("error getting cart: %v", err), http.StatusInternalServerError)
		return
	} else if err != nil {
		cart = &types.Cart{UserID: userID}
	}

	logger.Debug().Msg("Adding print to cart")

	// Add the print to the cart
	cart.Prints = append(cart.Prints, print)
	rawBuf := new(bytes.Buffer)
	if err := gob.NewEncoder(rawBuf).Encode(cart); err != nil {
		// If we can't encode, that is our fault, not the user's
		writeHttpError(r.Context(), w, fmt.Errorf("error updating cart: %v", err), http.StatusInternalServerError)
		return
	}

	if err := c.ensureCart(userID); err != nil {
		writeHttpError(r.Context(), w, err, http.StatusInternalServerError)
		return
	}

	// Set the cart
	if err := c.db.Set("carts:"+userID, rawBuf.Bytes()); err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error updating cart: %v", err), http.StatusInternalServerError)
		return
	}

	logger.Debug().Msg("Writing response")

	if err := json.NewEncoder(w).Encode(cart); err != nil {
		logger.Error().Err(err).Msg("Error writing cart response")
	}
}

func (c *CartHandlers) ensureCart(userID string) error {
	// Make sure they key exists and add it if it doesn't
	keys, err := getKeys(c.db, "carts")
	if err != nil {
		return fmt.Errorf("error getting carts: %w", err)
	}
	contains := false
	for _, key := range keys {
		if key == "carts:"+userID {
			contains = true
			break
		}
	}

	if !contains {
		keys = append(keys, "carts:"+userID)
		rawBuf := new(bytes.Buffer)
		if err := gob.NewEncoder(rawBuf).Encode(keys); err != nil {
			return fmt.Errorf("error updating cart: %w", err)
		}
		if err := c.db.Set("carts", rawBuf.Bytes()); err != nil {
			return fmt.Errorf("error updating cart: %w", err)
		}
	}

	return nil
}

func (c *CartHandlers) normalizePrint(print *types.Print) error {
	// Fetch the paper type by ID, if the key doesn't exist, return bad request
	paper, err := fetchOne[types.PaperType](c.db, fmt.Sprintf("papers:%s", print.PaperTypeID))
	if errors.Is(err, store.ErrKeyNotFound) {
		return fmt.Errorf("invalid paper ID given")
	} else if err != nil {
		return fmt.Errorf("error validating ID: %v", err)
	}

	// Check that width and height are not greater than the configured max size
	config := c.config.Load().(*types.Config)
	if print.Width > config.MaxSize && print.Height > config.MaxSize {
		return fmt.Errorf("print is too large")
	}

	// Set the correct cost
	rawMaterialsCost := (paper.CostPerSquareInch * print.Width * print.Height) + (config.Costs.InkPerSquareInch * print.Width * print.Height)
	print.Cost = rawMaterialsCost + (rawMaterialsCost * config.Costs.DesiredProfitMargin)

	return nil
}
