package app_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/app"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/config"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/service"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/store/memory"
)

type fakeHasher struct{}

func (fakeHasher) Hash(password string) (string, error) {
	return "hash:" + password, nil
}

func (fakeHasher) Verify(password, encoded string) (bool, error) {
	return encoded == "hash:"+password, nil
}

type errString string

func (e errString) Error() string { return string(e) }

const errPing errString = "down"

type downPinger struct{}

func (downPinger) Ping(context.Context) error { return errPing }

func newHandler(t *testing.T, hasher service.PasswordHasher) http.Handler {
	t.Helper()
	mem := memory.New()
	users, tokens, tasks := memory.Repos(mem)
	application, err := app.New(app.Options{
		Config: testConfig(),
		Users:  users,
		Tokens: tokens,
		Tasks:  tasks,
		Hasher: hasher,
	})
	if err != nil {
		t.Fatal(err)
	}
	return application.Handler
}

func testConfig() config.Config {
	return config.Config{
		Env:                "test",
		HTTPAddr:           ":8080",
		Store:              "memory",
		JWTSecret:          strings.Repeat("s", 32),
		JWTIssuer:          "go-api",
		AccessTokenTTL:     time.Minute,
		RefreshTokenTTL:    time.Hour,
		CORSOrigins:        []string{"http://localhost:3000"},
		RateLimitRPS:       1000,
		RateLimitBurst:     1000,
		AuthRateLimitRPS:   1000,
		AuthRateLimitBurst: 1000,
		RequestTimeout:     5 * time.Second,
		ShutdownTimeout:    time.Second,
		CleanupInterval:    time.Hour,
	}
}

func perform(handler http.Handler, method, path, token string, body any) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != nil {
		raw, _ := json.Marshal(body)
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func decodeData(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var payload struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode %s: %v", rec.Body.String(), err)
	}
	return payload.Data
}

func TestHealthSecurityAndValidation(t *testing.T) {
	handler := newHandler(t, fakeHasher{})

	rec := perform(handler, http.MethodGet, "/healthz", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("health = %d", rec.Code)
	}
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" || rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("security headers missing: %v", rec.Header())
	}

	rec = perform(handler, http.MethodPost, "/api/v1/auth/signup", "", map[string]string{
		"name": "A", "email": "bad", "password": "short",
	})
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("validation status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAuthAndTasks(t *testing.T) {
	handler := newHandler(t, fakeHasher{})
	rec := perform(handler, http.MethodPost, "/api/v1/auth/signup", "", map[string]string{
		"name": "Ada Lovelace", "email": "ada@example.com", "password": "correct-horse",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("signup = %d %s", rec.Code, rec.Body.String())
	}
	data := decodeData(t, rec)
	tokens := data["tokens"].(map[string]any)
	access := tokens["access_token"].(string)

	rec = perform(handler, http.MethodPost, "/api/v1/auth/signup", "", map[string]string{
		"name": "Ada Lovelace", "email": "ada@example.com", "password": "correct-horse",
	})
	if rec.Code != http.StatusConflict {
		t.Fatalf("duplicate = %d", rec.Code)
	}

	rec = perform(handler, http.MethodPost, "/api/v1/tasks", "", map[string]any{"title": "Learn Go"})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauth task = %d", rec.Code)
	}

	rec = perform(handler, http.MethodPost, "/api/v1/tasks", access, map[string]any{
		"title": "Learn Go", "description": "http tests", "priority": 2,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create task = %d %s", rec.Code, rec.Body.String())
	}
	if !strings.HasPrefix(rec.Header().Get("Location"), "/api/v1/tasks/") {
		t.Fatalf("location = %q", rec.Header().Get("Location"))
	}
	task := decodeData(t, rec)
	taskID := task["id"].(string)

	other := perform(handler, http.MethodPost, "/api/v1/auth/signup", "", map[string]string{
		"name": "Grace Hopper", "email": "grace@example.com", "password": "correct-horse",
	})
	otherAccess := decodeData(t, other)["tokens"].(map[string]any)["access_token"].(string)
	rec = perform(handler, http.MethodGet, "/api/v1/tasks/"+taskID, otherAccess, nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("cross get = %d %s", rec.Code, rec.Body.String())
	}

	rec = perform(handler, http.MethodGet, "/api/v1/tasks?limit=10", access, nil)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Learn Go") {
		t.Fatalf("list = %d %s", rec.Code, rec.Body.String())
	}

	rec = perform(handler, http.MethodPatch, "/api/v1/tasks/"+taskID, access, map[string]any{"done": true})
	if rec.Code != http.StatusOK {
		t.Fatalf("patch = %d %s", rec.Code, rec.Body.String())
	}
	rec = perform(handler, http.MethodDelete, "/api/v1/tasks/"+taskID, access, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete = %d", rec.Code)
	}
	rec = perform(handler, http.MethodGet, "/api/v1/me", access, nil)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "ada@example.com") {
		t.Fatalf("me = %d %s", rec.Code, rec.Body.String())
	}
}

func TestReadyzFailure(t *testing.T) {
	mem := memory.New()
	users, tokens, tasks := memory.Repos(mem)
	application, err := app.New(app.Options{
		Config: testConfig(),
		Users:  users,
		Tokens: tokens,
		Tasks:  tasks,
		Hasher: fakeHasher{},
		Pinger: downPinger{},
	})
	if err != nil {
		t.Fatal(err)
	}
	rec := perform(application.Handler, http.MethodGet, "/readyz", "", nil)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("readyz = %d", rec.Code)
	}
}
