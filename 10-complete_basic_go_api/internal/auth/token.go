package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Issuer struct {
	secret []byte
	ttl    time.Duration
	issuer string
	now    func() time.Time
}

func NewIssuer(secret []byte, ttl time.Duration, issuer string, now func() time.Time) (*Issuer, error) {
	if len(secret) < 32 {
		return nil, errors.New("jwt secret must be at least 32 bytes")
	}
	if ttl <= 0 {
		return nil, errors.New("access token ttl must be positive")
	}
	if issuer == "" {
		issuer = "go-api"
	}
	if now == nil {
		now = time.Now
	}
	return &Issuer{secret: append([]byte(nil), secret...), ttl: ttl, issuer: issuer, now: now}, nil
}

func (i *Issuer) IssueAccess(userID uuid.UUID) (string, time.Time, error) {
	now := i.now().UTC()
	expires := now.Add(i.ttl)
	id, err := uuid.NewV7()
	if err != nil {
		return "", time.Time{}, err
	}

	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		Issuer:    i.issuer,
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expires),
		ID:        id.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(i.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expires, nil
}

func (i *Issuer) ParseAccess(raw string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return i.secret, nil
	}, jwt.WithIssuer(i.issuer), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid {
		return uuid.Nil, errors.New("invalid access token")
	}
	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, errors.New("invalid access token")
	}
	return id, nil
}
