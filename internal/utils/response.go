package utils

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(
	w http.ResponseWriter,
	status int,
	data interface{},
	headers http.Header,
) error {
	w.Header().Set("Content-Type", "application/json")

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func ErrorJSON(
	w http.ResponseWriter,
	status int,
	message string,
) {
	_ = WriteJSON(
		w,
		status,
		map[string]string{"error": message},
		nil,
	)
}

// SetAuthCookies sets both access and refresh token cookies
func SetAuthCookies(w http.ResponseWriter, accessToken, refreshToken string) {
    // Access token cookie (short-lived)
    http.SetCookie(w, &http.Cookie{
        Name:     "access_token",
        Value:    accessToken,
        Path:     "/",                // sent to all endpoints
        HttpOnly: true,
        Secure:   false,              // true in prod
        SameSite: http.SameSiteLaxMode,
        MaxAge:   15 * 60,            // 15 minutes
    })

    // Refresh token cookie (long-lived)
    http.SetCookie(w, &http.Cookie{
        Name:     "refresh_token",
        Value:    refreshToken,
        Path:     "/auth/refresh",    // only sent to refresh endpoint
        HttpOnly: true,
        Secure:   false,              // true in prod
        SameSite: http.SameSiteLaxMode,
        MaxAge:   168 * 60 * 60,      // 7 days
    })
}

// DeleteAuthCookies deletes both access and refresh token cookies
func DeleteAuthCookies(w http.ResponseWriter) {
    // Delete access token cookie
    http.SetCookie(w, &http.Cookie{
        Name:     "access_token",
        Value:    "",
        Path:     "/",
        HttpOnly: true,
        Secure:   false,
        SameSite: http.SameSiteLaxMode,
        MaxAge:   -1,
    })

    // Delete refresh token cookie
    http.SetCookie(w, &http.Cookie{
        Name:     "refresh_token",
        Value:    "",
        Path:     "/auth/refresh",
        HttpOnly: true,
        Secure:   false,
        SameSite: http.SameSiteLaxMode,
        MaxAge:   -1,
    })
}
