package service

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/auth"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/store"
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, encoded string) (bool, error)
}

type Result struct {
	User         store.User
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

type Auth struct {
	users      store.UserRepository
	tokens     store.RefreshTokenRepository
	issuer     *auth.Issuer
	hasher     PasswordHasher
	refreshTTL time.Duration
	now        func() time.Time
	log        *slog.Logger
	dummy      string
}

func NewAuth(users store.UserRepository, tokens store.RefreshTokenRepository, issuer *auth.Issuer, hasher PasswordHasher, refreshTTL time.Duration, now func() time.Time, log *slog.Logger) (*Auth, error) {
	if now == nil {
		now = time.Now
	}
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	dummy, err := hasher.Hash("dummy-password-value")
	if err != nil {
		return nil, err
	}
	return &Auth{
		users:      users,
		tokens:     tokens,
		issuer:     issuer,
		hasher:     hasher,
		refreshTTL: refreshTTL,
		now:        now,
		log:        log,
		dummy:      dummy,
	}, nil
}

func (s *Auth) Signup(ctx context.Context, name, email, password string) (Result, error) {
	hash, err := s.hasher.Hash(password)
	if err != nil {
		return Result{}, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return Result{}, err
	}
	now := s.now().UTC()
	user, err := s.users.Create(ctx, store.User{
		ID:           id,
		Name:         strings.TrimSpace(name),
		Email:        normalizeEmail(email),
		PasswordHash: hash,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		return Result{}, mapStore(err)
	}
	return s.issue(ctx, user, nil)
}

func (s *Auth) Login(ctx context.Context, email, password string) (Result, error) {
	user, err := s.users.FindByEmail(ctx, normalizeEmail(email))
	if errors.Is(err, store.ErrNotFound) {
		_, _ = s.hasher.Verify(password, s.dummy)
		return Result{}, ErrInvalidCredentials
	}
	if err != nil {
		return Result{}, err
	}
	ok, err := s.hasher.Verify(password, user.PasswordHash)
	if err != nil {
		return Result{}, err
	}
	if !ok {
		return Result{}, ErrInvalidCredentials
	}
	return s.issue(ctx, user, nil)
}

func (s *Auth) Refresh(ctx context.Context, raw string) (Result, error) {
	if raw == "" || len(raw) > 512 {
		return Result{}, ErrInvalidToken
	}
	record, err := s.tokens.FindByHash(ctx, auth.HashRefreshToken(raw))
	if errors.Is(err, store.ErrNotFound) {
		return Result{}, ErrInvalidToken
	}
	if err != nil {
		return Result{}, err
	}
	if record.RevokedAt != nil {
		if err := s.tokens.RevokeAllForUser(ctx, record.UserID); err != nil {
			return Result{}, err
		}
		return Result{}, ErrInvalidToken
	}
	if !record.ExpiresAt.After(s.now()) {
		return Result{}, ErrInvalidToken
	}
	user, err := s.users.FindByID(ctx, record.UserID)
	if errors.Is(err, store.ErrNotFound) {
		return Result{}, ErrInvalidToken
	}
	if err != nil {
		return Result{}, err
	}
	return s.issue(ctx, user, &record.ID)
}

func (s *Auth) Logout(ctx context.Context, raw string) error {
	if raw == "" || len(raw) > 512 {
		return nil
	}
	return s.tokens.RevokeByHash(ctx, auth.HashRefreshToken(raw))
}

func (s *Auth) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.tokens.RevokeAllForUser(ctx, userID)
}

func (s *Auth) Current(ctx context.Context, id uuid.UUID) (store.User, error) {
	user, err := s.users.FindByID(ctx, id)
	if errors.Is(err, store.ErrNotFound) {
		return store.User{}, ErrInvalidToken
	}
	return user, err
}

func (s *Auth) RunCleanup(ctx context.Context, every time.Duration) error {
	s.cleanup(ctx)
	ticker := time.NewTicker(every)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			s.cleanup(ctx)
		}
	}
}

func (s *Auth) cleanup(ctx context.Context) {
	cctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := s.tokens.DeleteExpired(cctx, s.now()); err != nil && ctx.Err() == nil {
		s.log.Error("cleanup refresh tokens", "err", err)
	}
}

func (s *Auth) issue(ctx context.Context, user store.User, revokeID *uuid.UUID) (Result, error) {
	access, expires, err := s.issuer.IssueAccess(user.ID)
	if err != nil {
		return Result{}, err
	}
	raw, hash, err := auth.NewRefreshToken()
	if err != nil {
		return Result{}, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return Result{}, err
	}
	next := store.RefreshToken{
		ID:        id,
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: s.now().UTC().Add(s.refreshTTL),
	}
	if revokeID != nil {
		if err := s.tokens.Rotate(ctx, *revokeID, next); err != nil {
			if errors.Is(err, store.ErrConflict) {
				_ = s.tokens.RevokeAllForUser(ctx, user.ID)
				return Result{}, ErrInvalidToken
			}
			return Result{}, err
		}
	} else if err := s.tokens.Create(ctx, next); err != nil {
		return Result{}, mapStore(err)
	}
	return Result{
		User:         user,
		AccessToken:  access,
		RefreshToken: raw,
		ExpiresAt:    expires,
	}, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
