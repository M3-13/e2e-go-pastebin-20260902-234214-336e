package api

import (
	"encoding/json"
	"errors"
	"mime"
	"net/http"

	"pastebin/internal/store"
)

const maxBodySize = 1 << 20

type createRequest struct {
	Content          string `json:"content"`
	Language         string `json:"language"`
	ExpiresInSeconds *int   `json:"expires_in_seconds"`
}

func CreateHandler(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "application/json" {
			writeError(w, http.StatusUnsupportedMediaType, "unsupported media type")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

		var req createRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				writeError(w, http.StatusRequestEntityTooLarge, "payload too large")
				return
			}
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Content == "" {
			writeError(w, http.StatusBadRequest, "content is required")
			return
		}

		if req.ExpiresInSeconds != nil && *req.ExpiresInSeconds < 0 {
			writeError(w, http.StatusBadRequest, "expires_in_seconds must not be negative")
			return
		}

		paste, err := s.Create(req.Content, req.Language, req.ExpiresInSeconds)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		writeJSON(w, http.StatusCreated, paste)
	}
}
