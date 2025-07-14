package helpers

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
    Message string `json:"message" example:"string"`
}

func WriteJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Message: message})
}
