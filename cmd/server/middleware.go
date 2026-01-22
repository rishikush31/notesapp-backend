// cmd/server/middleware.go
package main

import (
	"context"
	"net/http"
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
func (app *application) logRequest(next http.Handler) http.Handler {
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
		cookie, err := r.Cookie("access_token") // name of your cookie
		if err != nil {
			http.Error(w, "missing access token cookie", http.StatusUnauthorized)
			return
		}

		// app.infoLog.Println("In auth middleware, this is the access_token cookie : ", cookie.Name, cookie.Value)

		tokenStr := cookie.Value

		// app.infoLog.Println("In auth middleware, this is the access_token : ", tokenStr)

		userID, err := app.authService.ValidateAccessToken(tokenStr)

		// app.infoLog.Println("In auth middleware, User id from the access_token : ", userID)
		
		if err != nil {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), contextKeyUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
