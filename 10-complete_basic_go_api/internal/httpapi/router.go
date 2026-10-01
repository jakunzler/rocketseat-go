package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func newRouter(api *API, trustProxy bool) http.Handler {
	r := chi.NewRouter()
	if trustProxy {
		r.Use(middleware.RealIP)
	}
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(api.securityHeaders)
	r.Use(api.cors)
	r.Use(api.logger)
	r.Use(api.metricsMiddleware)
	r.Use(api.limit.middleware)
	r.Use(preflight)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		writeError(w, r, http.StatusNotFound, "not_found", "route not found", nil)
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		writeError(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
	})

	r.Get("/healthz", api.handleHealth)
	r.Get("/readyz", api.handleReady)
	r.Handle("/metrics", promhttp.HandlerFor(api.metrics.reg, promhttp.HandlerOpts{}))

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(timeout(api.requestTimeout))
		r.Get("/version", api.handleVersion)

		r.Route("/auth", func(r chi.Router) {
			r.Use(api.authLimit.middleware)
			r.Post("/signup", api.handleSignup)
			r.Post("/login", api.handleLogin)
			r.Post("/refresh", api.handleRefresh)
			r.Post("/logout", api.handleLogout)
			r.With(api.requireAuth).Post("/logout-all", api.handleLogoutAll)
		})

		r.Group(func(r chi.Router) {
			r.Use(api.requireAuth)
			r.Get("/me", api.handleMe)
			r.Get("/tasks", api.handleListTasks)
			r.Post("/tasks", api.handleCreateTask)
			r.Get("/tasks/{id}", api.handleGetTask)
			r.Patch("/tasks/{id}", api.handleUpdateTask)
			r.Delete("/tasks/{id}", api.handleDeleteTask)
		})
	})

	return r
}
