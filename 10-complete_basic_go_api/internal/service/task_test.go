package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/store/memory"
)

func TestTaskOwnershipAndPagination(t *testing.T) {
	mem := memory.New()
	_, _, tasks := memory.Repos(mem)
	start := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	clock := start
	svc := NewTasks(tasks, func() time.Time { return clock })

	owner := uuid.Must(uuid.NewV7())
	other := uuid.Must(uuid.NewV7())
	ctx := context.Background()

	first, err := svc.Create(ctx, CreateTaskInput{UserID: owner, Title: "  First  ", Description: "a", Priority: 1})
	if err != nil {
		t.Fatal(err)
	}
	if first.Title != "First" {
		t.Fatalf("title = %q", first.Title)
	}
	clock = clock.Add(time.Second)
	second, err := svc.Create(ctx, CreateTaskInput{UserID: owner, Title: "Second", Priority: 2})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Create(ctx, CreateTaskInput{UserID: other, Title: "Hidden", Priority: 0}); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.Get(ctx, other, first.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross user get err = %v", err)
	}

	page, err := svc.List(ctx, owner, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 2 || len(page.Tasks) != 1 || page.Tasks[0].ID != second.ID {
		t.Fatalf("page = %+v", page)
	}

	done := true
	title := "Second updated"
	updated, err := svc.Update(ctx, UpdateTaskInput{UserID: owner, TaskID: second.ID, Title: &title, Done: &done})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Title != "Second updated" || !updated.Done {
		t.Fatalf("updated = %+v", updated)
	}
	if _, err := svc.Update(ctx, UpdateTaskInput{UserID: owner, TaskID: second.ID}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("empty update err = %v", err)
	}
	if err := svc.Delete(ctx, other, first.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross delete err = %v", err)
	}
	if err := svc.Delete(ctx, owner, first.ID); err != nil {
		t.Fatal(err)
	}
}
