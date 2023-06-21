package handlers

import "sync/atomic"

type PricingHandlers struct {
	conf atomic.Value
}

func NewPricingHandlers(conf atomic.Value) *PricingHandlers {
	return &PricingHandlers{conf: conf}
}

// TODO: Create endpoints that calculate the price of a print job given a width and height and one that can return the current max size
