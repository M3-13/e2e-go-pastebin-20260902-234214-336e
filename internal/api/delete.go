package api

import (
	"net/http"

	"pastebin/internal/store"
)

func DeleteHandler(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if !isValidID(id) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		if !s.Delete(id) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
