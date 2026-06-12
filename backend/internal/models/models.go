package models

import (
	"time"

	"github.com/google/uuid"
)

// #############################################################################
// User represents a registered user in the system.
// #############################################################################
type User struct {
	ID           uuid.UUID `json:"id" validate:"required,uuid"`
	Email        string    `json:"email" validate:"required,email,max=255"`
	PasswordHash string    `json:"-" validate:"required,min=8"`
	Role         string    `json:"role" validate:"required,oneof=user admin"`
	CreatedAt    time.Time `json:"created_at" validate:"required"`
}

// UserInput is used for user creation requests (plaintext password).
type UserInput struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=128"`
	Role     string `json:"role" validate:"omitempty,oneof=user admin"`
}

// LoginInput is used for authentication requests.
type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// #############################################################################
// Task represents a task entity owned by a user.
// #############################################################################
type Task struct {
	ID          uuid.UUID  `json:"id" validate:"required,uuid"`
	UserID      uuid.UUID  `json:"user_id" validate:"required,uuid"`
	Title       string     `json:"title" validate:"required,min=1,max=255"`
	Description string     `json:"description" validate:"max=5000"`
	Status      string     `json:"status" validate:"required,oneof=pending in_progress completed cancelled"`
	Priority    string     `json:"priority" validate:"required,oneof=low medium high critical"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at" validate:"required"`
}

// TaskInput is used for task creation/update requests.
type TaskInput struct {
	Title       string     `json:"title" validate:"required,min=1,max=255"`
	Description string     `json:"description" validate:"max=5000"`
	Status      string     `json:"status" validate:"required,oneof=pending in_progress completed cancelled"`
	Priority    string     `json:"priority" validate:"required,oneof=low medium high critical"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

// TaskFilter holds pagination, filtering, and sorting parameters for listing tasks.
type TaskFilter struct {
	Page     int    `json:"page" validate:"omitempty,min=1"`
	PerPage  int    `json:"per_page" validate:"omitempty,min=1,max=100"`
	Status   string `json:"status,omitempty" validate:"omitempty,oneof=pending in_progress completed cancelled"`
	Priority string `json:"priority,omitempty" validate:"omitempty,oneof=low medium high critical"`
	Search   string `json:"search,omitempty"`
	SortBy   string `json:"sort_by,omitempty" validate:"omitempty,oneof=title status priority due_date created_at"`
	SortDir  string `json:"sort_dir,omitempty" validate:"omitempty,oneof=asc desc"`
}

// TaskListResponse wraps paginated task results.
type TaskListResponse struct {
	Data       []Task `json:"data"`
	Total      int    `json:"total"`
	Page       int    `json:"page"`
	PerPage    int    `json:"per_page"`
	TotalPages int    `json:"total_pages"`
}

// #############################################################################
// ActivityLog represents an immutable audit trail entry.
// #############################################################################
type ActivityLog struct {
	ID        uuid.UUID              `json:"id" validate:"required,uuid"`
	TaskID    uuid.UUID              `json:"task_id" validate:"required,uuid"`
	UserID    uuid.UUID              `json:"user_id" validate:"required,uuid"`
	Action    string                 `json:"action" validate:"required"`
	Details   map[string]interface{} `json:"details" validate:"required"`
	CreatedAt time.Time              `json:"created_at" validate:"required"`
}

// ActivityLogInput is used for creating audit log entries.
type ActivityLogInput struct {
	TaskID  uuid.UUID              `json:"task_id" validate:"required,uuid"`
	UserID  uuid.UUID              `json:"user_id" validate:"required,uuid"`
	Action  string                 `json:"action" validate:"required"`
	Details map[string]interface{} `json:"details"`
}

// #############################################################################
// API Standard response wrappers.
// #############################################################################

// APIResponse is the standard envelope for all JSON API responses.
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ErrorDetail provides structured validation error information.
type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrorResponse wraps multiple field-level validation errors.
type ValidationErrorResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Errors  []ErrorDetail `json:"errors"`
}