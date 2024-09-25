package services

import (
	"encoding/json"
	"net/http"
)

func WriteResponseSuccess (w *http.ResponseWriter, body any) {

	writer := *w 
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(writer).Encode(body); err != nil {
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
	}
}

func WriteResponseErr (w *http.ResponseWriter, message string, statusCode int) {
	writer := *w
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
	errResponse := map[string]string{"error":message}
	if err := json.NewEncoder(writer).Encode(errResponse); err != nil {
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
	}
}