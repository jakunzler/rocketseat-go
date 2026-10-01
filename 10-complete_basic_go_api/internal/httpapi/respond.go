package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/service"
)

type errorObject struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	RequestID string            `json:"request_id,omitempty"`
	Fields    map[string]string `json:"fields,omitempty"`
}

type errorResponse struct {
	Error errorObject `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeData(w http.ResponseWriter, status int, data any) {
	writeJSON(w, status, map[string]any{"data": data})
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string, fields map[string]string) {
	writeJSON(w, status, errorResponse{Error: errorObject{
		Code:      code,
		Message:   message,
		RequestID: middleware.GetReqID(r.Context()),
		Fields:    fields,
	}})
}

func (a *API) writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		writeError(w, r, http.StatusNotFound, "not_found", "resource not found", nil)
	case errors.Is(err, service.ErrConflict):
		writeError(w, r, http.StatusConflict, "conflict", "resource already exists", nil)
	case errors.Is(err, service.ErrInvalidCredentials):
		w.Header().Set("WWW-Authenticate", "Bearer")
		writeError(w, r, http.StatusUnauthorized, "invalid_credentials", "invalid email or password", nil)
	case errors.Is(err, service.ErrInvalidToken):
		w.Header().Set("WWW-Authenticate", "Bearer")
		writeError(w, r, http.StatusUnauthorized, "invalid_token", "invalid or expired token", nil)
	case errors.Is(err, service.ErrInvalid):
		writeError(w, r, http.StatusUnprocessableEntity, "validation_error", "request validation failed", nil)
	case errors.Is(err, r.Context().Err()):
		writeError(w, r, http.StatusGatewayTimeout, "timeout", "request timed out", nil)
	default:
		a.log.Error("request failed", "err", err, "request_id", middleware.GetReqID(r.Context()))
		writeError(w, r, http.StatusInternalServerError, "internal_error", "internal server error", nil)
	}
}
