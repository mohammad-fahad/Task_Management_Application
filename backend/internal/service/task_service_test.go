package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/taskmanager/backend/internal/models"
	"github.com/taskmanager/backend/internal/repository"
)

// =============================================================================
// Mocks
// =============================================================================

// mockTaskRepository implements TaskRepository for testing.
type mockTaskRepository struct {
	tasks  map[uuid.UUID]*models.Task
	createFunc func(ctx context.Context, userID uuid.UUID, input *models.TaskInput) (*models.Task, error)
	findByIDFunc func(ctx context.Context, taskID, userID uuid.UUID) (*models.Task, error)
	listFunc    func(ctx context.Context, userID uuid.UUID, filter *models.TaskFilter) ([]models.Task, int, error)
	updateFunc  func(ctx context.Context, task *models.Task, input *models.TaskInput) (*models.Task, error)
	deleteFunc  func(ctx context.Context, taskID, userID uuid.UUID) error
}

var _ repository.TaskRepository = (*mockTaskRepository)(nil)

func (m *mockTaskRepository) Create(ctx context.Context, userID uuid.UUID, input *models.TaskInput) (*models.Task, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, userID, input)
	}
	now := time.Now()
	task := &models.Task{
		ID:          uuid.New(),
		UserID:      userID,
		Title:       input.Title,
		Description: input.Description,
		Status:      input.Status,
		Priority:    input.Priority,
		DueDate:     input.DueDate,
		CreatedAt:   now,
	}
	if m.tasks != nil {
		m.tasks[task.ID] = task
	}
	return task, nil
}

func (m *mockTaskRepository) FindByID(ctx context.Context, taskID, userID uuid.UUID) (*models.Task, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, taskID, userID)
	}
	if m.tasks != nil {
		if task, exists := m.tasks[taskID]; exists && task.UserID == userID {
			return task, nil
		}
	}
	return nil, repository.ErrTaskNotFound
}

func (m *mockTaskRepository) FindByIDUnscoped(ctx context.Context, taskID uuid.UUID) (*models.Task, error) {
	if m.tasks != nil {
		if task, exists := m.tasks[taskID]; exists {
			return task, nil
		}
	}
	return nil, repository.ErrTaskNotFound
}

func (m *mockTaskRepository) List(ctx context.Context, userID uuid.UUID, filter *models.TaskFilter) ([]models.Task, int, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, userID, filter)
	}
	return []models.Task{}, 0, nil
}

func (m *mockTaskRepository) Update(ctx context.Context, task *models.Task, input *models.TaskInput) (*models.Task, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, task, input)
	}
	task.Title = input.Title
	task.Description = input.Description
	task.Status = input.Status
	task.Priority = input.Priority
	task.DueDate = input.DueDate
	return task, nil
}

func (m *mockTaskRepository) Delete(ctx context.Context, taskID, userID uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, taskID, userID)
	}
	if m.tasks != nil {
		if _, exists := m.tasks[taskID]; exists {
			delete(m.tasks, taskID)
			return nil
		}
	}
	return repository.ErrTaskNotFound
}

// mockActivityLogRepository implements ActivityLogRepository for testing.
type mockActivityLogRepository struct {
	logs          []models.ActivityLog
	createFunc    func(ctx context.Context, input *models.ActivityLogInput) (*models.ActivityLog, error)
	listByTaskFunc func(ctx context.Context, taskID uuid.UUID, limit, offset int) ([]models.ActivityLog, int, error)
}

var _ repository.ActivityLogRepository = (*mockActivityLogRepository)(nil)

func (m *mockActivityLogRepository) Create(ctx context.Context, input *models.ActivityLogInput) (*models.ActivityLog, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, input)
	}
	log := &models.ActivityLog{
		ID:        uuid.New(),
		TaskID:    input.TaskID,
		UserID:    input.UserID,
		Action:    input.Action,
		Details:   input.Details,
		CreatedAt: time.Now(),
	}
	m.logs = append(m.logs, *log)
	return log, nil
}

func (m *mockActivityLogRepository) ListByTask(ctx context.Context, taskID uuid.UUID, limit, offset int) ([]models.ActivityLog, int, error) {
	if m.listByTaskFunc != nil {
		return m.listByTaskFunc(ctx, taskID, limit, offset)
	}
	return []models.ActivityLog{}, 0, nil
}

