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
	"github.com/thomastaylor312/printing-api/types"
)

type PictureHandlers struct {
	db      store.DataStore
	storage store.ImageStore
}

func NewPictureHandlers(db store.DataStore, storage store.ImageStore) *PictureHandlers {
	return &PictureHandlers{db: db, storage: storage}
}

// GetPictures gets all pictures from the database
func (p *PictureHandlers) GetPictures(w http.ResponseWriter, r *http.Request) {
	get[*types.Picture](p.db, "pictures", w, r)
}

// GetPicturesByUser gets all pictures from the database for a specific user
func (p *PictureHandlers) GetPicturesByUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	logger := httplog.LogEntry(r.Context()).With().Str("userID", userID).Logger()
	data, err := p.db.Get("pictures:" + userID)
	if err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error getting pictures: %v", err), http.StatusInternalServerError)
		return
	}
	// Decode data into a list of string keys
	var keys []string
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&keys); err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error decoding pictures: %v", err), http.StatusInternalServerError)
		return
	}

	pictures, err := fetchByKeys[types.Picture](p.db, keys)
	if err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error getting picture: %v", err), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(json.NewEncoder(w).Encode(pictures)); err != nil {
		logger.Error().Err(err).Msg("Error encoding response")
	}
}

// GetPictureInfo gets a picture from the database and populates the URL
func (p *PictureHandlers) GetPictureInfo(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	pictureID := chi.URLParam(r, "id")
	logger := httplog.LogEntry(r.Context()).With().Str("userID", userID).Str("pictureID", pictureID).Logger()
	logger.Debug().Msg("Getting picture info")
	// Decode picture data
	picture, err := p.getPicture(pictureID, userID, w, r)
	if err != nil {
		// Our helper writes the error for us
		return
	}
	// Get the URL for the picture
	url, err := p.storage.Get(userID, strconv.FormatUint(uint64(picture.ID()), 10))
	if err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error getting picture URL: %v", err), http.StatusInternalServerError)
		return
	}
	picture.URL = url
	if err := json.NewEncoder(w).Encode(picture); err != nil {
		logger.Error().Err(err).Msg("Error encoding response")
	}
}

// TODO: Create picture

// UploadPicture uploads a picture to the database
func (p *PictureHandlers) UploadPicture(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	pictureID := chi.URLParam(r, "id")
	logger := httplog.LogEntry(r.Context()).With().Str("userID", userID).Str("pictureID", pictureID).Logger()
	logger.Debug().Msg("Uploading picture")
	if r.ContentLength == 0 {
		writeHttpError(r.Context(), w, fmt.Errorf("Content-Length of picture must be provided"), http.StatusBadRequest)
		return
	}

	logger.Debug().Msg("Getting picture info")
	picture, err := p.getPicture(pictureID, userID, w, r)
	if err != nil {
		// Our helper writes the error for us
		return
	}
	// TODO: Detect content type and make sure it matches the content type header
	u, err := p.storage.Set(userID, strconv.FormatUint(uint64(picture.ID()), 10), uint(r.ContentLength), r.Body)
	if err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error uploading picture: %v", err), http.StatusInternalServerError)
		return
	}
	picture.URL = u

	if err := json.NewEncoder(w).Encode(picture); err != nil {
		logger.Error().Err(err).Msg("Error encoding response")
	}
}

// TODO: Delete picture

func (p *PictureHandlers) getPicture(pictureID string, userID string, w http.ResponseWriter, r *http.Request) (types.Picture, error) {
	var picture types.Picture
	data, err := p.db.Get("pictures:" + pictureID)
	if errors.Is(err, store.ErrKeyNotFound) {
		formattedErr := fmt.Errorf("picture not found")
		writeHttpError(r.Context(), w, formattedErr, http.StatusNotFound)
		return picture, formattedErr
	} else if err != nil {
		formattedErr := fmt.Errorf("error getting picture: %v", err)
		writeHttpError(r.Context(), w, formattedErr, http.StatusInternalServerError)
		return picture, formattedErr
	}

	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&picture); err != nil {
		formattedErr := fmt.Errorf("error decoding picture: %v", err)
		writeHttpError(r.Context(), w, formattedErr, http.StatusInternalServerError)
		return picture, formattedErr
	}
	if strconv.FormatUint(uint64(picture.UserID), 10) != userID {
		formattedErr := fmt.Errorf("user does not have picture with specified ID")
		writeHttpError(r.Context(), w, formattedErr, http.StatusNotFound)
		return picture, formattedErr
	}
	return picture, nil
}
