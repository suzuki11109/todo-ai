package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/google/uuid"
	"github.com/aki/todo-ai/internal/models"
	"github.com/aki/todo-ai/internal/repository"
	"go.uber.org/zap"
)

type TodoHandler struct {
	repo     *repository.TodoRepository
	logger   *zap.Logger
}

// NewTodoHandler creates a new todo handler
func NewTodoHandler(repo *repository.TodoRepository, logger *zap.Logger) *TodoHandler {
	return &TodoHandler{repo: repo, logger: logger}
}

// CreateTodo handles POST /api/todos
func (h *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	var req models.TodoCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized", "user not authenticated")
		return
	}

	todo := &models.Todo{
		UserID:    userID,
		Title:     req.Title,
		Completed: false,
		Priority:  req.Priority,
		DueDate:   req.DueDate,
		Notes:     req.Notes,
	}

	if err := h.repo.Create(r.Context(), todo); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create todo", err.Error())
		return
	}

	h.logger.Info("Todo created", zap.String("id", todo.ID.String()))
	respondJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Todo created successfully",
		Data:    todo,
	})
}

// GetTodo handles GET /api/todos/{id}
func (h *TodoHandler) GetTodo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		respondError(w, http.StatusBadRequest, "Invalid todo ID", "missing ID in URL")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid todo ID", err.Error())
		return
	}

	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized", "user not authenticated")
		return
	}

	todo, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get todo", err.Error())
		return
	}

	if todo == nil {
		respondError(w, http.StatusNotFound, "Todo not found", "no todo with that ID")
		return
	}

	// Check ownership
	if todo.UserID != userID {
		respondError(w, http.StatusForbidden, "Forbidden", "you don't own this todo")
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    todo,
	})
}

// ListTodos handles GET /api/todos
func (h *TodoHandler) ListTodos(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized", "user not authenticated")
		return
	}

	// Parse query parameters
	completedStr := r.URL.Query().Get("completed")
	var completed *bool
	if completedStr != "" {
		c, err := strconv.ParseBool(completedStr)
		if err == nil {
			completed = &c
		}
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50 // default
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	todos, err := h.repo.GetByUserID(r.Context(), userID, completed, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list todos", err.Error())
		return
	}

	count, err := h.repo.CountByUserID(r.Context(), userID, completed)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to count todos", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"todos": todos,
			"count": count,
			"limit": limit,
			"offset": offset,
		},
	})
}

// UpdateTodo handles PUT /api/todos/{id}
func (h *TodoHandler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		respondError(w, http.StatusBadRequest, "Invalid todo ID", "missing ID in URL")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid todo ID", err.Error())
		return
	}

	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized", "user not authenticated")
		return
	}

	// Verify ownership first
	todo, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get todo", err.Error())
		return
	}
	if todo == nil || todo.UserID != userID {
		respondError(w, http.StatusNotFound, "Todo not found", "")
		return
	}

	var req models.TodoUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Build update map
	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Completed != nil {
		updates["completed"] = *req.Completed
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.DueDate != nil {
		updates["due_date"] = req.DueDate
	}
	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}

	if len(updates) == 0 {
		respondError(w, http.StatusBadRequest, "No updates provided", "")
		return
	}

	if err := h.repo.Update(r.Context(), id, updates); err != nil {
		if err == sql.ErrNoRows {
			respondError(w, http.StatusNotFound, "Todo not found", "")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update todo", err.Error())
		return
	}

	h.logger.Info("Todo updated", zap.String("id", id.String()))
	respondJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Todo updated successfully",
	})
}

// DeleteTodo handles DELETE /api/todos/{id}
func (h *TodoHandler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		respondError(w, http.StatusBadRequest, "Invalid todo ID", "missing ID in URL")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid todo ID", err.Error())
		return
	}

	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized", "user not authenticated")
		return
	}

	if err := h.repo.Delete(r.Context(), id, userID); err != nil {
		if err == sql.ErrNoRows {
			respondError(w, http.StatusNotFound, "Todo not found", "")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete todo", err.Error())
		return
	}

	h.logger.Info("Todo deleted", zap.String("id", id.String()))
	respondJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Todo deleted successfully",
	})
}

// GetTodayTodos handles GET /api/todos/today
func (h *TodoHandler) GetTodayTodos(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized", "user not authenticated")
		return
	}

	todos, err := h.repo.GetTodaysTodos(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get today's todos", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    todos,
	})
}

// GetOverdueTodos handles GET /api/todos/overdue
func (h *TodoHandler) GetOverdueTodos(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized", "user not authenticated")
		return
	}

	todos, err := h.repo.GetOverdueTodos(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get overdue todos", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    todos,
	})
}

// GetStats handles GET /api/todos/stats
func (h *TodoHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized", "user not authenticated")
		return
	}

	total, err := h.repo.CountByUserID(r.Context(), userID, nil)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get stats", err.Error())
		return
	}

	completed, err := h.repo.CountByUserID(r.Context(), userID, boolPtr(true))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get stats", err.Error())
		return
	}

	pending, err := h.repo.CountByUserID(r.Context(), userID, boolPtr(false))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get stats", err.Error())
		return
	}

	overdue, err := h.repo.GetOverdueTodos(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get stats", err.Error())
		return
	}

	today, err := h.repo.GetTodaysTodos(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get stats", err.Error())
		return
	}

	stats := map[string]interface{}{
		"total":    total,
		"completed": completed,
		"pending":  pending,
		"overdue":  len(overdue),
		"today":    len(today),
		"completion_rate": func() float64 {
			if total == 0 {
				return 0
			}
			return float64(completed) / float64(total) * 100
		}(),
	}

	respondJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    stats,
	})
}

// respondJSON writes a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError writes an error response
func respondError(w http.ResponseWriter, status int, message, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.ErrorResponse{
		Error:   message,
		Details: details,
	})
}

func boolPtr(b bool) *bool {
	return &b
}