func (m *mockActivityLogRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.ActivityLog, int, error) {
	return []models.ActivityLog{}, 0, nil
}

func (m *mockActivityLogRepository) DeleteByTask(ctx context.Context, taskID uuid.UUID) error {
	return nil
}

// =============================================================================
// Test Fixtures
// =============================================================================

func validTaskInput() *models.TaskInput {
	dueDate := time.Now().Add(7 * 24 * time.Hour)
	return &models.TaskInput{
		Title:       "Test Task",
		Description: "This is a test task description",
		Status:      "pending",
		Priority:    "medium",
		DueDate:     &dueDate,
	}
}

func existingTask(userID uuid.UUID) *models.Task {
	dueDate := time.Now().Add(7 * 24 * time.Hour)
	return &models.Task{
		ID:          uuid.New(),
		UserID:      userID,
		Title:       "Existing Task",
		Description: "Original description",
		Status:      "pending",
		Priority:    "low",
		DueDate:     &dueDate,
		CreatedAt:   time.Now().Add(-1 * time.Hour),
	}
}

// =============================================================================
// Tests for CreateTask
// =============================================================================

func TestCreateTask_Success(t *testing.T) {
	// Arrange
	userID := uuid.New()
	mockTaskRepo := &mockTaskRepository{tasks: make(map[uuid.UUID]*models.Task)}
	mockActivityRepo := &mockActivityLogRepository{}
	svc := NewTaskService(mockTaskRepo, mockActivityRepo)

	input := validTaskInput()

	// Act
	task, err := svc.CreateTask(context.Background(), userID, input)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if task == nil {
		t.Fatal("expected task to be non-nil")
	}
	if task.Title != input.Title {
		t.Errorf("expected title %q, got %q", input.Title, task.Title)
	}
	if task.Status != "pending" {
		t.Errorf("expected status 'pending', got %q", task.Status)
	}
	if task.Priority != "medium" {
		t.Errorf("expected priority 'medium', got %q", task.Priority)
	}
	if task.UserID != userID {
		t.Errorf("expected user ID %v, got %v", userID, task.UserID)
	}
	if task.ID == uuid.Nil {
		t.Error("expected task ID to be set")
	}
	if task.CreatedAt.IsZero() {
		t.Error("expected created_at to be set")
	}

	// Verify activity log was created
	if len(mockActivityRepo.logs) != 1 {
		t.Fatalf("expected 1 activity log, got %d", len(mockActivityRepo.logs))
	}
	if mockActivityRepo.logs[0].Action != "task_created" {
		t.Errorf("expected action 'task_created', got %q", mockActivityRepo.logs[0].Action)
	}
}

func TestCreateTask_RepositoryError(t *testing.T) {
	// Arrange
	userID := uuid.New()
	expectedErr := errors.New("database connection lost")
	mockTaskRepo := &mockTaskRepository{
		createFunc: func(ctx context.Context, userID uuid.UUID, input *models.TaskInput) (*models.Task, error) {
			return nil, expectedErr
		},
	}
	mockActivityRepo := &mockActivityLogRepository{}
	svc := NewTaskService(mockTaskRepo, mockActivityRepo)

	input := validTaskInput()

	// Act
	task, err := svc.CreateTask(context.Background(), userID, input)

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if task != nil {
		t.Fatal("expected nil task on error")
	}
}

// =============================================================================
// Tests for GetTask
// =============================================================================

func TestGetTask_Success(t *testing.T) {
	// Arrange
	userID := uuid.New()
	task := existingTask(userID)
	mockTaskRepo := &mockTaskRepository{
		tasks: map[uuid.UUID]*models.Task{task.ID: task},
	}
	mockActivityRepo := &mockActivityLogRepository{}
	svc := NewTaskService(mockTaskRepo, mockActivityRepo)

	// Act
	result, err := svc.GetTask(context.Background(), task.ID, userID)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ID != task.ID {
		t.Errorf("expected task ID %v, got %v", task.ID, result.ID)
	}
}

func TestGetTask_NotFound(t *testing.T) {
	// Arrange
	userID := uuid.New()
	otherUserID := uuid.New()
	mockTaskRepo := &mockTaskRepository{tasks: make(map[uuid.UUID]*models.Task)}
	mockActivityRepo := &mockActivityLogRepository{}
	svc := NewTaskService(mockTaskRepo, mockActivityRepo)

	nonExistentID := uuid.New()

	// Act
	result, err := svc.GetTask(context.Background(), nonExistentID, userID)

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repository.ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result on not found")
	}

	// Should not see other user's task
	otherTask := existingTask(otherUserID)
	mockTaskRepo.tasks[otherTask.ID] = otherTask
	result, err = svc.GetTask(context.Background(), otherTask.ID, userID)
	if !errors.Is(err, repository.ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound for another user's task, got %v", err)
	}
}

