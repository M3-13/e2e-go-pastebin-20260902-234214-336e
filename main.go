package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"pastebin/internal/api"
	"pastebin/internal/store"
)

func main() {
	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		addr = "localhost:8080"
	}

	log.Printf("pastebin listening on %s", addr)
	if err := http.ListenAndServe(addr, newMux()); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func newMux() *http.ServeMux {
	s := store.New()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /pastes", api.CreateHandler(s))
	mux.HandleFunc("GET /pastes/{id}", api.GetHandler(s))
	mux.HandleFunc("GET /pastes", api.ListHandler(s))
	mux.HandleFunc("DELETE /pastes/{id}", api.DeleteHandler(s))
	mux.HandleFunc("GET /health", healthHandler)

	return mux
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
