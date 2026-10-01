package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/auth"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/store/memory"
)

type fakeHasher struct{}

func (fakeHasher) Hash(password string) (string, error) {
	if password == "" || len(password) > 256 {
		return "", errors.New("invalid password length")
	}
	return "hash:" + password, nil
}

func (fakeHasher) Verify(password, encoded string) (bool, error) {
	return encoded == "hash:"+password, nil
}

func newAuth(t *testing.T, now func() time.Time) *Auth {
	t.Helper()
	mem := memory.New()
	users, tokens, _ := memory.Repos(mem)
	issuer, err := auth.NewIssuer([]byte("01234567890123456789012345678901"), time.Minute, "go-api", now)
	if err != nil {
		t.Fatal(err)
	}
	if now == nil {
		now = time.Now
	}
	svc, err := NewAuth(users, tokens, issuer, fakeHasher{}, time.Hour, now, nil)
	if err != nil {
		t.Fatal(err)
	}
	return svc
}

func TestSignupLoginAndLogout(t *testing.T) {
	svc := newAuth(t, nil)
	ctx := context.Background()

	created, err := svc.Signup(ctx, "Ada Lovelace", "Ada@Example.com", "correct-horse")
	if err != nil {
		t.Fatal(err)
	}
	if created.User.Email != "ada@example.com" || created.AccessToken == "" || created.RefreshToken == "" {
		t.Fatalf("unexpected signup result: %+v", created.User)
	}

	if _, err := svc.Signup(ctx, "Ada Lovelace", "ada@example.com", "correct-horse"); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate signup err = %v", err)
	}

	if _, err := svc.Login(ctx, "ada@example.com", "wrong-password"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("bad password err = %v", err)
	}
	if _, err := svc.Login(ctx, "missing@example.com", "correct-horse"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("missing user err = %v", err)
	}

	logged, err := svc.Login(ctx, "ada@example.com", "correct-horse")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Logout(ctx, logged.RefreshToken); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Refresh(ctx, logged.RefreshToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("refresh after logout err = %v", err)
	}
}

func TestRefreshRotationAndReuse(t *testing.T) {
	now := time.Now().UTC()
	current := now
	svc := newAuth(t, func() time.Time { return current })
	ctx := context.Background()

	created, err := svc.Signup(ctx, "Grace Hopper", "grace@example.com", "correct-horse")
	if err != nil {
		t.Fatal(err)
	}
	first := created.RefreshToken

	rotated, err := svc.Refresh(ctx, first)
	if err != nil {
		t.Fatal(err)
	}
	if rotated.RefreshToken == first {
		t.Fatal("expected a new refresh token")
	}

	if _, err := svc.Refresh(ctx, first); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("reuse err = %v", err)
	}
	if _, err := svc.Refresh(ctx, rotated.RefreshToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("family after reuse err = %v", err)
	}

	fresh, err := svc.Login(ctx, "grace@example.com", "correct-horse")
	if err != nil {
		t.Fatal(err)
	}
	current = now.Add(2 * time.Hour)
	if _, err := svc.Refresh(ctx, fresh.RefreshToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expired refresh err = %v", err)
	}
}

func TestCurrentUser(t *testing.T) {
	svc := newAuth(t, nil)
	created, err := svc.Signup(context.Background(), "Ada", "ada2@example.com", "correct-horse")
	if err != nil {
		t.Fatal(err)
	}
	got, err := svc.Current(context.Background(), created.User.ID)
	if err != nil || got.Email != "ada2@example.com" {
		t.Fatalf("current = %+v err=%v", got, err)
	}
	if _, err := svc.Current(context.Background(), uuid.Must(uuid.NewV7())); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("missing current err = %v", err)
	}
}
