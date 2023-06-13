package store

import (
	"errors"
	"io"
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
	GenerateId() (uint, error)
}

type ImageStore interface {
	Get(container string, id string) (io.ReadCloser, error)
	Set(container string, id string, value io.ReadCloser) error
	Delete(container string, id string) error
}
