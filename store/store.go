package store

import (
	"errors"
	"io"
	"net/url"
)

var (
	ErrKeyNotFound   = errors.New("key not found")
	ErrImageNotFound = errors.New("image not found")
)

// DataStore is an interface for storing and retrieving data from a key value store
type DataStore interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte) error
	Delete(key string) error
	// GenerateId generates a unique id for the store
	GenerateId() (string, error)
}

type ImageStore interface {
	// Get an image from the store, returning the URL to the image, but not the image itself. This
	// allows an implementation to do things like generate a token to a CDN
	Get(container string, id string) (*url.URL, error)
	// Set an image to the store, returning the URL to the image. This allows an implementation
	// to do things like generating a token to a CDN
	Set(container string, id string, expected_length uint, value io.ReadCloser) (*url.URL, error)
	Delete(container string, id string) error
}
