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
)

type IDManager interface {
	ID() uint
	SetID(id uint)
}

// ValidationFunc is a function that validates a type and returns an error and HTTP status code if
// the validation fails
type ValidationFunc[T any] func(T) (int, error)

func get[T IDManager](db store.DataStore, name string, w http.ResponseWriter, r *http.Request) {
	logger := httplog.LogEntry(r.Context()).With().Str("name", name).Logger()
	logger.Debug().Msg("Getting keys")
	data, err := db.Get(name)
	if err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error getting %s: %v", name, err), http.StatusInternalServerError)
		return
	}
	// Decode data into a list of string keys
	var keys []string
	if err != nil && !errors.Is(err, store.ErrKeyNotFound) {
		writeHttpError(r.Context(), w, fmt.Errorf("error getting: %v", err), http.StatusInternalServerError)
		return
	} else if err != nil {
		keys = make([]string, 0)
	} else {
		if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&keys); err != nil {
			writeHttpError(r.Context(), w, fmt.Errorf("error decoding : %v", err), http.StatusInternalServerError)
			return
		}
	}

	logger.Debug().Msg("Getting data from all keys")
	allData, err := fetchByKeys[T](db, keys)
	if err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error getting %s: %v", name, err), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(allData); err != nil {
		logger.Error().Err(err).Msg("Error encoding response")
	}
}

func add[T IDManager](db store.DataStore, name string, w http.ResponseWriter, r *http.Request, validation ValidationFunc[T], additionalUpdate func(T) error) {
	logger := httplog.LogEntry(r.Context())
	var userData T

	if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error decoding: %v", err), http.StatusBadRequest)
		return
	}
	// TODO: Validate the type (will need to modify generic)
	id, err := db.GenerateId()
	if err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error generating id: %v", err), http.StatusInternalServerError)
		return
	}
	userData.SetID(id)
	rawBuf := new(bytes.Buffer)
	if err := gob.NewEncoder(rawBuf).Encode(userData); err != nil {
		// If we can't encode, that is our fault, not the user's
		writeHttpError(r.Context(), w, fmt.Errorf("error adding: %v", err), http.StatusInternalServerError)
		return
	}
	db_key := fmt.Sprintf("%s:%d", name, id)
	if err := db.Set(db_key, rawBuf.Bytes()); err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error adding: %v", err), http.StatusInternalServerError)
		return
	}

	// Perform any additional updates if they exist
	if additionalUpdate != nil {
		if err := additionalUpdate(userData); err != nil {
			writeHttpError(r.Context(), w, fmt.Errorf("error updating: %v", err), http.StatusInternalServerError)
			return
		}
	}

	data, err := db.Get(name)
	var keys []string
	if err != nil && !errors.Is(err, store.ErrKeyNotFound) {
		writeHttpError(r.Context(), w, fmt.Errorf("error getting: %v", err), http.StatusInternalServerError)
		return
	} else if err != nil {
		keys = make([]string, 1)
	} else {
		if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&keys); err != nil {
			writeHttpError(r.Context(), w, fmt.Errorf("error decoding : %v", err), http.StatusInternalServerError)
			return
		}
	}

	keys = append(keys, db_key)
	rawBuf = new(bytes.Buffer)
	if err := gob.NewEncoder(rawBuf).Encode(keys); err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error adding keys: %v", err), http.StatusInternalServerError)
		return
	}
	if err := db.Set(name, rawBuf.Bytes()); err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error adding keys: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(userData); err != nil {
		logger.Error().Err(err).Msg("Error writing response")
	}
}

func update[T IDManager](db store.DataStore, name string, w http.ResponseWriter, r *http.Request, validation ValidationFunc[T], additionalUpdate func(T) error) {
	logger := httplog.LogEntry(r.Context())
	stringID := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(stringID, 10, 64)
	if err != nil {
		writeHttpError(r.Context(), w, errors.New("id is not a uint"), http.StatusBadRequest)
		return
	}
	var userData T
	// Validate that we can decode the paper
	if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error decoding: %v", err), http.StatusBadRequest)
		return
	}
	if userData.ID() != uint(id) {
		writeHttpError(r.Context(), w, errors.New("given item does not have an ID that matches"), http.StatusBadRequest)
		return
	}
	rawBuf := new(bytes.Buffer)
	if err := gob.NewEncoder(rawBuf).Encode(userData); err != nil {
		// If we can't encode, that is our fault, not the user's
		writeHttpError(r.Context(), w, fmt.Errorf("error adding paper: %v", err), http.StatusInternalServerError)
		return
	}

	// Perform any additional updates if they exist
	if additionalUpdate != nil {
		if err := additionalUpdate(userData); err != nil {
			writeHttpError(r.Context(), w, fmt.Errorf("error updating: %v", err), http.StatusInternalServerError)
			return
		}
	}

	db_key := fmt.Sprintf("%s:%d", name, userData.ID())
	if err := db.Set(db_key, rawBuf.Bytes()); err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error adding paper: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(userData); err != nil {
		logger.Error().Err(err).Msg("Error writing paper response")
	}
}

func delete[T IDManager](db store.DataStore, name string, w http.ResponseWriter, r *http.Request, additionalDelete func() error) {
	id := chi.URLParam(r, "id")
	db_key := fmt.Sprintf("%s:%s", name, id)

	// Get the list of items and delete from there first. If we delete the item first, and then fail
	// to delete from the list, we will have an orphaned item
	data, err := db.Get(name)
	var keys []string
	if err != nil && !errors.Is(err, store.ErrKeyNotFound) {
		writeHttpError(r.Context(), w, fmt.Errorf("error getting ids: %v", err), http.StatusInternalServerError)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	} else {
		if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&keys); err != nil {
			writeHttpError(r.Context(), w, fmt.Errorf("error decoding ids: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Remove the key from the list
	for i, key := range keys {
		if key == db_key {
			keys = append(keys[:i], keys[i+1:]...)
			break
		}
	}

	// Perform any additional deletes if they exist
	if additionalDelete != nil {
		if err := additionalDelete(); err != nil {
			writeHttpError(r.Context(), w, fmt.Errorf("error deleting: %v", err), http.StatusInternalServerError)
			return
		}
	}

	rawBuf := new(bytes.Buffer)
	if err := gob.NewEncoder(rawBuf).Encode(keys); err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error updating ids: %v", err), http.StatusInternalServerError)
		return
	}
	if err := db.Set(name, rawBuf.Bytes()); err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error updating ids: %v", err), http.StatusInternalServerError)
		return
	}

	if err := db.Delete(db_key); err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error deleting: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
