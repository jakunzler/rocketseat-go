package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/store"
)

type Store struct {
	mu      sync.Mutex
	users   map[uuid.UUID]store.User
	byEmail map[string]uuid.UUID
	tokens  map[uuid.UUID]store.RefreshToken
	byHash  map[string]uuid.UUID
	tasks   map[uuid.UUID]store.Task
}

func New() *Store {
	return &Store{
		users:   map[uuid.UUID]store.User{},
		byEmail: map[string]uuid.UUID{},
		tokens:  map[uuid.UUID]store.RefreshToken{},
		byHash:  map[string]uuid.UUID{},
		tasks:   map[uuid.UUID]store.Task{},
	}
}

func (s *Store) Create(ctx context.Context, user store.User) (store.User, error) {
	if err := ctx.Err(); err != nil {
		return store.User{}, err
	}
	if user.ID == uuid.Nil || user.Email == "" {
		return store.User{}, store.ErrInvalid
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.byEmail[user.Email]; exists {
		return store.User{}, store.ErrConflict
	}
	s.users[user.ID] = user
	s.byEmail[user.Email] = user.ID
	return user, nil
}

func (s *Store) FindByEmail(ctx context.Context, email string) (store.User, error) {
	if err := ctx.Err(); err != nil {
		return store.User{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.byEmail[email]
	if !ok {
		return store.User{}, store.ErrNotFound
	}
	return s.users[id], nil
}

func (s *Store) FindByID(ctx context.Context, id uuid.UUID) (store.User, error) {
	if err := ctx.Err(); err != nil {
		return store.User{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.users[id]
	if !ok {
		return store.User{}, store.ErrNotFound
	}
	return user, nil
}

func (s *Store) CreateToken(ctx context.Context, token store.RefreshToken) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.byHash[token.TokenHash]; exists {
		return store.ErrConflict
	}
	if token.CreatedAt.IsZero() {
		token.CreatedAt = time.Now().UTC()
	}
	s.tokens[token.ID] = token
	s.byHash[token.TokenHash] = token.ID
	return nil
}

func (s *Store) FindByHash(ctx context.Context, hash string) (store.RefreshToken, error) {
	if err := ctx.Err(); err != nil {
		return store.RefreshToken{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.byHash[hash]
	if !ok {
		return store.RefreshToken{}, store.ErrNotFound
	}
	return s.tokens[id], nil
}

func (s *Store) Rotate(ctx context.Context, currentID uuid.UUID, next store.RefreshToken) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	current, ok := s.tokens[currentID]
	if !ok {
		return store.ErrNotFound
	}
	if current.RevokedAt != nil {
		return store.ErrConflict
	}
	if _, exists := s.byHash[next.TokenHash]; exists {
		return store.ErrConflict
	}

	now := time.Now().UTC()
	current.RevokedAt = &now
	s.tokens[current.ID] = current
	if next.CreatedAt.IsZero() {
		next.CreatedAt = now
	}
	s.tokens[next.ID] = next
	s.byHash[next.TokenHash] = next.ID
	return nil
}

func (s *Store) RevokeByHash(ctx context.Context, hash string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.byHash[hash]
	if !ok {
		return nil
	}
	token := s.tokens[id]
	if token.RevokedAt == nil {
		now := time.Now().UTC()
		token.RevokedAt = &now
		s.tokens[id] = token
	}
	return nil
}

func (s *Store) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	for id, token := range s.tokens {
		if token.UserID == userID && token.RevokedAt == nil {
			token.RevokedAt = &now
			s.tokens[id] = token
		}
	}
	return nil
}

func (s *Store) DeleteExpired(ctx context.Context, before time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, token := range s.tokens {
		if token.ExpiresAt.Before(before) {
			delete(s.byHash, token.TokenHash)
			delete(s.tokens, id)
		}
	}
	return nil
}

func (s *Store) CreateTask(ctx context.Context, task store.Task) (store.Task, error) {
	if err := ctx.Err(); err != nil {
		return store.Task{}, err
	}
	if task.ID == uuid.Nil || task.UserID == uuid.Nil || task.Title == "" {
		return store.Task{}, store.ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[task.ID] = task
	return task, nil
}

func (s *Store) List(ctx context.Context, filter store.TaskFilter) ([]store.Task, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	items := make([]store.Task, 0)
	for _, task := range s.tasks {
		if task.UserID == filter.UserID {
			items = append(items, task)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].ID.String() > items[j].ID.String()
		}
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	total := len(items)
	if filter.Offset >= total {
		return []store.Task{}, total, nil
	}
	end := total
	if filter.Limit > 0 {
		end = filter.Offset + filter.Limit
		if end > total {
			end = total
		}
	}
	out := make([]store.Task, end-filter.Offset)
	copy(out, items[filter.Offset:end])
	return out, total, nil
}

func (s *Store) FindTask(ctx context.Context, userID, id uuid.UUID) (store.Task, error) {
	if err := ctx.Err(); err != nil {
		return store.Task{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[id]
	if !ok || task.UserID != userID {
		return store.Task{}, store.ErrNotFound
	}
	return task, nil
}

func (s *Store) UpdateTask(ctx context.Context, task store.Task) (store.Task, error) {
	if err := ctx.Err(); err != nil {
		return store.Task{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.tasks[task.ID]
	if !ok || current.UserID != task.UserID {
		return store.Task{}, store.ErrNotFound
	}
	s.tasks[task.ID] = task
	return task, nil
}

func (s *Store) DeleteTask(ctx context.Context, userID, id uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[id]
	if !ok || task.UserID != userID {
		return store.ErrNotFound
	}
	delete(s.tasks, id)
	return nil
}

type Users struct{ *Store }
type Tokens struct{ *Store }
type Tasks struct{ *Store }

func (u Users) Create(ctx context.Context, user store.User) (store.User, error) {
	return u.Store.Create(ctx, user)
}
func (u Users) FindByEmail(ctx context.Context, email string) (store.User, error) {
	return u.Store.FindByEmail(ctx, email)
}
func (u Users) FindByID(ctx context.Context, id uuid.UUID) (store.User, error) {
	return u.Store.FindByID(ctx, id)
}

func (t Tokens) Create(ctx context.Context, token store.RefreshToken) error {
	return t.Store.CreateToken(ctx, token)
}
func (t Tokens) FindByHash(ctx context.Context, hash string) (store.RefreshToken, error) {
	return t.Store.FindByHash(ctx, hash)
}
func (t Tokens) Rotate(ctx context.Context, currentID uuid.UUID, next store.RefreshToken) error {
	return t.Store.Rotate(ctx, currentID, next)
}
func (t Tokens) RevokeByHash(ctx context.Context, hash string) error {
	return t.Store.RevokeByHash(ctx, hash)
}
func (t Tokens) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	return t.Store.RevokeAllForUser(ctx, userID)
}
func (t Tokens) DeleteExpired(ctx context.Context, before time.Time) error {
	return t.Store.DeleteExpired(ctx, before)
}

func (t Tasks) Create(ctx context.Context, task store.Task) (store.Task, error) {
	return t.Store.CreateTask(ctx, task)
}
func (t Tasks) List(ctx context.Context, filter store.TaskFilter) ([]store.Task, int, error) {
	return t.Store.List(ctx, filter)
}
func (t Tasks) FindByID(ctx context.Context, userID, id uuid.UUID) (store.Task, error) {
	return t.Store.FindTask(ctx, userID, id)
}
func (t Tasks) Update(ctx context.Context, task store.Task) (store.Task, error) {
	return t.Store.UpdateTask(ctx, task)
}
func (t Tasks) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return t.Store.DeleteTask(ctx, userID, id)
}

func Repos(s *Store) (store.UserRepository, store.RefreshTokenRepository, store.TaskRepository) {
	return Users{s}, Tokens{s}, Tasks{s}
}

var (
	_ store.UserRepository         = Users{}
	_ store.RefreshTokenRepository = Tokens{}
	_ store.TaskRepository         = Tasks{}
)
