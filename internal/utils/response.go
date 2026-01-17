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
