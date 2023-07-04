package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
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

	conf := atomic.Value{}

	conf.Store(types.Config{
		MaxSize: 17.0,
	})

	cartHandler := handlers.NewCartHandlers(db, conf)
	r := chi.NewRouter()
	r.Get("/carts", cartHandler.GetCarts)
	r.Get("/carts/{userId}", cartHandler.GetUserCart)
	r.Put("/carts/{userId}", cartHandler.PutCart)

	recorder := httptest.NewRecorder()
	require.NoError(t, err)
	cart := types.Cart{
		UserID: "1",
		Prints: []types.Print{
			{
				PictureID: "1",
				Width:     8,
				Height:    10,
			},
		},
	}
	buf := new(bytes.Buffer)
	err = json.NewEncoder(buf).Encode(cart)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPut, "/carts/1", buf)
	r.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code, "expected status code 200, got %d: %s", recorder.Code, recorder.Body)
	var returnedCart types.Cart
	err = json.NewDecoder(recorder.Body).Decode(&returnedCart)
	require.NoError(t, err)
	require.Equal(t, cart, returnedCart)

	// Get the user cart
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/carts/1", nil)
	r.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code, "expected status code 200, got %d: %s", recorder.Code, recorder.Body)
	err = json.NewDecoder(recorder.Body).Decode(&returnedCart)
	require.NoError(t, err)
	require.Equal(t, cart, returnedCart)

	// Now put an empty cart for another user
	recorder = httptest.NewRecorder()
	buf = new(bytes.Buffer)
	err = json.NewEncoder(buf).Encode(types.Cart{UserID: "2", Prints: []types.Print{}})
	require.NoError(t, err)
	req = httptest.NewRequest(http.MethodPut, "/carts/2", buf)
	r.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code, "expected status code 200, got %d: %s", recorder.Code, recorder.Body)

	// Now get all carts
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/carts", nil)
	r.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code, "expected status code 200, got %d: %s", recorder.Code, recorder.Body)
	var carts []types.Cart
	err = json.NewDecoder(recorder.Body).Decode(&carts)
	require.NoError(t, err)
	require.Equal(t, []types.Cart{cart, {UserID: "2"}}, carts)
}

func TestEmptyCart(t *testing.T) {
	tmpdir := t.TempDir()
	tmpfile := filepath.Join(tmpdir, "test.db")
	db, err := store.NewDiskDataStore(tmpfile)
	require.NoError(t, err)

	conf := atomic.Value{}

	cartHandler := handlers.NewCartHandlers(db, conf)
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
	require.Equal(t, types.Cart{UserID: "1"}, cart)
}

// TODO: Test failed verification of print size
// TODO: Test add single print to cart
