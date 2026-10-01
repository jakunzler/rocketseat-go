package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/store"
)

type CreateTaskInput struct {
	UserID      uuid.UUID
	Title       string
	Description string
	Priority    int
}

type UpdateTaskInput struct {
	UserID      uuid.UUID
	TaskID      uuid.UUID
	Title       *string
	Description *string
	Priority    *int
	Done        *bool
}

type Page struct {
	Tasks  []store.Task
	Total  int
	Limit  int
	Offset int
}

type Tasks struct {
	repo store.TaskRepository
	now  func() time.Time
}

func NewTasks(repo store.TaskRepository, now func() time.Time) *Tasks {
	if now == nil {
		now = time.Now
	}
	return &Tasks{repo: repo, now: now}
}

func (s *Tasks) Create(ctx context.Context, in CreateTaskInput) (store.Task, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return store.Task{}, err
	}
	now := s.now().UTC()
	task, err := s.repo.Create(ctx, store.Task{
		ID:          id,
		UserID:      in.UserID,
		Title:       strings.TrimSpace(in.Title),
		Description: in.Description,
		Priority:    in.Priority,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		return store.Task{}, mapStore(err)
	}
	return task, nil
}

func (s *Tasks) List(ctx context.Context, userID uuid.UUID, limit, offset int) (Page, error) {
	tasks, total, err := s.repo.List(ctx, store.TaskFilter{UserID: userID, Limit: limit, Offset: offset})
	if err != nil {
		return Page{}, mapStore(err)
	}
	if tasks == nil {
		tasks = []store.Task{}
	}
	return Page{Tasks: tasks, Total: total, Limit: limit, Offset: offset}, nil
}

func (s *Tasks) Get(ctx context.Context, userID, taskID uuid.UUID) (store.Task, error) {
	task, err := s.repo.FindByID(ctx, userID, taskID)
	if err != nil {
		return store.Task{}, mapStore(err)
	}
	return task, nil
}

func (s *Tasks) Update(ctx context.Context, in UpdateTaskInput) (store.Task, error) {
	if in.Title == nil && in.Description == nil && in.Priority == nil && in.Done == nil {
		return store.Task{}, ErrInvalid
	}
	current, err := s.repo.FindByID(ctx, in.UserID, in.TaskID)
	if err != nil {
		return store.Task{}, mapStore(err)
	}
	if in.Title != nil {
		current.Title = strings.TrimSpace(*in.Title)
	}
	if in.Description != nil {
		current.Description = *in.Description
	}
	if in.Priority != nil {
		current.Priority = *in.Priority
	}
	if in.Done != nil {
		current.Done = *in.Done
	}
	current.UpdatedAt = s.now().UTC()
	updated, err := s.repo.Update(ctx, current)
	if err != nil {
		return store.Task{}, mapStore(err)
	}
	return updated, nil
}

func (s *Tasks) Delete(ctx context.Context, userID, taskID uuid.UUID) error {
	return mapStore(s.repo.Delete(ctx, userID, taskID))
}
