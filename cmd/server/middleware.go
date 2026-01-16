// cmd/server/middleware.go
package main

import (
	"context"
	"net/http"
	"strings"
	"time"
	"github.com/google/uuid"
)

// context keys to add in context in downstream request to conume
type contextKey string 

const (
	contextKeyUserID contextKey = "userID"
	contextKeyReqID  contextKey = "requestID"
)

// Request logging middleware
func (app *application) 	(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		reqID := uuid.NewString()

		ctx := context.WithValue(r.Context(), contextKeyReqID, reqID)
		r = r.WithContext(ctx)

		app.infoLog.Printf(
			"req_id=%s %s - %s %s %s",
			reqID,
			r.RemoteAddr,
			r.Proto,
			r.Method,
			r.URL.RequestURI(),
		)

		next.ServeHTTP(w, r)

		app.infoLog.Printf(
			"req_id=%s completed in %v",
			reqID,
			time.Since(start),
		)
	})
}

// Auth middleware 
func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			return
		}

		tokenStr := parts[1]

		userID, err := app.authService.ValidateAccessToken(tokenStr)
		if err != nil {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), contextKeyUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