// =============================================================================
// Tests for ListTasks
// =============================================================================

func TestListTasks_Empty(t *testing.T) {
	// Arrange
	userID := uuid.New()
	mockTaskRepo := &mockTaskRepository{
		tasks: make(map[uuid.UUID]*models.Task),
	}
	mockActivityRepo := &mockActivityLogRepository{}
	svc := NewTaskService(mockTaskRepo, mockActivityRepo)

	// Act
	result, err := svc.ListTasks(context.Background(), userID, &models.TaskFilter{
		Page:    1,
		PerPage: 20,
	})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Data) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(result.Data))
	}
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
	if result.Page != 1 {
		t.Errorf("expected page 1, got %d", result.Page)
	}
	if result.PerPage != 20 {
		t.Errorf("expected per_page 20, got %d", result.PerPage)
	}
	if result.TotalPages != 1 {
		t.Errorf("expected total_pages 1, got %d", result.TotalPages)
	}
}

func TestListTasks_WithPagination(t *testing.T) {
	// Arrange
	userID := uuid.New()
	mockTaskRepo := &mockTaskRepository{
		listFunc: func(ctx context.Context, userID uuid.UUID, filter *models.TaskFilter) ([]models.Task, int, error) {
			// Simulate 50 tasks, page 2 of 10 per page
			allTasks := make([]models.Task, 0)
			for i := 0; i < 50; i++ {
				allTasks = append(allTasks, models.Task{
					ID:     uuid.New(),
					UserID: userID,
					Title:  "Task " + string(rune('A'+i)),
				})
			}
			start := (filter.Page - 1) * filter.PerPage
			if start > 50 {
				start = 50
			}
			end := start + filter.PerPage
			if end > 50 {
				end = 50
			}
			return allTasks[start:end], 50, nil
		},
	}
	mockActivityRepo := &mockActivityLogRepository{}
	svc := NewTaskService(mockTaskRepo, mockActivityRepo)

	// Act
	result, err := svc.ListTasks(context.Background(), userID, &models.TaskFilter{
		Page:    2,
		PerPage: 10,
	})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Total != 50 {
		t.Errorf("expected total 50, got %d", result.Total)
	}
	if result.Page != 2 {
		t.Errorf("expected page 2, got %d", result.Page)
	}
	if result.PerPage != 10 {
		t.Errorf("expected per_page 10, got %d", result.PerPage)
	}
	if result.TotalPages != 5 {
		t.Errorf("expected total_pages 5, got %d", result.TotalPages)
	}
}

func TestListTasks_DefaultPagination(t *testing.T) {
	// Arrange
	userID := uuid.New()
	mockTaskRepo := &mockTaskRepository{
		listFunc: func(ctx context.Context, userID uuid.UUID, filter *models.TaskFilter) ([]models.Task, int, error) {
			return []models.Task{}, 0, nil
		},
	}
	mockActivityRepo := &mockActivityLogRepository{}
	svc := NewTaskService(mockTaskRepo, mockActivityRepo)

	// Act (nil filter should use defaults)
	result, err := svc.ListTasks(context.Background(), userID, nil)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Page != 1 {
		t.Errorf("expected default page 1, got %d", result.Page)
	}
	if result.PerPage != 20 {
		t.Errorf("expected default per_page 20, got %d", result.PerPage)
	}
}

// =============================================================================
// Tests for UpdateTask
// =============================================================================

