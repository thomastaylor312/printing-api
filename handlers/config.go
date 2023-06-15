package handlers

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/go-chi/httplog"
	"github.com/thomastaylor312/printing-api/store"
	"github.com/thomastaylor312/printing-api/types"
)

type ConfigHandlers struct {
	db            store.DataStore
	currentConfig atomic.Value
}

func NewConfigHandlers(db store.DataStore, config atomic.Value) *ConfigHandlers {
	return &ConfigHandlers{db: db, currentConfig: config}
}

// GetConfig gets the current configuration from the database
func (c *ConfigHandlers) GetConfig(w http.ResponseWriter, r *http.Request) {
	logger := httplog.LogEntry(r.Context())
	data, err := c.db.Get("config")
	if errors.Is(err, store.ErrKeyNotFound) {
		if err := json.NewEncoder(w).Encode(types.Config{}); err != nil {
			logger.Error().Err(err).Msg("Error encoding response")
		}
		return
	} else if err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error getting config: %v", err), http.StatusInternalServerError)
		return
	}

	var config types.Config
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&config); err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error decoding config: %v", err), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(config); err != nil {
		logger.Error().Err(err).Msg("Error encoding response")
	}
}

// PutConfig updates the current configuration in the database
func (c *ConfigHandlers) PutConfig(w http.ResponseWriter, r *http.Request) {
	logger := httplog.LogEntry(r.Context())
	var config types.Config
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error decoding config: %v", err), http.StatusInternalServerError)
		return
	}

	// TODO: Validate config data
	rawBuf := new(bytes.Buffer)
	if err := gob.NewEncoder(rawBuf).Encode(config); err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error encoding config: %v", err), http.StatusInternalServerError)
		return
	}

	if err := c.db.Set("config", rawBuf.Bytes()); err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error putting config: %v", err), http.StatusInternalServerError)
		return
	}

	// If we stored, update the current config
	c.currentConfig.Store(config)

	if err := json.NewEncoder(w).Encode(config); err != nil {
		logger.Error().Err(err).Msg("Error encoding response")
	}
}
