package main

import (
	"bytes"
	"encoding/gob"
	"go/types"
	"net/http"
	"path/filepath"
	"sync/atomic"

	"github.com/adrg/xdg"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog"
	"github.com/thomastaylor312/printing-api/handlers"
	"github.com/thomastaylor312/printing-api/store"
)

func main() {
	logger := httplog.NewLogger("httplog-example", httplog.Options{
		JSON: true,
	})

	db, err := store.NewDiskDataStore(filepath.Join(xdg.DataHome, "printing-api", "db"))
	if err != nil {
		logger.Fatal().Err(err).Msg("Error creating data store")
	}

	storage := store.NewDiskImageStore(filepath.Join(xdg.DataHome, "printing-api", "storage"))

	// Do an initial fetch of the config
	data, err := db.Get("config")
	if err != nil {
		logger.Fatal().Err(err).Msg("Error getting config information on startup")
	}
	var config types.Config
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&config); err != nil {
		logger.Fatal().Err(err).Msg("Error getting config information on startup")
	}

	conf := atomic.Value{}

	conf.Store(config)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.RedirectSlashes)
	r.Use(httplog.RequestLogger(logger))
	r.Use(middleware.Recoverer)

	// Gets the current user data for the logged in user
	r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi"))
	})

	// A login route using Chi that starts an OIDC flow
	r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// TODO: Figure out the OIDC flow we want to do and then issue a jwt with the claims we want
		_, err := oidc.NewProvider(ctx, "https://accounts.google.com")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	})

	r.Group(func(r chi.Router) {
		// TODO: jwt middleware: https://github.com/go-chi/jwtauth
		r.Route("/api", func(r chi.Router) {
			paperHandler := handlers.NewPaperHandlers(db)
			r.Get("/papers", paperHandler.GetPapers)

			cartHandler := handlers.NewCartHandlers(db, conf)
			r.Get("/carts/{userId}", cartHandler.GetUserCart)
			r.Put("/carts/{userId}", cartHandler.PutCart)
			r.Put("/carts/{userId}/print", cartHandler.AddPrintToCart)

			orderHandler := handlers.NewOrderHandlers(db, conf)
			r.Get("/orders/{userId}", orderHandler.GetOrdersByUser)
			r.Get("/orders/{userId}/{id}", orderHandler.GetOrderForUser)
			r.Post("/orders/{userId}", orderHandler.AddOrder)
			r.Put("/orders/{userId}/{id}", orderHandler.UpdateOrder)

			// For pictures, create a new group that uses the content type middleware
			r.Group(func(r chi.Router) {
				r.Use(middleware.AllowContentType("image/jpeg", "image/png", "image/tiff"))

				pictureHandler := handlers.NewPictureHandlers(db, storage)
				r.Post("/pictures/{userId}", pictureHandler.CreatePicture)
				r.Get("/pictures/{userId}", pictureHandler.GetPicturesByUser)
				r.Get("/pictures/{userId}/{id}", pictureHandler.GetPictureInfo)
				r.Put("/pictures/{userId}/{id}", pictureHandler.UploadPicture)
				r.Delete("/pictures/{userId}/{id}", pictureHandler.DeletePicture)
			})
		})
	})

	// Mount the admin sub-router
	r.Group(func(r chi.Router) {
		// TODO: jwt middleware: https://github.com/go-chi/jwtauth
		r.Mount("/admin/api", adminRouter(db, storage, conf))
		// TODO: Admin routes
	})

	http.ListenAndServe(":3333", r)
}

// A completely separate router for administrator routes
func adminRouter(db store.DataStore, storage store.ImageStore, conf atomic.Value) http.Handler {
	r := chi.NewRouter()
	r.Use(AdminOnly)

	paperHandler := handlers.NewPaperHandlers(db)
	r.Post("/papers", paperHandler.AddPaper)
	r.Put("/papers/{id}", paperHandler.UpdatePaper)
	r.Delete("/papers/{id}", paperHandler.DeletePaper)

	cartHandler := handlers.NewCartHandlers(db, conf)
	r.Get("/carts", cartHandler.GetCarts)
	r.Get("/carts/{userId}", cartHandler.GetUserCart)

	orderHandler := handlers.NewOrderHandlers(db, conf)
	r.Get("/orders", orderHandler.GetOrders)
	r.Get("/orders/{userId}", orderHandler.GetOrdersByUser)
	r.Get("/orders/{userId}/{id}", orderHandler.GetOrderForUser)
	r.Put("/orders/{userId}/{id}", orderHandler.UpdateOrder)
	r.Delete("/orders/{userId}/{id}", orderHandler.DeleteOrder)

	pictureHandler := handlers.NewPictureHandlers(db, storage)
	r.Get("/pictures", pictureHandler.GetPictures)
	r.Get("/pictures/{userId}", pictureHandler.GetPicturesByUser)
	r.Get("/pictures/{userId}/{id}", pictureHandler.GetPictureInfo)

	configHandler := handlers.NewConfigHandlers(db, conf)
	r.Get("/config", configHandler.GetConfig)
	r.Put("/config", configHandler.PutConfig)
	return r
}

func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//   ctx := r.Context()
		//   perm, ok := ctx.Value("acl.permission").(YourPermissionType)
		//   if !ok || !perm.IsAdmin() {
		// 	http.Error(w, http.StatusText(403), 403)
		// 	return
		//   }
		next.ServeHTTP(w, r)
	})
}
