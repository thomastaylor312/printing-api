package payment

import (
	"net/url"

	"github.com/thomastaylor312/printing-api/types"
)

type Payment interface {
	// CreateOrder creates a new order with the payment provider and returns the external order ID
	// and optional URL to the payment page
	CreateOrder(order types.Order) (string, *url.URL, error)
	// ValidateOrderPaid checks if the order with the provided external ID has been paid for
	ValidateOrderPaid(externalOrderID string) (bool, error)
}
