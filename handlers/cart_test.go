package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"github.com/thomastaylor312/printing-api/handlers"
	"github.com/thomastaylor312/printing-api/store"
	"github.com/thomastaylor312/printing-api/types"
)

func TestHappyPath(t *testing.T) {
	tmpdir := t.TempDir()
	tmpfile := filepath.Join(tmpdir, "test.db")
	db, err := store.NewDiskDataStore(tmpfile)
	if err != nil {
		t.Fatal(err)
	}
	cartHandler := handlers.NewCartHandlers(db)
	r := chi.NewRouter()
	r.Get("/carts", cartHandler.GetCarts)
	r.Get("/carts/{userId}", cartHandler.GetUserCart)
	r.Post("/carts/{userId}", cartHandler.PutCart)
	// TODO: Actually test a full flow
}

func TestEmptyCart(t *testing.T) {
	tmpdir := t.TempDir()
	tmpfile := filepath.Join(tmpdir, "test.db")
	db, err := store.NewDiskDataStore(tmpfile)
	require.NoError(t, err)

	cartHandler := handlers.NewCartHandlers(db)
	r := chi.NewRouter()
	r.Get("/carts/{userId}", cartHandler.GetUserCart)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/carts/1", nil)
	r.ServeHTTP(recorder, req)

	// Check that the response is a 200
	require.Equal(t, http.StatusOK, recorder.Code, "expected status code 200, got %d: %s", recorder.Code, recorder.Body)

	// Check that we can deserialize the response
	var cart types.Cart
	err = json.NewDecoder(recorder.Body).Decode(&cart)
	require.NoError(t, err)
	require.Equal(t, types.Cart{UserID: 1}, cart)
}
