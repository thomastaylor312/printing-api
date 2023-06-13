package handlers

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog"
	"github.com/thomastaylor312/printing-api/store"
	"github.com/thomastaylor312/printing-api/types"
)

type CartHandlers struct {
	db store.DataStore
}

func NewCartHandlers(db store.DataStore) *CartHandlers {
	return &CartHandlers{db: db}
}

// GetCarts gets all carts
func (c *CartHandlers) GetCarts(w http.ResponseWriter, r *http.Request) {
	logger := httplog.LogEntry(r.Context()).With().Str("name", "carts").Logger()
	logger.Debug().Msg("Getting keys")

	keys, err := getKeys(c.db, "carts")
	if err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error getting carts: %v", err), http.StatusInternalServerError)
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
		id, _ := strconv.ParseUint(userID, 10, 64)
		cart = &types.Cart{UserID: uint(id)}
	}

	if err := json.NewEncoder(w).Encode(json.NewEncoder(w).Encode(cart)); err != nil {
		logger.Error().Err(err).Msg("Error encoding response")
	}
}

func (c *CartHandlers) PutCart(w http.ResponseWriter, r *http.Request) {
	logger := httplog.LogEntry(r.Context())
	userID := chi.URLParam(r, "userId")
	var cart types.Cart
	// Validate that we can decode the cart
	if err := json.NewDecoder(r.Body).Decode(&cart); err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error decoding: %v", err), http.StatusBadRequest)
		return
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

	// Make sure they key exists and add it if it doesn't
	keys, err := getKeys(c.db, "carts")
	if err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error getting carts: %v", err), http.StatusInternalServerError)
		return
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
			writeHttpError(r.Context(), w, fmt.Errorf("error updating cart: %v", err), http.StatusInternalServerError)
			return
		}
		if err := c.db.Set("carts", rawBuf.Bytes()); err != nil {
			writeHttpError(r.Context(), w, fmt.Errorf("error updating cart: %v", err), http.StatusInternalServerError)
			return
		}
	}

	if err := json.NewEncoder(w).Encode(cart); err != nil {
		logger.Error().Err(err).Msg("Error writing paper response")
	}
}
