package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestIssueAndParseAccess(t *testing.T) {
	secret := []byte("01234567890123456789012345678901")
	issuer, err := NewIssuer(secret, time.Minute, "go-api", time.Now)
	if err != nil {
		t.Fatal(err)
	}

	userID := uuid.Must(uuid.NewV7())
	raw, exp, err := issuer.IssueAccess(userID)
	if err != nil {
		t.Fatal(err)
	}
	if !exp.After(time.Now()) {
		t.Fatal("expected future expiry")
	}

	got, err := issuer.ParseAccess(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got != userID {
		t.Fatalf("subject = %s, want %s", got, userID)
	}
}

func TestParseRejectsExpiredAndForeignTokens(t *testing.T) {
	secret := []byte("01234567890123456789012345678901")
	issuedAt := time.Now().Add(-2 * time.Minute)
	issuer, err := NewIssuer(secret, time.Minute, "go-api", func() time.Time { return issuedAt })
	if err != nil {
		t.Fatal(err)
	}
	raw, _, err := issuer.IssueAccess(uuid.Must(uuid.NewV7()))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := issuer.ParseAccess(raw); err == nil {
		t.Fatal("expected expired token to fail")
	}

	other, err := NewIssuer([]byte("abcdefghijklmnopqrstuvwxyz012345"), time.Minute, "go-api", time.Now)
	if err != nil {
		t.Fatal(err)
	}
	foreign, _, err := other.IssueAccess(uuid.Must(uuid.NewV7()))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := issuer.ParseAccess(foreign); err == nil {
		t.Fatal("expected foreign signature to fail")
	}

	if _, err := issuer.ParseAccess(strings.Replace(raw, ".", ".x", 1)); err == nil {
		t.Fatal("expected tampered token to fail")
	}
}

func TestNewIssuerRejectsShortSecret(t *testing.T) {
	if _, err := NewIssuer([]byte("short"), time.Minute, "go-api", nil); err == nil {
		t.Fatal("expected short secret to fail")
	}
}

func TestRefreshTokenHashIsStable(t *testing.T) {
	raw, hash, err := NewRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	if hash != HashRefreshToken(raw) {
		t.Fatal("hash mismatch")
	}
	if len(raw) < 32 || hash == raw {
		t.Fatal("unexpected refresh token encoding")
	}
}
