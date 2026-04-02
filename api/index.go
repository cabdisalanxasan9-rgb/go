package handler

import (
	"encoding/json"
	"net/http"
	"time"
)

type response struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

// Handler is the Vercel serverless entrypoint.
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(response{
		Status:    "ok",
		Service:   "go-toolkit-api",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}
