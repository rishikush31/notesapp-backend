// cmd/server/routes.go
package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (app *application) routes() http.Handler {

	r := mux.NewRouter()

	// Middleware for route request logging
	r.Use(app.logRequest)

	// Health
	r.HandleFunc("/health", app.healthCheck).Methods(http.MethodGet)

	// Auth routes
	r.HandleFunc("/auth/register", app.register).Methods(http.MethodPost)
	r.HandleFunc("/auth/login", app.login).Methods(http.MethodPost)
	r.HandleFunc("/auth/google", app.googleLogin).Methods(http.MethodPost)
	r.HandleFunc("/auth/refresh", app.refreshToken).Methods(http.MethodPost)
	r.HandleFunc("/auth/logout", app.logout).Methods(http.MethodPost)

	// Protected routes make a subroute of main route and add the auth check middleware
	api := r.PathPrefix("/api").Subrouter()
	api.Use(app.authenticate)

	api.HandleFunc("/user",app.getUser).Methods(http.MethodGet)
	api.HandleFunc("/notes", app.createNote).Methods(http.MethodPost)
	api.HandleFunc("/notes", app.listNotes).Methods(http.MethodGet)
	api.HandleFunc("/notes/{id}", app.getNote).Methods(http.MethodGet)
	api.HandleFunc("/notes/{id}", app.updateNote).Methods(http.MethodPut)
	api.HandleFunc("/notes/{id}", app.deleteNote).Methods(http.MethodDelete)

	return r
}
