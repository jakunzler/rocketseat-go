package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/buildinfo"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

func (a *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeData(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) handleReady(w http.ResponseWriter, r *http.Request) {
	if a.pinger != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := a.pinger.Ping(ctx); err != nil {
			a.log.Warn("readiness failed", "err", err)
			writeError(w, r, http.StatusServiceUnavailable, "unavailable", "database is not ready", nil)
			return
		}
	}
	writeData(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (a *API) handleVersion(w http.ResponseWriter, r *http.Request) {
	writeData(w, http.StatusOK, map[string]string{
		"version": buildinfo.Version,
		"commit":  buildinfo.Commit,
	})
}
