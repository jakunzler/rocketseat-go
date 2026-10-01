//go:build integration

package postgres

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/store"
)

func TestRepositoryRoundTrip(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Fatal("TEST_DATABASE_URL is required")
	}
	parsed, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(parsed.ConnConfig.Database, "test") {
		t.Fatalf("refusing to use database %q", parsed.ConnConfig.Database)
	}

	ctx := context.Background()
	pool, err := NewPool(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	if err := Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `TRUNCATE TABLE users CASCADE`); err != nil {
		t.Fatal(err)
	}

	repo := NewRepository(pool)
	users, tokens, tasks := Repos(repo)
	now := time.Now().UTC()
	userID := uuid.Must(uuid.NewV7())
	user, err := users.Create(ctx, store.User{
		ID:           userID,
		Name:         "Ada",
		Email:        "ada@example.com",
		PasswordHash: "hash",
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := users.Create(ctx, user); !errors.Is(err, store.ErrConflict) {
		t.Fatalf("duplicate user err = %v", err)
	}

	task, err := tasks.Create(ctx, store.Task{
		ID:          uuid.Must(uuid.NewV7()),
		UserID:      user.ID,
		Title:       "Ship API",
		Description: "integration",
		Priority:    3,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		t.Fatal(err)
	}
	listed, total, err := tasks.List(ctx, store.TaskFilter{UserID: user.ID, Limit: 10, Offset: 0})
	if err != nil || total != 1 || len(listed) != 1 {
		t.Fatalf("list total=%d len=%d err=%v", total, len(listed), err)
	}
	task.Done = true
	task.UpdatedAt = now.Add(time.Second)
	if _, err := tasks.Update(ctx, task); err != nil {
		t.Fatal(err)
	}
	if _, err := tasks.FindByID(ctx, uuid.Must(uuid.NewV7()), task.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("foreign find err = %v", err)
	}

	rawHash := "abc123"
	tokenID := uuid.Must(uuid.NewV7())
	if err := tokens.Create(ctx, store.RefreshToken{
		ID:        tokenID,
		UserID:    user.ID,
		TokenHash: rawHash,
		ExpiresAt: now.Add(time.Hour),
	}); err != nil {
		t.Fatal(err)
	}
	next := store.RefreshToken{
		ID:        uuid.Must(uuid.NewV7()),
		UserID:    user.ID,
		TokenHash: "rotated",
		ExpiresAt: now.Add(2 * time.Hour),
	}
	if err := tokens.Rotate(ctx, tokenID, next); err != nil {
		t.Fatal(err)
	}
	if err := tokens.Rotate(ctx, tokenID, store.RefreshToken{
		ID:        uuid.Must(uuid.NewV7()),
		UserID:    user.ID,
		TokenHash: "again",
		ExpiresAt: now.Add(time.Hour),
	}); !errors.Is(err, store.ErrConflict) {
		t.Fatalf("second rotate err = %v", err)
	}
	if err := tokens.DeleteExpired(ctx, now.Add(3*time.Hour)); err != nil {
		t.Fatal(err)
	}
}
