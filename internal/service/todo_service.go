package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/aki/todo-ai/internal/models"
	"github.com/aki/todo-ai/internal/repository"
)

type TodoService struct {
	todoRepo *repository.TodoRepository
}

// NewTodoService creates a new todo service
func NewTodoService(todoRepo *repository.TodoRepository) *TodoService {
	return &TodoService{todoRepo: todoRepo}
}

// CreateTodo creates a new todo item
func (s *TodoService) CreateTodo(ctx context.Context, userID uuid.UUID, req *models.TodoCreateRequest) (*models.Todo, error) {
	// Validate
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	if req.Priority < 1 || req.Priority > 5 {
		req.Priority = 3 // default priority
	}

	todo := &models.Todo{
		UserID:    userID,
		Title:     req.Title,
		Completed: false,
		Priority:  req.Priority,
		DueDate:   req.DueDate,
		Notes:     req.Notes,
	}

	if err := s.todoRepo.Create(ctx, todo); err != nil {
		return nil, fmt.Errorf("failed to create todo: %w", err)
	}

	return todo, nil
}

// GetTodo retrieves a todo by ID (ensuring ownership)
func (s *TodoService) GetTodo(ctx context.Context, userID, todoID uuid.UUID) (*models.Todo, error) {
	todo, err := s.todoRepo.GetByID(ctx, todoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}

	if todo == nil {
		return nil, fmt.Errorf("todo not found")
	}

	// Check ownership
	if todo.UserID != userID {
		return nil, fmt.Errorf("unauthorized: you don't own this todo")
	}

	return todo, nil
}

// ListTodos retrieves todos for a user with filters
func (s *TodoService) ListTodos(ctx context.Context, userID uuid.UUID, completed *bool, limit, offset int) ([]models.Todo, int64, error) {
	todos, err := s.todoRepo.GetByUserID(ctx, userID, completed, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list todos: %w", err)
	}

	count, err := s.todoRepo.CountByUserID(ctx, userID, completed)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count todos: %w", err)
	}

	return todos, count, nil
}

// UpdateTodo updates an existing todo (ensuring ownership)
func (s *TodoService) UpdateTodo(ctx context.Context, userID, todoID uuid.UUID, req *models.TodoUpdateRequest) error {
	// Verify ownership first
	_, err := s.GetTodo(ctx, userID, todoID)
	if err != nil {
		return err
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Completed != nil {
		updates["completed"] = *req.Completed
	}
	if req.Priority != nil {
		// Validate priority
		if *req.Priority < 1 || *req.Priority > 5 {
			return fmt.Errorf("priority must be between 1 and 5")
		}
		updates["priority"] = *req.Priority
	}
	if req.DueDate != nil {
		updates["due_date"] = req.DueDate
	}
	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}

	if len(updates) == 0 {
		return fmt.Errorf("no updates provided")
	}

	if err := s.todoRepo.Update(ctx, todoID, updates); err != nil {
		if err == repository.ErrNotFound {
			return fmt.Errorf("todo not found")
		}
		return fmt.Errorf("failed to update todo: %w", err)
	}

	return nil
}

// DeleteTodo deletes a todo (ensuring ownership)
func (s *TodoService) DeleteTodo(ctx context.Context, userID, todoID uuid.UUID) error {
	// Verify ownership first
	_, err := s.GetTodo(ctx, userID, todoID)
	if err != nil {
		return err
	}

	if err := s.todoRepo.Delete(ctx, todoID, userID); err != nil {
		if err == repository.ErrNotFound {
			return fmt.Errorf("todo not found")
		}
		return fmt.Errorf("failed to delete todo: %w", err)
	}

	return nil
}

// GetStats returns statistics for a user's todos
func (s *TodoService) GetStats(ctx context.Context, userID uuid.UUID) (*models.TodoStats, error) {
	total, err := s.todoRepo.CountByUserID(ctx, userID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	completed, err := s.todoRepo.CountByUserID(ctx, userID, boolPtr(true))
	if err != nil {
		return nil, fmt.Errorf("failed to get completed count: %w", err)
	}

	pending, err := s.todoRepo.CountByUserID(ctx, userID, boolPtr(false))
	if err != nil {
		return nil, fmt.Errorf("failed to get pending count: %w", err)
	}

	overdue, err := s.todoRepo.GetOverdueTodos(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue todos: %w", err)
	}

	today, err := s.todoRepo.GetTodaysTodos(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get today's todos: %w", err)
	}

	stats := &models.TodoStats{
		Total:          total,
		Completed:      completed,
		Pending:        pending,
		Overdue:        int64(len(overdue)),
		Today:          int64(len(today)),
		CompletionRate: calculateCompletionRate(total, completed),
	}

	return stats, nil
}

// GetOverdueTodos returns overdue todos
func (s *TodoService) GetOverdueTodos(ctx context.Context, userID uuid.UUID) ([]models.Todo, error) {
	return s.todoRepo.GetOverdueTodos(ctx, userID)
}

// GetTodaysTodos returns today's todos
func (s *TodoService) GetTodaysTodos(ctx context.Context, userID uuid.UUID) ([]models.Todo, error) {
	return s.todoRepo.GetTodaysTodos(ctx, userID)
}

// Helper functions
func calculateCompletionRate(total, completed int64) float64 {
	if total == 0 {
		return 0.0
	}
	return float64(completed) / float64(total) * 100
}

func boolPtr(b bool) *bool {
	return &b
}