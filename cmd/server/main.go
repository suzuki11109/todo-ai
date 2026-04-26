package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aki/todo-ai/internal/config"
	"github.com/aki/todo-ai/internal/handler"
	"github.com/aki/todo-ai/internal/middleware"
	"github.com/aki/todo-ai/internal/repository"
	"github.com/aki/todo-ai/pkg/database"
	"github.com/aki/todo-ai/pkg/logger"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger, err := logger.New(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	logger.Info("Starting TodoAI server",
		zap.String("app", cfg.App.Name),
		zap.String("env", cfg.App.Env),
		zap.String("port", cfg.App.Port),
	)

	// Initialize database
	dbConfig := &database.Config{
		Host:         cfg.Database.Host,
		Port:         cfg.Database.Port,
		User:         cfg.Database.User,
		Password:     cfg.Database.Password,
		Name:         cfg.Database.Name,
		SSLMode:      cfg.Database.SSLMode,
		MaxOpenConns: cfg.Database.MaxOpenConns,
		MaxIdleConns: cfg.Database.MaxIdleConns,
		MaxConnLifetime: cfg.Database.MaxConnLifetime,
	}

	db, err := database.New(dbConfig, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize repositories
	todoRepo := repository.NewTodoRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Initialize handlers
	healthHandler := handler.NewHealthHandler(db, logger)
	todoHandler := handler.NewTodoHandler(todoRepo, logger)
	userHandler := handler.NewUserHandler(userRepo, todoRepo, nil, logger) // auth middleware will be added later

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(viper.GetString("JWT_SECRET"), logger)

	// Update user handler with auth
	userHandler = handler.NewUserHandler(userRepo, todoRepo, authMiddleware, logger)

	// Initialize middleware
	corsMiddleware := middleware.NewCORSMiddleware(cfg.CORS, logger)
	recoveryMiddleware := middleware.NewRecoveryMiddleware(logger)
	timeoutMiddleware := middleware.NewTimeoutMiddleware(30 * time.Second)
	rateLimitMiddleware := middleware.NewRateLimiterMiddleware(100, 200, logger) // 100 r/s, burst 200
	loggingMiddleware := middleware.NewLoggingMiddleware(logger)

	// Setup router
	router := mux.NewRouter()

	// Apply global middleware
	router.Use(recoveryMiddleware.Middleware)
	router.Use(loggingMiddleware.Middleware)
	router.Use(corsMiddleware.Middleware)
	router.Use(rateLimitMiddleware.Middleware)
	router.Use(timeoutMiddleware.Middleware)

	// Health endpoints (no auth required)
	router.HandleFunc("/health", healthHandler.HealthCheck).Methods("GET")
	router.HandleFunc("/readiness", healthHandler.ReadinessCheck).Methods("GET")

	// Public auth routes (no authentication required)
	authPublic := router.PathPrefix("/api/auth").Subrouter()
	authPublic.HandleFunc("/register", userHandler.Register).Methods("POST")
	authPublic.HandleFunc("/login", userHandler.Login).Methods("POST")
	authPublic.HandleFunc("/refresh", userHandler.RefreshToken).Methods("POST")

	// Protected API routes (require authentication)
	api := router.PathPrefix("/api").Subrouter()
	api.Use(authMiddleware.Middleware)

	// Protected auth: get current user
	api.HandleFunc("/auth/me", userHandler.Me).Methods("GET")

	// Todo routes
	todos := api.PathPrefix("/todos").Subrouter()
	todos.HandleFunc("", todoHandler.CreateTodo).Methods("POST")
	todos.HandleFunc("", todoHandler.ListTodos).Methods("GET")
	todos.HandleFunc("/{id}", todoHandler.GetTodo).Methods("GET")
	todos.HandleFunc("/{id}", todoHandler.UpdateTodo).Methods("PUT")
	todos.HandleFunc("/{id}", todoHandler.DeleteTodo).Methods("DELETE")
	todos.HandleFunc("/today", todoHandler.GetTodayTodos).Methods("GET")
	todos.HandleFunc("/overdue", todoHandler.GetOverdueTodos).Methods("GET")
	todos.HandleFunc("/stats", todoHandler.GetStats).Methods("GET")

	// Serve static files
	if cfg.App.Env == "development" {
		router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	} else {
		router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	}

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.App.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Server listening", zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit

	logger.Info("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}