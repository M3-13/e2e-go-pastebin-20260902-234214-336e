package api

import (
	"net/http"

	"pastebin/internal/store"
)

func GetHandler(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if !isValidID(id) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}

		p, ok := s.Get(id)
		if !ok {
			writeError(w, http.StatusNotFound, "not found")
			return
		}

		writeJSON(w, http.StatusOK, p)
	}
}
