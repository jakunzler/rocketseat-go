package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
	ErrInvalid  = errors.New("invalid")
)

type User struct {
	ID           uuid.UUID
	Name         string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type Task struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Title       string
	Description string
	Priority    int
	Done        bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TaskFilter struct {
	UserID uuid.UUID
	Limit  int
	Offset int
}

type UserRepository interface {
	Create(ctx context.Context, user User) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByID(ctx context.Context, id uuid.UUID) (User, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token RefreshToken) error
	FindByHash(ctx context.Context, hash string) (RefreshToken, error)
	Rotate(ctx context.Context, currentID uuid.UUID, next RefreshToken) error
	RevokeByHash(ctx context.Context, hash string) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context, before time.Time) error
}

type TaskRepository interface {
	Create(ctx context.Context, task Task) (Task, error)
	List(ctx context.Context, filter TaskFilter) ([]Task, int, error)
	FindByID(ctx context.Context, userID, id uuid.UUID) (Task, error)
	Update(ctx context.Context, task Task) (Task, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
}
