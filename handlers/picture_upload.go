package handlers

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

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
	url, err := p.storage.Get(userID, picture.ID())
	if err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error getting picture URL: %v", err), http.StatusInternalServerError)
		return
	}
	picture.URL = url
	if err := json.NewEncoder(w).Encode(picture); err != nil {
		logger.Error().Err(err).Msg("Error encoding response")
	}
}

// CreatePicture creates a picture in the database
func (p *PictureHandlers) CreatePicture(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")

	logger := httplog.LogEntry(r.Context()).With().Str("userID", userID).Logger()
	logger.Debug().Msg("Creating picture")
	add[*types.Picture](p.db, "pictures", w, r, nil, func(picture *types.Picture) error {
		// Add the order to the user's list of orders
		userPicturesKey := fmt.Sprintf("pictures:%s", picture.UserID)
		keys, err := getKeys(p.db, userPicturesKey)
		if err != nil {
			return fmt.Errorf("error getting current pictures: %v", err)
		}

		keys = append(keys, fmt.Sprintf("pictures:%s", picture.ID()))
		rawBuf := new(bytes.Buffer)
		if err := gob.NewEncoder(rawBuf).Encode(keys); err != nil {
			return fmt.Errorf("error adding keys: %v", err)
		}
		if err := p.db.Set(userPicturesKey, rawBuf.Bytes()); err != nil {
			return fmt.Errorf("error adding keys: %v", err)
		}

		return nil
	})
}

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
	u, err := p.storage.Set(userID, picture.ID(), uint(r.ContentLength), r.Body)
	if err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error uploading picture: %v", err), http.StatusInternalServerError)
		return
	}
	picture.URL = u

	if err := json.NewEncoder(w).Encode(picture); err != nil {
		logger.Error().Err(err).Msg("Error encoding response")
	}
}

// DeletePicture deletes a picture from the database and the bucket
func (p *PictureHandlers) DeletePicture(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	pictureID := chi.URLParam(r, "id")
	logger := httplog.LogEntry(r.Context()).With().Str("userID", userID).Str("pictureID", pictureID).Logger()
	logger.Debug().Msg("Deleting picture")
	// Decode picture data, checking that the user actually owns this picture
	_, err := p.getPicture(pictureID, userID, w, r)
	if err != nil {
		// Our helper writes the error for us
		return
	}

	// Delete the picture from the bucket
	userId := chi.URLParam(r, "userId")
	orderId := chi.URLParam(r, "id")
	logger.Debug().Msg("Deleting picture from storage")
	if err := p.storage.Delete(userId, orderId); err != nil {
		writeHttpError(r.Context(), w, fmt.Errorf("error deleting picture: %v", err), http.StatusInternalServerError)
		return
	}

	// TODO: We'll probably want to do this first, so that if it fails we don't delete the picture
	// Now delete from the database
	delete[*types.Order](p.db, "pictures", w, r, func() error {
		// Delete the picture from the list of all user pictures
		userPicturesKey := fmt.Sprintf("pictures:%s", userId)
		keys, err := getKeys(p.db, userPicturesKey)
		if err != nil {
			return fmt.Errorf("error getting keys: %v", err)
		}

		db_key := fmt.Sprintf("pictures:%s", orderId)
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
		if err := p.db.Set(userPicturesKey, rawBuf.Bytes()); err != nil {
			return fmt.Errorf("error adding keys: %v", err)
		}

		return nil
	})
}

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
	if picture.UserID != userID {
		formattedErr := fmt.Errorf("user does not have picture with specified ID")
		writeHttpError(r.Context(), w, formattedErr, http.StatusNotFound)
		return picture, formattedErr
	}
	return picture, nil
}
