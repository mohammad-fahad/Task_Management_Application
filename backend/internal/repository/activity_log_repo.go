package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/taskmanager/backend/internal/models"
)

var (
	ErrActivityLogNotFound = errors.New("activity log not found")
)

// PostgresActivityLogRepository implements ActivityLogRepository using PostgreSQL.
type PostgresActivityLogRepository struct {
	db *sql.DB
}

// NewPostgresActivityLogRepository creates a new instance of PostgresActivityLogRepository.
func NewPostgresActivityLogRepository(db *sql.DB) *PostgresActivityLogRepository {
	return &PostgresActivityLogRepository{db: db}
}

// Create inserts a new activity log entry.
func (r *PostgresActivityLogRepository) Create(ctx context.Context, input *models.ActivityLogInput) (*models.ActivityLog, error) {
	detailsJSON, err := json.Marshal(input.Details)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal activity details: %w", err)
	}

	if input.Details == nil {
		detailsJSON = []byte("{}")
	}

	log := &models.ActivityLog{}
	err = r.db.QueryRowContext(
		ctx,
		`INSERT INTO activity_logs (task_id, user_id, action, details) 
		 VALUES ($1, $2, $3, $4) 
		 RETURNING id, task_id, user_id, action, details, created_at`,
		input.TaskID, input.UserID, input.Action, detailsJSON,
	).Scan(&log.ID, &log.TaskID, &log.UserID, &log.Action, &detailsJSON, &log.CreatedAt)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(detailsJSON, &log.Details); err != nil {
		log.Details = map[string]interface{}{}
	}

	return log, nil
}

// ListByTask retrieves all activity logs for a given task, ordered by creation time.
func (r *PostgresActivityLogRepository) ListByTask(ctx context.Context, taskID uuid.UUID, limit, offset int) ([]models.ActivityLog, int, error) {
	if limit < 1 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var total int
	err := r.db.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM activity_logs WHERE task_id = $1`,
		taskID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, task_id, user_id, action, details, created_at 
		 FROM activity_logs WHERE task_id = $1 
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		taskID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	logs := make([]models.ActivityLog, 0)
	for rows.Next() {
		var log models.ActivityLog
		var detailsJSON []byte
		err := rows.Scan(&log.ID, &log.TaskID, &log.UserID, &log.Action, &detailsJSON, &log.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		if err := json.Unmarshal(detailsJSON, &log.Details); err != nil {
			log.Details = map[string]interface{}{}
		}
		logs = append(logs, log)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// ListByUser retrieves all activity logs for a given user, ordered by creation time.
func (r *PostgresActivityLogRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.ActivityLog, int, error) {
	if limit < 1 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var total int
	err := r.db.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM activity_logs WHERE user_id = $1`,
		userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, task_id, user_id, action, details, created_at 
		 FROM activity_logs WHERE user_id = $1 
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	logs := make([]models.ActivityLog, 0)
	for rows.Next() {
		var log models.ActivityLog
		var detailsJSON []byte
		err := rows.Scan(&log.ID, &log.TaskID, &log.UserID, &log.Action, &detailsJSON, &log.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		if err := json.Unmarshal(detailsJSON, &log.Details); err != nil {
			log.Details = map[string]interface{}{}
		}
		logs = append(logs, log)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// DeleteByTask removes all activity logs associated with a specific task.
func (r *PostgresActivityLogRepository) DeleteByTask(ctx context.Context, taskID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM activity_logs WHERE task_id = $1`, taskID)
	return err
}

// Ensure interface compliance at compile time.
var _ ActivityLogRepository = (*PostgresActivityLogRepository)(nil)