package httpapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/service"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/store"
	"github.com/jakunzler/rocketseat-go/10-complete_basic_go_api/internal/validator"
)

type createTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
}

func (req createTaskRequest) Valid(context.Context) validator.Evaluator {
	var eval validator.Evaluator
	title := strings.TrimSpace(req.Title)
	eval.CheckField(validator.NotBlank(title) && validator.MaxChars(title, 140), "title", "must be between 1 and 140 characters")
	eval.CheckField(validator.MaxChars(req.Description, 4000), "description", "must be at most 4000 characters")
	eval.CheckField(req.Priority >= 0 && req.Priority <= 100, "priority", "must be between 0 and 100")
	return eval
}

type updateTaskRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Priority    *int    `json:"priority"`
	Done        *bool   `json:"done"`
}

func (req updateTaskRequest) Valid(context.Context) validator.Evaluator {
	var eval validator.Evaluator
	if req.Title == nil && req.Description == nil && req.Priority == nil && req.Done == nil {
		eval.AddFieldError("body", "must include at least one field")
		return eval
	}
	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		eval.CheckField(validator.NotBlank(title) && validator.MaxChars(title, 140), "title", "must be between 1 and 140 characters")
	}
	if req.Description != nil {
		eval.CheckField(validator.MaxChars(*req.Description, 4000), "description", "must be at most 4000 characters")
	}
	if req.Priority != nil {
		eval.CheckField(*req.Priority >= 0 && *req.Priority <= 100, "priority", "must be between 0 and 100")
	}
	return eval
}

type taskJSON struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
	Done        bool   `json:"done"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func toTaskJSON(task store.Task) taskJSON {
	return taskJSON{
		ID:          task.ID.String(),
		UserID:      task.UserID.String(),
		Title:       task.Title,
		Description: task.Description,
		Priority:    task.Priority,
		Done:        task.Done,
		CreatedAt:   task.CreatedAt.UTC().Format(timeRFC3339),
		UpdatedAt:   task.UpdatedAt.UTC().Format(timeRFC3339),
	}
}

const timeRFC3339 = "2006-01-02T15:04:05.999999999Z07:00"

func (a *API) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFrom(r.Context())
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "unauthorized", "missing bearer token", nil)
		return
	}
	req, ok := decodeValid[createTaskRequest](w, r)
	if !ok {
		return
	}
	task, err := a.tasks.Create(r.Context(), service.CreateTaskInput{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
	})
	if err != nil {
		a.writeServiceError(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/v1/tasks/"+task.ID.String())
	writeData(w, http.StatusCreated, toTaskJSON(task))
}

func (a *API) handleListTasks(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFrom(r.Context())
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "unauthorized", "missing bearer token", nil)
		return
	}
	limit, offset, fields := pageParams(r)
	if len(fields) > 0 {
		writeError(w, r, http.StatusUnprocessableEntity, "validation_error", "request validation failed", fields)
		return
	}
	page, err := a.tasks.List(r.Context(), userID, limit, offset)
	if err != nil {
		a.writeServiceError(w, r, err)
		return
	}
	items := make([]taskJSON, 0, len(page.Tasks))
	for _, task := range page.Tasks {
		items = append(items, toTaskJSON(task))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": items,
		"meta": map[string]int{"limit": page.Limit, "offset": page.Offset, "total": page.Total},
	})
}

func (a *API) handleGetTask(w http.ResponseWriter, r *http.Request) {
	userID, taskID, ok := a.taskIDs(w, r)
	if !ok {
		return
	}
	task, err := a.tasks.Get(r.Context(), userID, taskID)
	if err != nil {
		a.writeServiceError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, toTaskJSON(task))
}

func (a *API) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	userID, taskID, ok := a.taskIDs(w, r)
	if !ok {
		return
	}
	req, ok := decodeValid[updateTaskRequest](w, r)
	if !ok {
		return
	}
	task, err := a.tasks.Update(r.Context(), service.UpdateTaskInput{
		UserID:      userID,
		TaskID:      taskID,
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		Done:        req.Done,
	})
	if err != nil {
		a.writeServiceError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, toTaskJSON(task))
}

func (a *API) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	userID, taskID, ok := a.taskIDs(w, r)
	if !ok {
		return
	}
	if err := a.tasks.Delete(r.Context(), userID, taskID); err != nil {
		a.writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) taskIDs(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	userID, ok := userIDFrom(r.Context())
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "unauthorized", "missing bearer token", nil)
		return uuid.Nil, uuid.Nil, false
	}
	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, http.StatusUnprocessableEntity, "validation_error", "request validation failed", map[string]string{
			"id": "must be a valid uuid",
		})
		return uuid.Nil, uuid.Nil, false
	}
	return userID, taskID, true
}

func pageParams(r *http.Request) (int, int, map[string]string) {
	limit := 20
	offset := 0
	fields := map[string]string{}
	if raw := r.URL.Query().Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > 100 {
			fields["limit"] = "must be between 1 and 100"
		} else {
			limit = n
		}
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 || n > 10000 {
			fields["offset"] = "must be between 0 and 10000"
		} else {
			offset = n
		}
	}
	if len(fields) == 0 {
		return limit, offset, nil
	}
	return limit, offset, fields
}
