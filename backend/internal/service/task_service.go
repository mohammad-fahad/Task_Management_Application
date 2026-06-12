package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/taskmanager/backend/internal/models"
	"github.com/taskmanager/backend/internal/repository"
)

// TaskService handles task business logic with activity log integration.
type TaskService struct {
	taskRepo     repository.TaskRepository
	activityRepo repository.ActivityLogRepository
}

// NewTaskService creates a new TaskService instance.
func NewTaskService(
	taskRepo repository.TaskRepository,
	activityRepo repository.ActivityLogRepository,
) *TaskService {
	return &TaskService{
		taskRepo:     taskRepo,
		activityRepo: activityRepo,
	}
}

// CreateTask creates a new task and logs the creation activity.
func (s *TaskService) CreateTask(ctx context.Context, userID uuid.UUID, input *models.TaskInput) (*models.Task, error) {
	task, err := s.taskRepo.Create(ctx, userID, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Log activity
	_, logErr := s.activityRepo.Create(ctx, &models.ActivityLogInput{
		TaskID: task.ID,
		UserID: userID,
		Action: "task_created",
		Details: map[string]interface{}{
			"title":    task.Title,
			"status":   task.Status,
			"priority": task.Priority,
		},
	})
	if logErr != nil {
		// Non-fatal: log but don't fail the task creation
		return task, fmt.Errorf("task created but failed to log activity: %w", logErr)
	}

	return task, nil
}

// GetTask retrieves a task by ID, scoped to a user.
func (s *TaskService) GetTask(ctx context.Context, taskID, userID uuid.UUID) (*models.Task, error) {
	task, err := s.taskRepo.FindByID(ctx, taskID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return task, nil
}

// GetTaskUnscoped retrieves a task by ID without user scoping (admin only).
func (s *TaskService) GetTaskUnscoped(ctx context.Context, taskID uuid.UUID) (*models.Task, error) {
	task, err := s.taskRepo.FindByIDUnscoped(ctx, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return task, nil
}

// ListTasks retrieves tasks with pagination, filtering, and sorting.
func (s *TaskService) ListTasks(ctx context.Context, userID uuid.UUID, filter *models.TaskFilter) (*models.TaskListResponse, error) {
	tasks, total, err := s.taskRepo.List(ctx, userID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	if filter == nil {
		filter = &models.TaskFilter{Page: 1, PerPage: 20}
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 {
		filter.PerPage = 20
	}

	totalPages := (total + filter.PerPage - 1) / filter.PerPage
	if totalPages < 1 {
		totalPages = 1
	}

	return &models.TaskListResponse{
		Data:       tasks,
		Total:      total,
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalPages: totalPages,
	}, nil
}

// UpdateTask updates a task and logs the changes in activity log.
func (s *TaskService) UpdateTask(ctx context.Context, taskID, userID uuid.UUID, input *models.TaskInput) (*models.Task, error) {
	// Fetch current task to detect changes
	currentTask, err := s.taskRepo.FindByID(ctx, taskID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to find task for update: %w", err)
	}

	// Perform update
	updatedTask, err := s.taskRepo.Update(ctx, currentTask, input)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	// Build activity details from detected changes
	details := s.buildUpdateDetails(currentTask, updatedTask)

	if len(details) > 0 {
		action := "task_updated"
		if currentTask.Status != updatedTask.Status {
			action = "status_changed"
		}
		if currentTask.Priority != updatedTask.Priority {
			action = "priority_changed"
		}
		if hasDateChanged(currentTask.DueDate, updatedTask.DueDate) {
			action = "due_date_changed"
		}

		_, logErr := s.activityRepo.Create(ctx, &models.ActivityLogInput{
			TaskID:  taskID,
			UserID:  userID,
			Action:  action,
			Details: details,
		})
		if logErr != nil {
			return updatedTask, fmt.Errorf("task updated but failed to log activity: %w", logErr)
		}
	}

	return updatedTask, nil
}

// DeleteTask deletes a task and logs the deletion activity.
func (s *TaskService) DeleteTask(ctx context.Context, taskID, userID uuid.UUID) error {
	// Fetch task first to log what was deleted
	task, err := s.taskRepo.FindByID(ctx, taskID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			return err
		}
		return fmt.Errorf("failed to find task for deletion: %w", err)
	}

	// Log deletion activity before deleting
	_, logErr := s.activityRepo.Create(ctx, &models.ActivityLogInput{
		TaskID: taskID,
		UserID: userID,
		Action: "task_deleted",
		Details: map[string]interface{}{
			"title": task.Title,
		},
	})
	if logErr != nil {
		// Non-fatal: continue with deletion
	}

	// Delete the task
	err = s.taskRepo.Delete(ctx, taskID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			return err
		}
		return fmt.Errorf("failed to delete task: %w", err)
	}

	return nil
}

// GetTaskActivity retrieves the activity log for a specific task.
func (s *TaskService) GetTaskActivity(ctx context.Context, taskID uuid.UUID, limit, offset int) ([]models.ActivityLog, int, error) {
	if limit < 1 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	logs, total, err := s.activityRepo.ListByTask(ctx, taskID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get task activity: %w", err)
	}

	return logs, total, nil
}

// buildUpdateDetails compares old and new task states and builds change details.
func (s *TaskService) buildUpdateDetails(oldTask, newTask *models.Task) map[string]interface{} {
	details := make(map[string]interface{})

	if oldTask.Title != newTask.Title {
		details["title"] = map[string]string{
			"from": oldTask.Title,
			"to":   newTask.Title,
		}
	}

	if oldTask.Description != newTask.Description {
		details["description"] = map[string]string{
			"from": oldTask.Description,
			"to":   newTask.Description,
		}
	}

	if oldTask.Status != newTask.Status {
		details["status"] = map[string]string{
			"from": oldTask.Status,
			"to":   newTask.Status,
		}
	}

	if oldTask.Priority != newTask.Priority {
		details["priority"] = map[string]string{
			"from": oldTask.Priority,
			"to":   newTask.Priority,
		}
	}

	if hasDateChanged(oldTask.DueDate, newTask.DueDate) {
		details["due_date"] = map[string]interface{}{
			"from": formatTimePtr(oldTask.DueDate),
			"to":   formatTimePtr(newTask.DueDate),
		}
	}

	return details
}

// hasDateChanged checks if two nullable time pointers differ.
func hasDateChanged(old, new *time.Time) bool {
	if old == nil && new == nil {
		return false
	}
	if old == nil || new == nil {
		return true
	}
	return !old.Equal(*new)
}

// formatTimePtr formats a nullable time pointer to a string for activity logs.
func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}