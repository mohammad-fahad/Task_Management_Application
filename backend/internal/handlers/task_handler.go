package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/taskmanager/backend/internal/middleware"
	"github.com/taskmanager/backend/internal/models"
	"github.com/taskmanager/backend/internal/repository"
	"github.com/taskmanager/backend/internal/service"
)

// TaskHandler handles task-related HTTP requests.
type TaskHandler struct {
	taskService *service.TaskService
	validate    *validator.Validate
}

// NewTaskHandler creates a new TaskHandler instance.
func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
		validate:    validator.New(),
	}
}

// CreateTask handles POST /api/tasks
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeUnauthorizedError(w, "authentication required")
		return
	}

	var input models.TaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeValidationError(w, "invalid request body", models.ErrorDetail{
			Field:   "body",
			Message: "failed to parse JSON request body",
		})
		return
	}

	input.Status = "pending" // Default status for new tasks

	if err := h.validate.Struct(&input); err != nil {
		writeValidationErrors(w, err)
		return
	}

	task, err := h.taskService.CreateTask(r.Context(), userID, &input)
	if err != nil {
		writeInternalError(w, "failed to create task")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "task created successfully",
		Data:    task,
	})
}

// GetTask handles GET /api/tasks/{id}
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeUnauthorizedError(w, "authentication required")
		return
	}

	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeValidationError(w, "invalid task ID", models.ErrorDetail{
			Field:   "id",
			Message: "task ID must be a valid UUID",
		})
		return
	}

	task, err := h.taskService.GetTask(r.Context(), taskID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			writeNotFoundError(w, "task not found")
			return
		}
		writeInternalError(w, "failed to get task")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    task,
	})
}

// ListTasks handles GET /api/tasks with pagination, filtering, and sorting.
func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeUnauthorizedError(w, "authentication required")
		return
	}

	// Parse query parameters
	filter := &models.TaskFilter{}

	// Pagination
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	} else {
		filter.Page = 1
	}

	if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
		if perPage, err := strconv.Atoi(perPageStr); err == nil && perPage > 0 && perPage <= 100 {
			filter.PerPage = perPage
		}
	} else {
		filter.PerPage = 20
	}

	// Filtering
	filter.Status = r.URL.Query().Get("status")
	filter.Priority = r.URL.Query().Get("priority")
	filter.Search = r.URL.Query().Get("search")

	// Sorting
	filter.SortBy = r.URL.Query().Get("sort_by")
	if filter.SortBy == "" {
		filter.SortBy = "created_at"
	}
	filter.SortDir = r.URL.Query().Get("sort_dir")
	if filter.SortDir == "" {
		filter.SortDir = "desc"
	}

	// Validate filter
	if err := h.validate.Struct(filter); err != nil {
		writeValidationErrors(w, err)
		return
	}

	result, err := h.taskService.ListTasks(r.Context(), userID, filter)
	if err != nil {
		writeInternalError(w, "failed to list tasks")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    result,
	})
}

// UpdateTask handles PUT /api/tasks/{id}
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeUnauthorizedError(w, "authentication required")
		return
	}

	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeValidationError(w, "invalid task ID", models.ErrorDetail{
			Field:   "id",
			Message: "task ID must be a valid UUID",
		})
		return
	}

	var input models.TaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeValidationError(w, "invalid request body", models.ErrorDetail{
			Field:   "body",
			Message: "failed to parse JSON request body",
		})
		return
	}

	if err := h.validate.Struct(&input); err != nil {
		writeValidationErrors(w, err)
		return
	}

	task, err := h.taskService.UpdateTask(r.Context(), taskID, userID, &input)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			writeNotFoundError(w, "task not found")
			return
		}
		writeInternalError(w, "failed to update task")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "task updated successfully",
		Data:    task,
	})
}

// DeleteTask handles DELETE /api/tasks/{id}
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeUnauthorizedError(w, "authentication required")
		return
	}

	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeValidationError(w, "invalid task ID", models.ErrorDetail{
			Field:   "id",
			Message: "task ID must be a valid UUID",
		})
		return
	}

	err = h.taskService.DeleteTask(r.Context(), taskID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			writeNotFoundError(w, "task not found")
			return
		}
		writeInternalError(w, "failed to delete task")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "task deleted successfully",
	})
}

// GetTaskActivity handles GET /api/tasks/{id}/activity
func (h *TaskHandler) GetTaskActivity(w http.ResponseWriter, r *http.Request) {
	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeValidationError(w, "invalid task ID", models.ErrorDetail{
			Field:   "id",
			Message: "task ID must be a valid UUID",
		})
		return
	}

	limit := 50
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	logs, total, err := h.taskService.GetTaskActivity(r.Context(), taskID, limit, offset)
	if err != nil {
		writeInternalError(w, "failed to get task activity")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"activity_logs": logs,
			"total":         total,
			"limit":         limit,
			"offset":        offset,
		},
	})
}