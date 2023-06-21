package handlers

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog"
	"github.com/thomastaylor312/printing-api/store"
	"github.com/thomastaylor312/printing-api/types"
)

type OrderHandlers struct {
	db store.DataStore
}

func NewOrderHandlers(db store.DataStore) *OrderHandlers {
	return &OrderHandlers{db: db}
}

// GetOrders gets all orders from the database
func (o *OrderHandlers) GetOrders(w http.ResponseWriter, r *http.Request) {
	get[*types.Order](o.db, "orders", w, r)
}

// GetOrdersByUser gets all orders from the database for a specific user
func (o *OrderHandlers) GetOrdersByUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	logger := httplog.LogEntry(r.Context()).With().Str("userID", userID).Logger()
	data, err := o.db.Get("orders:" + userID)
	if err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error getting orders: %v", err), http.StatusInternalServerError)
		return
	}
	// Decode data into a list of string keys
	var keys []string
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&keys); err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error decoding orders: %v", err), http.StatusInternalServerError)
		return
	}

	orders, err := fetchByKeys[types.Order](o.db, keys)
	if err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error getting order: %v", err), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(json.NewEncoder(w).Encode(orders)); err != nil {
		logger.Error().Err(err).Msg("Error encoding response")
	}
}

// GetOrderForUser gets a specific order for a specific user
func (o *OrderHandlers) GetOrderForUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	orderID := chi.URLParam(r, "id")
	logger := httplog.LogEntry(r.Context()).With().Str("userID", userID).Str("orderID", orderID).Logger()
	data, err := o.db.Get("orders:" + userID)
	if err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error getting orders: %v", err), http.StatusInternalServerError)
		return
	}
	// Decode data into a list of string keys
	var keys []string
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&keys); err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error decoding orders: %v", err), http.StatusInternalServerError)
		return
	}

	var orderKey string
	for _, key := range keys {
		if key == fmt.Sprintf("orders:%s", orderID) {
			orderKey = key
			break
		}
	}

	if orderKey == "" {
		writeHttpError(r.Context(), w, fmt.Errorf("order not found"), http.StatusNotFound)
		return
	}

	order, err := fetchOne[types.Order](o.db, orderKey)
	if err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error getting order: %v", err), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(order); err != nil {
		logger.Error().Err(err).Msg("Error encoding response")
	}
}

func (o *OrderHandlers) AddOrder(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	logger := httplog.LogEntry(r.Context()).With().Str("userID", userID).Logger()
	id, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error parsing user id: %v", err), http.StatusBadRequest)
		return
	}
	// TODO: Use the square checkout API to generate a payment link and return it, calculating the shipping price as well
	add[*types.Order](o.db, "orders", w, r, validateOrderFunc(uint(id)), func(order *types.Order) error {
		// Add the order to the user's list of orders
		userOrdersKey := fmt.Sprintf("orders:%d", order.UserID)
		keys, err := getKeys(o.db, userOrdersKey)
		if err != nil {
			return fmt.Errorf("error getting keys: %v", err)
		}

		keys = append(keys, fmt.Sprintf("orders:%d", order.ID()))
		rawBuf := new(bytes.Buffer)
		if err := gob.NewEncoder(rawBuf).Encode(keys); err != nil {
			return fmt.Errorf("error adding keys: %v", err)
		}
		if err := o.db.Set(userOrdersKey, rawBuf.Bytes()); err != nil {
			return fmt.Errorf("error adding keys: %v", err)
		}

		// Clear the user's cart
		rawBuf = new(bytes.Buffer)
		if err := gob.NewEncoder(rawBuf).Encode(types.Cart{}); err != nil {
			// This is ok if we don't actually succeed, it will just have stale data
			logger.Warn().Err(err).Msg("Error clearing cart")
		}
		if err := o.db.Set(fmt.Sprintf("carts:%d", order.UserID), rawBuf.Bytes()); err != nil {
			logger.Warn().Err(err).Msg("Error clearing cart")
		}

		return nil
	})
}

func (o *OrderHandlers) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	update[*types.Order](o.db, "orders", w, r, nil, nil)
}

func (o *OrderHandlers) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "userId")
	orderId := chi.URLParam(r, "id")
	delete[*types.Order](o.db, "orders", w, r, func() error {
		// Add the order to the user's list of orders
		userOrdersKey := fmt.Sprintf("orders:%s", userId)
		keys, err := getKeys(o.db, userOrdersKey)
		if err != nil {
			return fmt.Errorf("error getting keys: %v", err)
		}

		db_key := fmt.Sprintf("orders:%s", orderId)
		for i, key := range keys {
			if key == db_key {
				keys = append(keys[:i], keys[i+1:]...)
				break
			}
		}
		rawBuf := new(bytes.Buffer)
		if err := gob.NewEncoder(rawBuf).Encode(keys); err != nil {
			return fmt.Errorf("error adding keys: %v", err)
		}
		if err := o.db.Set(userOrdersKey, rawBuf.Bytes()); err != nil {
			return fmt.Errorf("error adding keys: %v", err)
		}

		return nil
	})
}

func validateOrderFunc(currentUserID uint) ValidationFunc[*types.Order] {
	return func(order *types.Order) (int, error) {
		// We shouldn't get here ever because we are validating the owner has this path, but just in
		// case, we check
		if order.UserID != currentUserID {
			return http.StatusForbidden, fmt.Errorf("user %d cannot create order for another user", currentUserID)
		}
		// TODO: Additional validation around shipping profile, calculated cost, etc.
		return 0, nil
	}
}
