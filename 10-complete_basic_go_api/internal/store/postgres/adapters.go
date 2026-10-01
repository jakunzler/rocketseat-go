package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/store"
)

type Users struct{ repo *Repository }
type Tokens struct{ repo *Repository }
type Tasks struct{ repo *Repository }

func Repos(repo *Repository) (store.UserRepository, store.RefreshTokenRepository, store.TaskRepository) {
	return Users{repo}, Tokens{repo}, Tasks{repo}
}

func (u Users) Create(ctx context.Context, user store.User) (store.User, error) {
	return u.repo.Create(ctx, user)
}
func (u Users) FindByEmail(ctx context.Context, email string) (store.User, error) {
	return u.repo.FindByEmail(ctx, email)
}
func (u Users) FindByID(ctx context.Context, id uuid.UUID) (store.User, error) {
	return u.repo.FindByID(ctx, id)
}

func (t Tokens) Create(ctx context.Context, token store.RefreshToken) error {
	return t.repo.CreateToken(ctx, token)
}
func (t Tokens) FindByHash(ctx context.Context, hash string) (store.RefreshToken, error) {
	return t.repo.FindByHash(ctx, hash)
}
func (t Tokens) Rotate(ctx context.Context, currentID uuid.UUID, next store.RefreshToken) error {
	return t.repo.Rotate(ctx, currentID, next)
}
func (t Tokens) RevokeByHash(ctx context.Context, hash string) error {
	return t.repo.RevokeByHash(ctx, hash)
}
func (t Tokens) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	return t.repo.RevokeAllForUser(ctx, userID)
}
func (t Tokens) DeleteExpired(ctx context.Context, before time.Time) error {
	return t.repo.DeleteExpired(ctx, before)
}

func (t Tasks) Create(ctx context.Context, task store.Task) (store.Task, error) {
	return t.repo.CreateTask(ctx, task)
}
func (t Tasks) List(ctx context.Context, filter store.TaskFilter) ([]store.Task, int, error) {
	return t.repo.List(ctx, filter)
}
func (t Tasks) FindByID(ctx context.Context, userID, id uuid.UUID) (store.Task, error) {
	return t.repo.FindTaskByID(ctx, userID, id)
}
func (t Tasks) Update(ctx context.Context, task store.Task) (store.Task, error) {
	return t.repo.UpdateTask(ctx, task)
}
func (t Tasks) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return t.repo.DeleteTask(ctx, userID, id)
}

var (
	_ store.UserRepository         = Users{}
	_ store.RefreshTokenRepository = Tokens{}
	_ store.TaskRepository         = Tasks{}
)
