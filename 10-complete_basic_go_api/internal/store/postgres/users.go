package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/store"
)

func (r *Repository) Create(ctx context.Context, user store.User) (store.User, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO users (id, name, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, email, password_hash, created_at, updated_at
	`, user.ID, user.Name, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt)
	return scanUser(row)
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (store.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`, email)
	return scanUser(row)
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (store.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id)
	return scanUser(row)
}

func scanUser(row pgx.Row) (store.User, error) {
	var user store.User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return store.User{}, mapErr(err)
	}
	user.CreatedAt = user.CreatedAt.UTC()
	user.UpdatedAt = user.UpdatedAt.UTC()
	return user, nil
}

func utc(t time.Time) time.Time {
	return t.UTC()
}
