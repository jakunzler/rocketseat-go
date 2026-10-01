package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/store"
)

func (r *Repository) CreateTask(ctx context.Context, task store.Task) (store.Task, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO tasks (id, user_id, title, description, priority, done, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, user_id, title, description, priority, done, created_at, updated_at
	`, task.ID, task.UserID, task.Title, task.Description, task.Priority, task.Done, task.CreatedAt, task.UpdatedAt)
	return scanTask(row)
}

func (r *Repository) List(ctx context.Context, filter store.TaskFilter) ([]store.Task, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM tasks WHERE user_id = $1`, filter.UserID).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, title, description, priority, done, created_at, updated_at
		FROM tasks
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3
	`, filter.UserID, filter.Limit, filter.Offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	tasks := make([]store.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

func (r *Repository) FindTaskByID(ctx context.Context, userID, id uuid.UUID) (store.Task, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, title, description, priority, done, created_at, updated_at
		FROM tasks
		WHERE id = $1 AND user_id = $2
	`, id, userID)
	return scanTask(row)
}

func (r *Repository) UpdateTask(ctx context.Context, task store.Task) (store.Task, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE tasks
		SET title = $3, description = $4, priority = $5, done = $6, updated_at = $7
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, title, description, priority, done, created_at, updated_at
	`, task.ID, task.UserID, task.Title, task.Description, task.Priority, task.Done, task.UpdatedAt)
	return scanTask(row)
}

func (r *Repository) DeleteTask(ctx context.Context, userID, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM tasks WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return nil
}

func scanTask(row interface{ Scan(dest ...any) error }) (store.Task, error) {
	var task store.Task
	err := row.Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Priority, &task.Done, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return store.Task{}, mapErr(err)
	}
	task.CreatedAt = utc(task.CreatedAt)
	task.UpdatedAt = utc(task.UpdatedAt)
	return task, nil
}
