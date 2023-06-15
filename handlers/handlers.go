package handlers

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/httplog"
	"github.com/thomastaylor312/printing-api/store"
)

func writeHttpError(ctx context.Context, w http.ResponseWriter, err error, code int) {
	logger := httplog.LogEntry(ctx)
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"error": err.Error(),
	}); err != nil {
		logger.Error().Err(err).Msg("Error writing error response")
	}
}

func fetchByKeys[V any](db store.DataStore, keys []string) ([]V, error) {
	outlist := make([]V, len(keys))
	for i, key := range keys {
		data, err := db.Get(key)
		if err != nil {
			return nil, err
		}
		var out V
		if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&out); err != nil {
			return nil, fmt.Errorf("error decoding: %v", err)
		}
		outlist[i] = out
	}
	return outlist, nil
}

func fetchOne[V any](db store.DataStore, key string) (*V, error) {
	data, err := db.Get(key)
	if err != nil {
		return nil, err
	}
	var out V
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&out); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}
	return &out, nil
}

func getKeys(db store.DataStore, key string) ([]string, error) {
	data, err := db.Get(key)
	var keys []string
	if err != nil && !errors.Is(err, store.ErrKeyNotFound) {
		return nil, err
	} else if err != nil {
		return make([]string, 0), nil
	} else {
		reader := bytes.NewReader(data)
		if err := gob.NewDecoder(reader).Decode(&keys); err != nil {
			return nil, fmt.Errorf("error decoding bytes from store: %v", err)
		}
	}

	return keys, nil
}
