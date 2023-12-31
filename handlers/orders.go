package handlers

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sync/atomic"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog"
	"github.com/thomastaylor312/printing-api/payment"
	"github.com/thomastaylor312/printing-api/store"
	"github.com/thomastaylor312/printing-api/types"
)

type OrderHandlers struct {
	db      store.DataStore
	conf    atomic.Value
	payment payment.Payment
}

func NewOrderHandlers(db store.DataStore, conf atomic.Value, payment payment.Payment) *OrderHandlers {
	return &OrderHandlers{db: db, conf: conf, payment: payment}
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

	// Get the user's cart
	cart, err := fetchOne[types.Cart](o.db, fmt.Sprintf("carts:%s", userID))
	if errors.Is(err, store.ErrKeyNotFound) || (cart != nil && len(cart.Prints) == 0) {
		writeHttpError(r.Context(), w, fmt.Errorf("cart is empty, unable to place order"), http.StatusBadRequest)
	} else if err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error getting cart: %v", err), http.StatusInternalServerError)
		return
	}
	// Parse the body for the shipping details
	var shippingDetails types.ShippingDetails
	if err := json.NewDecoder(r.Body).Decode(&shippingDetails); err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error decoding shipping details: %v", err), http.StatusBadRequest)
		return
	}

	conf := o.conf.Load().(*types.Config)
	shippingDetails, err = normalizeShippingDetails(shippingDetails, *conf)
	if err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("shipping details are not valid: %v", err), http.StatusBadRequest)
	}

	var subtotal float64
	// Print prices were set and validated when they were added to the cart
	for _, print := range cart.Prints {
		subtotal += print.Cost
	}
	// Round the float to 2 decimal places
	subtotal = math.Round(subtotal*100) / 100

	order := &types.Order{
		UserID:          userID,
		Prints:          cart.Prints,
		ShippingDetails: shippingDetails,
		PrintsSubtotal:  subtotal,
		OrderTotal:      subtotal + shippingDetails.ShippingProfile.Cost,
	}

	// Create the order in the payment provider
	externalOrderID, checkoutURL, err := o.payment.CreateOrder(*order)
	if err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error creating order: %v", err), http.StatusInternalServerError)
		return
	}

	order.ExternalOrderID = externalOrderID
	order.PaymentLink = checkoutURL

	order, err = addOne[*types.Order](o.db, "orders", order, validateOrderFunc(userID), func(order *types.Order) error {
		// Add the order to the user's list of orders
		userOrdersKey := fmt.Sprintf("orders:%s", order.UserID)
		keys, err := getKeys(o.db, userOrdersKey)
		if err != nil {
			return fmt.Errorf("error getting keys: %v", err)
		}

		keys = append(keys, fmt.Sprintf("orders:%s", order.ID()))
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
		if err := o.db.Set(fmt.Sprintf("carts:%s", order.UserID), rawBuf.Bytes()); err != nil {
			logger.Warn().Err(err).Msg("Error clearing cart")
		}

		return nil
	})

	if err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error adding order to database: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(order); err != nil {
		logger.Error().Err(err).Msg("Error writing response")
	}
}

func (o *OrderHandlers) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	update[*types.Order](o.db, "orders", w, r, nil, nil)
}

// ConfirmOrderPayed is a user-facing endpoint that is called when the user has payed for their
// order. This should validate that the order was payed and update it accordingly.
func (o *OrderHandlers) ConfirmOrderPayed(w http.ResponseWriter, r *http.Request) {
	orderId := chi.URLParam(r, "id")
	order, err := fetchOne[types.Order](o.db, fmt.Sprintf("orders:%s", orderId))
	if errors.Is(err, store.ErrKeyNotFound) {
		writeHttpError(r.Context(), w, fmt.Errorf("order not found"), http.StatusNotFound)
		return
	} else if err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error getting order: %v", err), http.StatusInternalServerError)
		return
	}

	paid, err := o.payment.ValidateOrderPaid(order.ExternalOrderID)
	if err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error validating order payment: %v", err), http.StatusInternalServerError)
		return
	} else if !paid {
		writeHttpError(r.Context(), w, fmt.Errorf("order has not been paid"), http.StatusBadRequest)
		return
	}

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

func validateOrderFunc(currentUserID string) ValidationFunc[*types.Order] {
	return func(order *types.Order) (int, error) {
		// We shouldn't get here ever because we are validating the owner has this path, but just in
		// case, we check
		if order.UserID != currentUserID {
			return http.StatusForbidden, fmt.Errorf("user %s cannot create order for another user", currentUserID)
		}
		// TODO: Additional validation around shipping profile, calculated cost, etc.
		return 0, nil
	}
}

func normalizeShippingDetails(details types.ShippingDetails, conf types.Config) (types.ShippingDetails, error) {
	if details.TrackingNumber != nil {
		return types.ShippingDetails{}, fmt.Errorf("tracking number cannot be set when creating an order")
	}
	shippingMethod := details.ShippingProfile.ShippingMethod
	var shippingProfile *types.ShippingProfile
	for _, profile := range conf.Costs.ShippingProfiles {
		if shippingMethod == profile.ShippingMethod {
			shippingProfile = &profile
			break
		}
	}
	if shippingProfile == nil {
		return types.ShippingDetails{}, fmt.Errorf("invalid shipping method: %s", shippingMethod)
	}
	details.ShippingProfile = *shippingProfile

	return details, nil
}