func TestUpdateTask_Success(t *testing.T) {
	// Arrange
	userID := uuid.New()
	task := existingTask(userID)
	mockTaskRepo := &mockTaskRepository{
		tasks: map[uuid.UUID]*models.Task{task.ID: task},
	}
	mockActivityRepo := &mockActivityLogRepository{}
	svc := NewTaskService(mockTaskRepo, mockActivityRepo)

	newDueDate := time.Now().Add(14 * 24 * time.Hour)
	input := &models.TaskInput{
		Title:       "Updated Title",
		Description: "Updated description",
		Status:      "in_progress",
		Priority:    "high",
		DueDate:     &newDueDate,
	}

	// Act
	updated, err := svc.UpdateTask(context.Background(), task.ID, userID, input)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Title != input.Title {
		t.Errorf("expected title %q, got %q", input.Title, updated.Title)
	}
	if updated.Description != input.Description {
		t.Errorf("expected description %q, got %q", input.Description, updated.Description)
	}
	if updated.Status != "in_progress" {
		t.Errorf("expected status 'in_progress', got %q", updated.Status)
	}
	if updated.Priority != "high" {
		t.Errorf("expected priority 'high', got %q", updated.Priority)
	}

	// Verify activity log was created with change tracking
	if len(mockActivityRepo.logs) < 1 {
		t.Fatal("expected at least 1 activity log")
	}
	log := mockActivityRepo.logs[0]
	if log.Action != "status_changed" {
		t.Errorf("expected action 'status_changed', got %q", log.Action)
	}
	if log.TaskID != task.ID {
		t.Errorf("expected task ID %v, got %v", task.ID, log.TaskID)
	}
	if log.UserID != userID {
		t.Errorf("expected user ID %v, got %v", userID, log.UserID)
	}
}

func TestUpdateTask_PartialUpdate(t *testing.T) {
	// Arrange
	userID := uuid.New()
	task := existingTask(userID)
	mockTaskRepo := &mockTaskRepository{
		tasks: map[uuid.UUID]*models.Task{task.ID: task},
	}
	mockActivityRepo := &mockActivityLogRepository{}
	svc := NewTaskService(mockTaskRepo, mockActivityRepo)

	// Only update title and status (keep same priority)
	input := &models.TaskInput{
		Title:       "New Title Only",
		Description: task.Description,
		Status:      "completed",
		Priority:    task.Priority,
		DueDate:     task.DueDate,
	}

	// Act
	updated, err := svc.UpdateTask(context.Background(), task.ID, userID, input)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Title != "New Title Only" {
		t.Errorf("expected title 'New Title Only', got %q", updated.Title)
	}
	if updated.Status != "completed" {
		t.Errorf("expected status 'completed', got %q", updated.Status)
	}

	// Should still log the changes
	if len(mockActivityRepo.logs) < 1 {
		t.Fatal("expected at least 1 activity log for changes")
	}
}

func TestUpdateTask_NoChanges(t *testing.T) {
	// Arrange
	userID := uuid.New()
	task := existingTask(userID)
	mockTaskRepo := &mockTaskRepository{
		tasks: map[uuid.UUID]*models.Task{task.ID: task},
	}
	mockActivityRepo := &mockActivityLogRepository{}
	svc := NewTaskService(mockTaskRepo, mockActivityRepo)

	// Update with the same values
	input := &models.TaskInput{
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Priority:    task.Priority,
		DueDate:     task.DueDate,
	}

	// Act
	_, err := svc.UpdateTask(context.Background(), task.ID, userID, input)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// No activity log should be created if nothing changed
	if len(mockActivityRepo.logs) > 0 {
		t.Errorf("expected 0 activity logs for no changes, got %d", len(mockActivityRepo.logs))
	}
}

