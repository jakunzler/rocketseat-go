package service

import (
	"errors"

	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/store"
)

var (
	ErrNotFound           = errors.New("not found")
	ErrConflict           = errors.New("conflict")
	ErrInvalid            = errors.New("invalid")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
)

func mapStore(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, store.ErrNotFound):
		return ErrNotFound
	case errors.Is(err, store.ErrConflict):
		return ErrConflict
	case errors.Is(err, store.ErrInvalid):
		return ErrInvalid
	default:
		return err
	}
}
