package main

import (
	"net/http"
	"path/filepath"
	"time"

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
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.RedirectSlashes)
	r.Use(httplog.RequestLogger(logger))
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

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

			cartHandler := handlers.NewCartHandlers(db)
			r.Get("/carts/{userId}", cartHandler.GetUserCart)
			r.Put("/carts/{userId}", cartHandler.PutCart)

			orderHandler := handlers.NewOrderHandlers(db)
			r.Get("/orders/{userId}", orderHandler.GetOrdersByUser)
			r.Get("/orders/{userId}/{id}", orderHandler.GetOrderForUser)
			r.Post("/orders/{userId}", orderHandler.AddOrder)
			r.Put("/orders/{userId}/{id}", orderHandler.UpdateOrder)
		})
	})

	// Mount the admin sub-router
	r.Group(func(r chi.Router) {
		// TODO: jwt middleware: https://github.com/go-chi/jwtauth
		r.Mount("/admin/api", adminRouter(db))
		// TODO: Admin routes
	})

	http.ListenAndServe(":3333", r)
}

// A completely separate router for administrator routes
func adminRouter(db store.DataStore) http.Handler {
	r := chi.NewRouter()
	r.Use(AdminOnly)
	// TODO Admin routes
	paperHandler := handlers.NewPaperHandlers(db)
	r.Post("/papers", paperHandler.AddPaper)
	r.Put("/papers/{id}", paperHandler.UpdatePaper)
	r.Delete("/papers/{id}", paperHandler.DeletePaper)
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