func TestUpdateTask_NotFound(t *testing.T) {
	// Arrange
	userID := uuid.New()
	mockTaskRepo := &mockTaskRepository{}
	mockActivityRepo := &mockActivityLogRepository{}
	svc := NewTaskService(mockTaskRepo, mockActivityRepo)

	// Act
	_, err := svc.UpdateTask(context.Background(), uuid.New(), userID, validTaskInput())

	// Assert
	if err == nil {
		t.Fatal("expected error for non-existent task")
	}
	if !errors.Is(err, repository.ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

// =============================================================================
// Tests for DeleteTask
// =============================================================================

func TestDeleteTask_Success(t *testing.T) {
	// Arrange
	userID := uuid.New()
	task := existingTask(userID)
	mockTaskRepo := &mockTaskRepository{
		tasks: map[uuid.UUID]*models.Task{task.ID: task},
	}
	mockActivityRepo := &mockActivityLogRepository{}
	svc := NewTaskService(mockTaskRepo, mockActivityRepo)

	// Act
	err := svc.DeleteTask(context.Background(), task.ID, userID)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify task was deleted
	_, findErr := mockTaskRepo.FindByID(context.Background(), task.ID, userID)
	if !errors.Is(findErr, repository.ErrTaskNotFound) {
		t.Error("expected task to be deleted")
	}

	// Verify activity log was created before deletion
	if len(mockActivityRepo.logs) != 1 {
		t.Fatalf("expected 1 activity log, got %d", len(mockActivityRepo.logs))
	}
	if mockActivityRepo.logs[0].Action != "task_deleted" {
		t.Errorf("expected action 'task_deleted', got %q", mockActivityRepo.logs[0].Action)
	}
	if mockActivityRepo.logs[0].Details["title"] != task.Title {
		t.Errorf("expected details title %q, got %v", task.Title, mockActivityRepo.logs[0].Details["title"])
	}
}

func TestDeleteTask_NotFound(t *testing.T) {
	// Arrange
	userID := uuid.New()
	mockTaskRepo := &mockTaskRepository{}
	mockActivityRepo := &mockActivityLogRepository{}
	svc := NewTaskService(mockTaskRepo, mockActivityRepo)

	// Act
	err := svc.DeleteTask(context.Background(), uuid.New(), userID)

	// Assert
	if err == nil {
		t.Fatal("expected error for non-existent task")
	}
	if !errors.Is(err, repository.ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

// =============================================================================
// Tests for GetTaskActivity
// =============================================================================

func TestGetTaskActivity_Success(t *testing.T) {
	// Arrange
	userID := uuid.New()
	taskID := uuid.New()
	mockTaskRepo := &mockTaskRepository{}

	// Create activity logs
	logsList := []models.ActivityLog{}
	for i := 0; i < 3; i++ {
		logsList = append(logsList, models.ActivityLog{
			ID:     uuid.New(),
			TaskID: taskID,
			UserID: userID,
			Action: "task_updated",
			Details: map[string]interface{}{
				"field": "status",
			},
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Hour),
		})
	}

	mockActivityRepo := &mockActivityLogRepository{
		listByTaskFunc: func(ctx context.Context, taskID uuid.UUID, limit, offset int) ([]models.ActivityLog, int, error) {
			return logsList, len(logsList), nil
		},
	}
	svc := NewTaskService(mockTaskRepo, mockActivityRepo)

	// Act
	logs, total, err := svc.GetTaskActivity(context.Background(), taskID, 10, 0)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if total != 3 {
		t.Errorf("expected total 3, got %d", total)
	}
	if len(logs) != 3 {
		t.Errorf("expected 3 logs, got %d", len(logs))
	}
}

// =============================================================================
// Tests for helper functions
// =============================================================================

func TestHasDateChanged(t *testing.T) {
	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)

	tests := []struct {
		name string
		old  *time.Time
		new  *time.Time
		want bool
	}{
		{"both nil", nil, nil, false},
		{"old nil, new set", nil, &now, true},
		{"old set, new nil", &now, nil, true},
		{"same time", &now, &now, false},
		{"different time", &now, &tomorrow, true},
		{"equal but different instances", &now, func() *time.Time { t := now; return &t }(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasDateChanged(tt.old, tt.new)
			if got != tt.want {
				t.Errorf("hasDateChanged(%v, %v) = %v, want %v", tt.old, tt.new, got, tt.want)
			}
		})
	}
}

func TestFormatTimePtr(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name string
		t    *time.Time
		want string
	}{
		{"nil time", nil, ""},
		{"valid time", &now, "2024-01-15T10:30:00Z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTimePtr(tt.t)
			if got != tt.want {
				t.Errorf("formatTimePtr() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildUpdateDetails(t *testing.T) {
	// Arrange
	userID := uuid.New()
	oldTask := existingTask(userID)
	newTask := *oldTask
	newTask.Title = "Updated Title"
	newTask.Status = "in_progress"
	newTask.Priority = "high"
	newDueDate := time.Now().Add(14 * 24 * time.Hour)
	newTask.DueDate = &newDueDate

	svc := NewTaskService(nil, nil)

	// Act
	details := svc.buildUpdateDetails(oldTask, &newTask)

	// Assert
	if _, exists := details["title"]; !exists {
		t.Error("expected title change in details")
	}
	if _, exists := details["status"]; !exists {
		t.Error("expected status change in details")
	}
	if _, exists := details["priority"]; !exists {
		t.Error("expected priority change in details")
	}
	if _, exists := details["due_date"]; !exists {
		t.Error("expected due_date change in details")
	}
}