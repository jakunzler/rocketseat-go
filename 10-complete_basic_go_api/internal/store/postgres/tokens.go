package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/store"
)

func (r *Repository) CreateToken(ctx context.Context, token store.RefreshToken) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, COALESCE($5, now()))
	`, token.ID, token.UserID, token.TokenHash, token.ExpiresAt, nullTime(token.CreatedAt))
	return mapErr(err)
}

func (r *Repository) FindByHash(ctx context.Context, hash string) (store.RefreshToken, error) {
	var token store.RefreshToken
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`, hash).Scan(&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt, &token.RevokedAt, &token.CreatedAt)
	if err != nil {
		return store.RefreshToken{}, mapErr(err)
	}
	token.ExpiresAt = token.ExpiresAt.UTC()
	token.CreatedAt = token.CreatedAt.UTC()
	if token.RevokedAt != nil {
		revoked := token.RevokedAt.UTC()
		token.RevokedAt = &revoked
	}
	return token, nil
}

func (r *Repository) Rotate(ctx context.Context, currentID uuid.UUID, next store.RefreshToken) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = now()
		WHERE id = $1 AND revoked_at IS NULL
	`, currentID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return store.ErrConflict
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)
	`, next.ID, next.UserID, next.TokenHash, next.ExpiresAt)
	if err != nil {
		return mapErr(err)
	}
	return tx.Commit(ctx)
}

func (r *Repository) RevokeByHash(ctx context.Context, hash string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = now()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`, hash)
	return err
}

func (r *Repository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = now()
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID)
	return err
}

func (r *Repository) DeleteExpired(ctx context.Context, before time.Time) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE expires_at < $1`, before)
	return err
}

func nullTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}
