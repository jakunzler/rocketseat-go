package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/time/rate"
)

func TestIPLimiterBlocksAfterBurst(t *testing.T) {
	limiter := newIPLimiter(rate.Limit(0.001), 1, nil)
	handler := limiter.middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	request := func() *http.Request {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
		req.RemoteAddr = "10.0.0.8:1234"
		return req
	}

	first := httptest.NewRecorder()
	handler.ServeHTTP(first, request())
	if first.Code != http.StatusNoContent {
		t.Fatalf("first = %d", first.Code)
	}

	second := httptest.NewRecorder()
	handler.ServeHTTP(second, request())
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second = %d", second.Code)
	}
}
