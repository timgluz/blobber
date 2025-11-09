package response

import (
	"encoding/json"
	"net/http"
)

const ContentTypeJSON = "application/json"
const ContentTypeOctetStream = "application/octet-stream"

func RenderJSON(w http.ResponseWriter, data any) error {
	w.Header().Set("Content-Type", ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(data)
}

type StatusResponse struct {
	Success bool   `json:"success"`
	Error   bool   `json:"error,omitempty"`
	Message string `json:"message"`
}

func RenderSuccessJSON(w http.ResponseWriter, message string, statusCode int) error {
	w.Header().Set("Content-Type", ContentTypeJSON)
	w.WriteHeader(statusCode)

	response := StatusResponse{
		Success: true,
		Message: message,
	}

	return json.NewEncoder(w).Encode(response)
}

func RenderErrorJSON(w http.ResponseWriter, message string, statusCode int) error {
	w.Header().Set("Content-Type", ContentTypeJSON)
	w.WriteHeader(statusCode)
	response := StatusResponse{
		Success: false,
		Error:   true,
		Message: message,
	}

	return json.NewEncoder(w).Encode(response)
}
