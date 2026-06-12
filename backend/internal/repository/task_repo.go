package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/taskmanager/backend/internal/models"
)

var (
	ErrTaskNotFound = errors.New("task not found")
)

// PostgresTaskRepository implements TaskRepository using PostgreSQL.
type PostgresTaskRepository struct {
	db *sql.DB
}

// NewPostgresTaskRepository creates a new instance of PostgresTaskRepository.
func NewPostgresTaskRepository(db *sql.DB) *PostgresTaskRepository {
	return &PostgresTaskRepository{db: db}
}

// Create inserts a new task and returns the created task entity.
func (r *PostgresTaskRepository) Create(ctx context.Context, userID uuid.UUID, input *models.TaskInput) (*models.Task, error) {
	task := &models.Task{}
	err := r.db.QueryRowContext(
		ctx,
		`INSERT INTO tasks (user_id, title, description, status, priority, due_date) 
		 VALUES ($1, $2, $3, $4, $5, $6) 
		 RETURNING id, user_id, title, description, status, priority, due_date, created_at`,
		userID, input.Title, input.Description, input.Status, input.Priority, PtrToNullTime(input.DueDate),
	).Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status, &task.Priority, &task.DueDate, &task.CreatedAt)

	if err != nil {
		return nil, err
	}

	// Convert NullTime for due_date scan (already handled by driver with *time.Time)
	return task, nil
}

// FindByID retrieves a task by its ID, scoped to a specific user.
func (r *PostgresTaskRepository) FindByID(ctx context.Context, taskID, userID uuid.UUID) (*models.Task, error) {
	task := &models.Task{}
	var dueDate sql.NullTime

	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, user_id, title, description, status, priority, due_date, created_at 
		 FROM tasks WHERE id = $1 AND user_id = $2`,
		taskID, userID,
	).Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status, &task.Priority, &dueDate, &task.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	task.DueDate = NullTimeToPtr(dueDate)
	return task, nil
}

// FindByIDUnscoped retrieves a task by ID without user scoping (for admin).
func (r *PostgresTaskRepository) FindByIDUnscoped(ctx context.Context, taskID uuid.UUID) (*models.Task, error) {
	task := &models.Task{}
	var dueDate sql.NullTime

	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, user_id, title, description, status, priority, due_date, created_at 
		 FROM tasks WHERE id = $1`,
		taskID,
	).Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status, &task.Priority, &dueDate, &task.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	task.DueDate = NullTimeToPtr(dueDate)
	return task, nil
}

// List retrieves tasks with pagination, filtering, and sorting support.
func (r *PostgresTaskRepository) List(ctx context.Context, userID uuid.UUID, filter *models.TaskFilter) ([]models.Task, int, error) {
	if filter == nil {
		filter = &models.TaskFilter{Page: 1, PerPage: 20}
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 || filter.PerPage > 100 {
		filter.PerPage = 20
	}

	// Build dynamic query
	var whereClauses []string
	var args []interface{}
	argIdx := 1

	whereClauses = append(whereClauses, fmt.Sprintf("user_id = $%d", argIdx))
	args = append(args, userID)
	argIdx++

	if filter.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}

	if filter.Priority != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("priority = $%d", argIdx))
		args = append(args, filter.Priority)
		argIdx++
	}

	if filter.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx))
		searchPattern := "%" + filter.Search + "%"
		args = append(args, searchPattern)
		argIdx++
	}

	whereStr := strings.Join(whereClauses, " AND ")

	// Count query
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tasks WHERE %s", whereStr)
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Sorting
	allowedSortFields := map[string]bool{
		"title": true, "status": true, "priority": true,
		"due_date": true, "created_at": true,
	}

	sortBy := "created_at"
	if filter.SortBy != "" && allowedSortFields[filter.SortBy] {
		sortBy = filter.SortBy
	}

	sortDir := "DESC"
	if filter.SortDir == "asc" {
		sortDir = "ASC"
	}

	// Pagination
	offset := (filter.Page - 1) * filter.PerPage
	args = append(args, filter.PerPage, offset)

	dataQuery := fmt.Sprintf(
		`SELECT id, user_id, title, description, status, priority, due_date, created_at 
		 FROM tasks WHERE %s ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		whereStr, sortBy, sortDir, argIdx, argIdx+1,
	)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	tasks := make([]models.Task, 0)
	for rows.Next() {
		var task models.Task
		var dueDate sql.NullTime
		err := rows.Scan(&task.ID, &task.UserID, &task.Title, &task.Description,
			&task.Status, &task.Priority, &dueDate, &task.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		task.DueDate = NullTimeToPtr(dueDate)
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// Update modifies an existing task and returns the updated entity.
func (r *PostgresTaskRepository) Update(ctx context.Context, task *models.Task, input *models.TaskInput) (*models.Task, error) {
	var dueDate sql.NullTime
	err := r.db.QueryRowContext(
		ctx,
		`UPDATE tasks SET title = $1, description = $2, status = $3, priority = $4, due_date = $5
		 WHERE id = $6 AND user_id = $7
		 RETURNING id, user_id, title, description, status, priority, due_date, created_at`,
		input.Title, input.Description, input.Status, input.Priority,
		PtrToNullTime(input.DueDate), task.ID, task.UserID,
	).Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status, &task.Priority, &dueDate, &task.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	task.DueDate = NullTimeToPtr(dueDate)
	return task, nil
}

// Delete removes a task by its identifier, scoped to a user.
func (r *PostgresTaskRepository) Delete(ctx context.Context, taskID, userID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = $1 AND user_id = $2`, taskID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrTaskNotFound
	}

	return nil
}

// Ensure interface compliance at compile time.
var _ TaskRepository = (*PostgresTaskRepository)(nil)