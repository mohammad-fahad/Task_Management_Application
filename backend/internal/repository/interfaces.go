package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/taskmanager/backend/internal/models"
)

// #############################################################################
// UserRepository defines the interface for user data access operations.
// #############################################################################
type UserRepository interface {
	// Create inserts a new user and returns the created user entity.
	Create(ctx context.Context, input *models.UserInput) (*models.User, error)

	// FindByID retrieves a user by their unique identifier.
	FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)

	// FindByEmail retrieves a user by their email address.
	FindByEmail(ctx context.Context, email string) (*models.User, error)

	// ExistsByEmail checks whether a user with the given email already exists.
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// Update updates a user's information.
	Update(ctx context.Context, user *models.User) error

	// Delete removes a user by their identifier.
	Delete(ctx context.Context, id uuid.UUID) error
}

// #############################################################################
// TaskRepository defines the interface for task data access operations.
// #############################################################################
type TaskRepository interface {
	// Create inserts a new task and returns the created task entity.
	Create(ctx context.Context, userID uuid.UUID, input *models.TaskInput) (*models.Task, error)

	// FindByID retrieves a task by its ID, scoped to a specific user.
	FindByID(ctx context.Context, taskID, userID uuid.UUID) (*models.Task, error)

	// FindByIDUnscoped retrieves a task by ID without user scoping (for admin).
	FindByIDUnscoped(ctx context.Context, taskID uuid.UUID) (*models.Task, error)

	// List retrieves tasks with pagination, filtering, and sorting support.
	List(ctx context.Context, userID uuid.UUID, filter *models.TaskFilter) ([]models.Task, int, error)

	// Update modifies an existing task and returns the updated entity.
	Update(ctx context.Context, task *models.Task, input *models.TaskInput) (*models.Task, error)

	// Delete removes a task by its identifier, scoped to a user.
	Delete(ctx context.Context, taskID, userID uuid.UUID) error
}

// #############################################################################
// ActivityLogRepository defines the interface for audit log data access.
// #############################################################################
type ActivityLogRepository interface {
	// Create inserts a new activity log entry.
	Create(ctx context.Context, input *models.ActivityLogInput) (*models.ActivityLog, error)

	// ListByTask retrieves all activity logs for a given task, ordered by creation time.
	ListByTask(ctx context.Context, taskID uuid.UUID, limit, offset int) ([]models.ActivityLog, int, error)

	// ListByUser retrieves all activity logs for a given user, ordered by creation time.
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.ActivityLog, int, error)

	// DeleteByTask removes all activity logs associated with a specific task.
	DeleteByTask(ctx context.Context, taskID uuid.UUID) error
}

// #############################################################################
// DBTransaction defines the interface for database transaction operations.
// #############################################################################
type DBTransaction interface {
	// BeginTx starts a new database transaction.
	BeginTx(ctx context.Context) (interface{}, error)

	// CommitTx commits an existing transaction.
	CommitTx(tx interface{}) error

	// RollbackTx rolls back an existing transaction.
	RollbackTx(tx interface{}) error
}