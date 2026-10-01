package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/validator"
)

func decodeValid[T validator.Validator](w http.ResponseWriter, r *http.Request) (T, bool) {
	var dst T
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	media := strings.ToLower(strings.TrimSpace(strings.Split(r.Header.Get("Content-Type"), ";")[0]))
	if media != "application/json" {
		writeError(w, r, http.StatusUnsupportedMediaType, "unsupported_media_type", "content-type must be application/json", nil)
		return dst, false
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&dst); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, r, http.StatusRequestEntityTooLarge, "payload_too_large", "request body is too large", nil)
			return dst, false
		}
		writeError(w, r, http.StatusBadRequest, "malformed_json", "request body is not valid json", nil)
		return dst, false
	}
	var extra struct{}
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		writeError(w, r, http.StatusBadRequest, "malformed_json", "request body must contain a single json value", nil)
		return dst, false
	}
	if problems := dst.Valid(r.Context()); len(problems) > 0 {
		writeError(w, r, http.StatusUnprocessableEntity, "validation_error", "request validation failed", problems)
		return dst, false
	}
	return dst, true
}

func (a *API) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		const prefix = "Bearer "
		if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeError(w, r, http.StatusUnauthorized, "unauthorized", "missing bearer token", nil)
			return
		}
		userID, err := a.issuer.ParseAccess(strings.TrimSpace(header[len(prefix):]))
		if err != nil {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeError(w, r, http.StatusUnauthorized, "unauthorized", "invalid or expired token", nil)
			return
		}
		next.ServeHTTP(w, r.WithContext(withUserID(r.Context(), userID)))
	})
}
