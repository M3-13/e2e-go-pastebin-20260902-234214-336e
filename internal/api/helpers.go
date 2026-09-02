package api

import (
	"encoding/json"
	"net/http"
	"regexp"
)

var idPattern = regexp.MustCompile(`^[0-9a-f]{32}$`)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func isValidID(id string) bool {
	return idPattern.MatchString(id)
}
