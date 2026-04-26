package models

import (
	"time"

	"github.com/google/uuid"
)

// Todo represents a todo item in the database
type Todo struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	Priority  int       `json:"priority"`
	DueDate   *time.Time `json:"due_date,omitempty"`
	Notes     string    `json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TodoCreateRequest represents the request to create a new todo
type TodoCreateRequest struct {
	Title    string     `json:"title" validate:"required,min=1,max=255"`
	Priority int        `json:"priority" validate:"min=1,max=5"`
	DueDate  *time.Time `json:"due_date,omitempty"`
	Notes    string     `json:"notes" validate:"max=1000"`
}

// TodoUpdateRequest represents the request to update a todo
type TodoUpdateRequest struct {
	Title     *string    `json:"title,omitempty"`
	Completed *bool      `json:"completed,omitempty"`
	Priority  *int       `json:"priority,omitempty"`
	DueDate   *time.Time `json:"due_date,omitempty"`
	Notes     *string    `json:"notes,omitempty"`
}

// User represents a user in the database
type User struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username" validate:"required,min=3,max=50,alphanum"`
	Email    string    `json:"email" validate:"required,email"`
	Password string    `json:"-" validate:"required,min=8"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserCreateRequest represents the request to create a new user
type UserCreateRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// UserLoginRequest represents the request to login
type UserLoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// APIResponse standard JSON API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ErrorResponse standard error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// TodoStats holds statistics about todos
type TodoStats struct {
	Total          int64   `json:"total"`
	Completed      int64   `json:"completed"`
	Pending        int64   `json:"pending"`
	Overdue        int64   `json:"overdue"`
	Today          int64   `json:"today"`
	CompletionRate float64 `json:"completion_rate"`
}