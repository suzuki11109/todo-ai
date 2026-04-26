package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"github.com/aki/todo-ai/internal/models"
	"github.com/aki/todo-ai/internal/repository"
	"github.com/aki/todo-ai/internal/middleware"
	"go.uber.org/zap"
)

type UserHandler struct {
	userRepo   *repository.UserRepository
	todoRepo   *repository.TodoRepository
	auth       *middleware.AuthMiddleware
	logger     *zap.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(
	userRepo *repository.UserRepository,
	todoRepo *repository.TodoRepository,
	auth *middleware.AuthMiddleware,
	logger *zap.Logger,
) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
		todoRepo: todoRepo,
		auth:     auth,
		logger:   logger,
	}
}

// Register handles POST /api/auth/register
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.UserCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate username and email availability
	exists, err := h.userRepo.ExistsByUsername(r.Context(), req.Username)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Database error", err.Error())
		return
	}
	if exists {
		respondError(w, http.StatusConflict, "Username already taken", "")
		return
	}

	exists, err = h.userRepo.ExistsByEmail(r.Context(), req.Email)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Database error", err.Error())
		return
	}
	if exists {
		respondError(w, http.StatusConflict, "Email already registered", "")
		return
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to hash password", err.Error())
		return
	}

	user := &models.User{
		ID:       uuid.New(),
		Username: req.Username,
		Email:    req.Email,
		Password: string(passwordHash),
	}

	// Create user
	if err := h.userRepo.Create(r.Context(), user); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create user", err.Error())
		return
	}

	// Generate JWT token
	token, err := h.auth.GenerateToken(user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate token", err.Error())
		return
	}

	h.logger.Info("User registered",
		zap.String("username", user.Username),
		zap.String("user_id", user.ID.String()),
	)

	respondJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "User registered successfully",
		Data: map[string]interface{}{
			"user": map[string]interface{}{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
			},
			"token": token,
		},
	})
}

// Login handles POST /api/auth/login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user by username
	user, err := h.userRepo.GetByUsername(r.Context(), req.Username)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Database error", err.Error())
		return
	}

	if user == nil {
		respondError(w, http.StatusUnauthorized, "Invalid credentials", "")
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		h.logger.Warn("Failed login attempt",
			zap.String("username", req.Username),
			zap.String("user_id", user.ID.String()),
		)
		respondError(w, http.StatusUnauthorized, "Invalid credentials", "")
		return
	}

	// Generate JWT token
	token, err := h.auth.GenerateToken(user.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate token", err.Error())
		return
	}

	h.logger.Info("User logged in",
		zap.String("username", user.Username),
		zap.String("user_id", user.ID.String()),
	)

	respondJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Login successful",
		Data: map[string]interface{}{
			"user": map[string]interface{}{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
			},
			"token": token,
		},
	})
}

// RefreshToken handles POST /api/auth/refresh
func (h *UserHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	newToken, err := h.auth.RefreshToken(req.Token)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid token", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Token refreshed successfully",
		Data: map[string]interface{}{
			"token": newToken,
		},
	})
}

// Me handles GET /api/auth/me
func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized", "user not authenticated")
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Database error", err.Error())
		return
	}

	if user == nil {
		respondError(w, http.StatusNotFound, "User not found", "")
		return
	}

	// Return user without password hash
	respondJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"created_at": user.CreatedAt,
		},
	})
}

// GenerateInviteCode generates a one-time invite code (for future features)
func (h *UserHandler) GenerateInviteCode(w http.ResponseWriter, r *http.Request) {
	// Generate random 32-character hex code
	bytes := make([]byte, 16)
	rand.Read(bytes)
	code := hex.EncodeToString(bytes)

	respondJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Invite code generated",
		Data: map[string]interface{}{
			"code": code,
			"expires_at": time.Now().Add(24 * time.Hour),
		},
	})
}