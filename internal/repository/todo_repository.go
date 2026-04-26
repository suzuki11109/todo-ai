package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/aki/todo-ai/internal/models"
	"github.com/aki/todo-ai/pkg/database"
)

var (
	ErrNotFound = sql.ErrNoRows
)

type TodoRepository struct {
	db *database.DB
}

// NewTodoRepository creates a new todo repository
func NewTodoRepository(db *database.DB) *TodoRepository {
	return &TodoRepository{db: db}
}

// Create inserts a new todo into the database
func (r *TodoRepository) Create(ctx context.Context, todo *models.Todo) error {
	if todo.ID == uuid.Nil {
		todo.ID = uuid.New()
	}

	query := `
		INSERT INTO todos (id, user_id, title, completed, priority, due_date, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		todo.ID, todo.UserID, todo.Title, todo.Completed,
		todo.Priority, todo.DueDate, todo.Notes,
	).Scan(&todo.CreatedAt, &todo.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert todo: %w", err)
	}

	return nil
}

// GetByID retrieves a todo by its ID
func (r *TodoRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Todo, error) {
	todo := &models.Todo{}
	query := `
		SELECT id, user_id, title, completed, priority, due_date, notes, created_at, updated_at
		FROM todos
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&todo.ID, &todo.UserID, &todo.Title, &todo.Completed,
		&todo.Priority, &todo.DueDate, &todo.Notes, &todo.CreatedAt, &todo.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found, not an error
		}
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}

	return todo, nil
}

// GetByUserID retrieves all todos for a specific user with optional filtering
func (r *TodoRepository) GetByUserID(ctx context.Context, userID uuid.UUID, completed *bool, limit, offset int) ([]models.Todo, error) {
	query := `
		SELECT id, user_id, title, completed, priority, due_date, notes, created_at, updated_at
		FROM todos
		WHERE user_id = $1
	`
	args := []interface{}{userID}
	argCount := 1

	// Add optional filters
	if completed != nil {
		argCount++
		query += fmt.Sprintf(" AND completed = $%d", argCount)
		args = append(args, *completed)
	}

	// Add ordering
	query += " ORDER BY priority ASC, due_date ASC NULLS LAST, created_at DESC"

	// Add pagination
	if limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, limit)

		if offset > 0 {
			argCount++
			query += fmt.Sprintf(" OFFSET $%d", argCount)
			args = append(args, offset)
		}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query todos: %w", err)
	}
	defer rows.Close()

	var todos []models.Todo
	for rows.Next() {
		var todo models.Todo
		if err := rows.Scan(
			&todo.ID, &todo.UserID, &todo.Title, &todo.Completed,
			&todo.Priority, &todo.DueDate, &todo.Notes, &todo.CreatedAt, &todo.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan todo: %w", err)
		}
		todos = append(todos, todo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return todos, nil
}

// Update updates an existing todo
func (r *TodoRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil // Nothing to update
	}

	// Build dynamic UPDATE query
	setClause := ""
	args := []interface{}{id}
	argCount := 1

	for key, value := range updates {
		argCount++
		setClause += fmt.Sprintf("%s = $%d, ", key, argCount)
		args = append(args, value)
	}

	// Remove trailing comma and space
	setClause = setClause[:len(setClause)-2]

	query := fmt.Sprintf(`
		UPDATE todos
		SET %s
		WHERE id = $1
	`, setClause)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update todo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Delete removes a todo by ID
func (r *TodoRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	query := `DELETE FROM todos WHERE id = $1 AND user_id = $2`
	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// CountByUserID gets the count of todos for a user (optionally filtered)
func (r *TodoRepository) CountByUserID(ctx context.Context, userID uuid.UUID, completed *bool) (int64, error) {
	query := `SELECT COUNT(*) FROM todos WHERE user_id = $1`
	args := []interface{}{userID}
	argCount := 1

	if completed != nil {
		argCount++
		query += fmt.Sprintf(" AND completed = $%d", argCount)
		args = append(args, *completed)
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count todos: %w", err)
	}

	return count, nil
}

// GetTodaysTodos retrieves todos due today
func (r *TodoRepository) GetTodaysTodos(ctx context.Context, userID uuid.UUID) ([]models.Todo, error) {
	today := time.Now().Format("2006-01-02")
	query := `
		SELECT id, user_id, title, completed, priority, due_date, notes, created_at, updated_at
		FROM todos
		WHERE user_id = $1 AND DATE(due_date) = $2 AND completed = false
		ORDER BY priority ASC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, today)
	if err != nil {
		return nil, fmt.Errorf("failed to query today's todos: %w", err)
	}
	defer rows.Close()

	var todos []models.Todo
	for rows.Next() {
		var todo models.Todo
		if err := rows.Scan(
			&todo.ID, &todo.UserID, &todo.Title, &todo.Completed,
			&todo.Priority, &todo.DueDate, &todo.Notes, &todo.CreatedAt, &todo.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan todo: %w", err)
		}
		todos = append(todos, todo)
	}

	return todos, nil
}

// GetOverdueTodos retrieves overdue todos
func (r *TodoRepository) GetOverdueTodos(ctx context.Context, userID uuid.UUID) ([]models.Todo, error) {
	query := `
		SELECT id, user_id, title, completed, priority, due_date, notes, created_at, updated_at
		FROM todos
		WHERE user_id = $1 AND due_date < $2 AND completed = false
		ORDER BY due_date ASC, priority DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query overdue todos: %w", err)
	}
	defer rows.Close()

	var todos []models.Todo
	for rows.Next() {
		var todo models.Todo
		if err := rows.Scan(
			&todo.ID, &todo.UserID, &todo.Title, &todo.Completed,
			&todo.Priority, &todo.DueDate, &todo.Notes, &todo.CreatedAt, &todo.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan todo: %w", err)
		}
		todos = append(todos, todo)
	}

	return todos, nil
}