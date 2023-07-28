package payment

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/thomastaylor312/printing-api/types"
)

type Square struct {
	baseHeaders http.Header
	baseURL     url.URL
	redirectUrl url.URL
	locationID  string
}

type PaymentLinkResponse struct {
	PaymentLink struct {
		ID              string `json:"id"`
		Version         int    `json:"version"`
		OrderID         string `json:"order_id"`
		CheckoutOptions struct {
			AllowTipping          bool   `json:"allow_tipping"`
			RedirectURL           string `json:"redirect_url"`
			AskForShippingAddress bool   `json:"ask_for_shipping_address"`
			ShippingFee           struct {
				Name   string `json:"name"`
				Charge struct {
					Amount   int    `json:"amount"`
					Currency string `json:"currency"`
				} `json:"charge"`
			} `json:"shipping_fee"`
		} `json:"checkout_options"`
		URL       string    `json:"url"`
		LongURL   string    `json:"long_url"`
		CreatedAt time.Time `json:"created_at"`
	} `json:"payment_link"`
	RelatedResources struct {
		Orders []SquareOrder `json:"orders"`
	} `json:"related_resources"`
}

type OrderResponse struct {
	Order SquareOrder `json:"order"`
}

// SquareOrder is the abbreviated struct for the Square Order API, only including the fields we
// actually need
type SquareOrder struct {
	ID          string    `json:"id"`
	LocationID  string    `json:"location_id"`
	ReferenceID string    `json:"reference_id"`
	CustomerID  string    `json:"customer_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	State       string    `json:"state"`
	Version     int       `json:"version"`
}

type SquareErrorResponse struct {
	Errors []struct {
		Code     string  `json:"code"`
		Category string  `json:"category"`
		Detail   *string `json:"detail"`
		Field    *string `json:"field"`
	} `json:"errors"`
}

// NewSquare creates a new Square payment handler configured to use the provided token for
// authentication and the provided domain for the Square API. This should be _just_ the domain name,
// not the full URL. Will fail if the URL can't parse.
//
// Redirect URL is the URL that the user will be redirected to after they complete their order and
// location ID is the ID of the square location to use for the order
func NewSquare(token string, domain string, redirectUrl url.URL, locationID string) (*Square, error) {
	baseURL, err := url.Parse(fmt.Sprintf("https://%s/v2/", domain))
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}
	baseHeaders := http.Header{
		"Square-Version": {"2023-06-08"},
		"Authorization":  {"Bearer " + token},
	}
	return &Square{
		baseHeaders: baseHeaders,
		baseURL:     *baseURL,
		redirectUrl: redirectUrl,
		locationID:  locationID,
	}, nil
}

// NewSquareFromEnv is a helper function to create a new Square payment handler from configuration
// given by environment variable
func NewSquareFromEnv() (*Square, error) {
	// Get the payment URL from env var
	paymentURL := os.Getenv("PAYMENT_API_REDIRECT_URL")
	if paymentURL == "" {
		return nil, errors.New("PAYMENT_API_REDIRECT_URL must be set")
	}
	parsedURL, err := url.Parse(paymentURL)
	if err != nil {
		return nil, fmt.Errorf("PAYMENT_API_REDIRECT_URL must be a valid URL: %w", err)
	}

	// Get the location ID from env var
	locationID := os.Getenv("PAYMENT_API_LOCATION_ID")
	if locationID == "" {
		return nil, errors.New("PAYMENT_API_LOCATION_ID must be set")
	}

	// Get the token from env var
	token := os.Getenv("PAYMENT_API_TOKEN")
	if token == "" {
		return nil, errors.New("PAYMENT_API_TOKEN must be set")
	}

	// Get the domain from env var
	domain := os.Getenv("PAYMENT_API_DOMAIN")
	if domain == "" {
		return nil, errors.New("PAYMENT_API_DOMAIN must be set")
	}

	return NewSquare(token, domain, *parsedURL, locationID)
}

func (s *Square) CreateOrder(order types.Order) (externalOrderID string, paymentLink *url.URL, err error) {
	// Set up the request body
	lineItems := make([]map[string]interface{}, len(order.Prints))
	for i, print := range order.Prints {
		lineItems[i] = map[string]interface{}{
			"quantity": "1",
			"base_price_money": map[string]interface{}{
				"amount":   print.Cost,
				"currency": "USD",
			},
			"item_type": "ITEM",
			"name":      "Custom Print",
		}
	}

	requestBody, err := json.Marshal(map[string]interface{}{
		"idempotency_key": order.ID(),
		"checkout_options": map[string]interface{}{
			"allow_tipping":            false,
			"ask_for_shipping_address": true,
			"shipping_fee": map[string]interface{}{
				"charge": map[string]interface{}{
					"amount":   order.ShippingDetails.ShippingProfile.Cost,
					"currency": "USD",
				},
				"name": order.ShippingDetails.ShippingProfile.Name,
			},
			"redirect_url": s.redirectUrl.String(),
		},
		"order": map[string]interface{}{
			"location_id":  s.locationID,
			"customer_id":  order.UserID,
			"line_items":   lineItems,
			"reference_id": "internal-id",
		},
	})
	if err != nil {
		return "", nil, fmt.Errorf("failed to create order page: %w", err)
	}

	headers := s.baseHeaders.Clone()
	headers.Set("Content-Type", "application/json")
	// Set up the request
	req := http.Request{
		Method: http.MethodPost,
		URL:    s.baseURL.JoinPath("online-checkout/payment-links"),
		Body:   io.NopCloser(bytes.NewReader(requestBody)),
		Header: headers,
	}

	resp, err := http.DefaultClient.Do(&req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create order page: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp SquareErrorResponse
		// Ignore the error if it fails to decode as we can't do anything about it. Just log it
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			// TODO: Set up zerolog outside of http requests
			log.Printf("failed to decode error response from square: %s", err)
		}
		return "", nil, fmt.Errorf("failed to create order page. Got status code %d with errors: %v", resp.StatusCode, errorResp.Errors)
	}

	var paymentLinkResp PaymentLinkResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentLinkResp); err != nil {
		return "", nil, fmt.Errorf("failed to create order page: %w", err)
	}
	if len(paymentLinkResp.RelatedResources.Orders) != 1 {
		return "", nil, fmt.Errorf("failed to create order page: expected 1 order, got %d", len(paymentLinkResp.RelatedResources.Orders))
	}
	externalOrderID = paymentLinkResp.RelatedResources.Orders[0].ID
	paymentLink, err = url.Parse(paymentLinkResp.PaymentLink.URL)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create order page: %w", err)
	}
	return externalOrderID, paymentLink, nil
}

func (s *Square) ValidateOrderPaid(orderID string) (bool, error) {
	headers := s.baseHeaders.Clone()
	headers.Set("Content-Type", "application/json")
	// Set up the request
	req := http.Request{
		Method: http.MethodGet,
		URL:    s.baseURL.JoinPath("orders/" + orderID),
		Header: headers,
	}

	resp, err := http.DefaultClient.Do(&req)
	if err != nil {
		return false, fmt.Errorf("failed to validate order: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp SquareErrorResponse
		// Ignore the error if it fails to decode as we can't do anything about it. Just log it
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			log.Printf("failed to decode error response from square: %s", err)
		}
		return false, fmt.Errorf("failed to validate order. Got status code %d with errors: %v", resp.StatusCode, errorResp.Errors)
	}

	var orderResp OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&orderResp); err != nil {
		return false, fmt.Errorf("failed to validate order: %w", err)
	}

	return orderResp.Order.State == "COMPLETED", nil
}
