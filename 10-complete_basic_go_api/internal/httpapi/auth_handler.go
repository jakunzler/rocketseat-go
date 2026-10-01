package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/service"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/store"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/validator"
)

type signupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (req signupRequest) Valid(context.Context) validator.Evaluator {
	var eval validator.Evaluator
	name := strings.TrimSpace(req.Name)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	eval.CheckField(validator.NotBlank(name) && validator.MaxChars(name, 80), "name", "must be between 1 and 80 characters")
	eval.CheckField(validator.NotBlank(email), "email", "must not be blank")
	eval.CheckField(validator.MaxChars(email, 254), "email", "must be at most 254 characters")
	eval.CheckField(validator.Matches(email, validator.EmailRX), "email", "must be a valid email")
	eval.CheckField(validator.PasswordOK(req.Password), "password", "must be between 12 and 128 characters")
	return eval
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (req loginRequest) Valid(context.Context) validator.Evaluator {
	var eval validator.Evaluator
	email := strings.ToLower(strings.TrimSpace(req.Email))
	eval.CheckField(validator.NotBlank(email) && validator.Matches(email, validator.EmailRX), "email", "must be a valid email")
	eval.CheckField(req.Password != "" && len(req.Password) <= 256, "password", "must not be blank")
	return eval
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (req refreshRequest) Valid(context.Context) validator.Evaluator {
	var eval validator.Evaluator
	eval.CheckField(validator.NotBlank(req.RefreshToken) && len(req.RefreshToken) <= 512, "refresh_token", "is invalid")
	return eval
}

type userJSON struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type tokensJSON struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

type authJSON struct {
	User   userJSON   `json:"user"`
	Tokens tokensJSON `json:"tokens"`
}

func toUserJSON(user store.User) userJSON {
	return userJSON{
		ID:        user.ID.String(),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}

func toAuthJSON(result service.Result) authJSON {
	return authJSON{
		User: toUserJSON(result.User),
		Tokens: tokensJSON{
			AccessToken:  result.AccessToken,
			RefreshToken: result.RefreshToken,
			ExpiresAt:    result.ExpiresAt,
			TokenType:    "Bearer",
		},
	}
}

func (a *API) handleSignup(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeValid[signupRequest](w, r)
	if !ok {
		return
	}
	result, err := a.authn.Signup(r.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrConflict) {
			writeError(w, r, http.StatusConflict, "conflict", "email already registered", nil)
			return
		}
		a.writeServiceError(w, r, err)
		return
	}
	writeData(w, http.StatusCreated, toAuthJSON(result))
}

func (a *API) handleLogin(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeValid[loginRequest](w, r)
	if !ok {
		return
	}
	result, err := a.authn.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		a.writeServiceError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, toAuthJSON(result))
}

func (a *API) handleRefresh(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeValid[refreshRequest](w, r)
	if !ok {
		return
	}
	result, err := a.authn.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		a.writeServiceError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, toAuthJSON(result))
}

func (a *API) handleLogout(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeValid[refreshRequest](w, r)
	if !ok {
		return
	}
	if err := a.authn.Logout(r.Context(), req.RefreshToken); err != nil {
		a.writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) handleLogoutAll(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFrom(r.Context())
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "unauthorized", "missing bearer token", nil)
		return
	}
	if err := a.authn.LogoutAll(r.Context(), userID); err != nil {
		a.writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) handleMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFrom(r.Context())
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "unauthorized", "missing bearer token", nil)
		return
	}
	user, err := a.authn.Current(r.Context(), userID)
	if err != nil {
		a.writeServiceError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, toUserJSON(user))
}
