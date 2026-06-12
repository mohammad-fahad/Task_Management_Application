package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/taskmanager/backend/internal/handlers"
	"github.com/taskmanager/backend/internal/middleware"
	"github.com/taskmanager/backend/internal/repository"
	"github.com/taskmanager/backend/internal/service"
)

func main() {
	// =========================================================================
	// Configuration
	// =========================================================================
	cfg := loadConfig()

	// =========================================================================
	// Database Connection
	// =========================================================================
	db, err := openDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// =========================================================================
	// Run Migrations
	// =========================================================================
	if cfg.RunMigrations {
		if err := runMigrations(db, cfg.MigrationsPath); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		log.Println("Database migrations completed successfully")
	}

	// =========================================================================
	// Initialize Repositories (Data Layer)
	// =========================================================================
	userRepo := repository.NewPostgresUserRepository(db)
	taskRepo := repository.NewPostgresTaskRepository(db)
	activityRepo := repository.NewPostgresActivityLogRepository(db)

	// =========================================================================
	// Initialize Services (Business Logic Layer)
	// =========================================================================
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	taskService := service.NewTaskService(taskRepo, activityRepo)

	// =========================================================================
	// Initialize Middleware
	// =========================================================================
	authGuard := middleware.NewAuthGuard(authService)
	recoveryMiddleware := middleware.NewRecoveryMiddleware(cfg.Debug)

	// =========================================================================
	// Initialize Handlers (Presentation Layer)
	// =========================================================================
	authHandler := handlers.NewAuthHandler(authService)
	taskHandler := handlers.NewTaskHandler(taskService)

	// =========================================================================
	// Router Setup
	// =========================================================================
	r := chi.NewRouter()

	// Global middleware
	r.Use(recoveryMiddleware.Recovery)
	r.Use(middleware.RequestLoggingMiddleware)
	r.Use(middleware.CORSMiddleware)
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Timeout(30 * time.Second))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
	})

	// =========================================================================
	// API Routes v1
	// =========================================================================
	r.Route("/api/v1", func(r chi.Router) {
		// Public auth routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/logout", authHandler.Logout)
		})

		// Protected routes (require authentication)
		r.Group(func(r chi.Router) {
			r.Use(authGuard.Authenticate)

			// Current user
			r.Get("/auth/me", authHandler.Me)

			// Tasks routes
			r.Route("/tasks", func(r chi.Router) {
				r.Post("/", taskHandler.CreateTask)
				r.Get("/", taskHandler.ListTasks)
				r.Get("/{id}", taskHandler.GetTask)
				r.Put("/{id}", taskHandler.UpdateTask)
				r.Delete("/{id}", taskHandler.DeleteTask)
				r.Get("/{id}/activity", taskHandler.GetTaskActivity)
			})

			// Admin-only routes
			r.Route("/admin", func(r chi.Router) {
				r.Use(authGuard.RequireRole("admin"))
				// Admin-specific routes can be added here
			})
		})
	})

	// =========================================================================
	// Server Configuration
	// =========================================================================
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// =========================================================================
	// Graceful Shutdown
	// =========================================================================
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		log.Printf("Environment: %s", cfg.Environment)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	<-done
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}

// Config holds all application configuration.
type Config struct {
	Port           string
	Environment    string
	DatabaseURL    string
	JWTSecret      string
	MigrationsPath string
	RunMigrations  bool
	Debug          bool
}

// loadConfig loads configuration from environment variables with defaults.
func loadConfig() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/taskmanager?sslmode=disable"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "change-me-in-production-use-a-strong-secret-key"
	}

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "file://migrations"
	}

	runMigrations := os.Getenv("RUN_MIGRATIONS") != "false"

	debug := os.Getenv("DEBUG") == "true"

	return Config{
		Port:           port,
		Environment:    env,
		DatabaseURL:    dbURL,
		JWTSecret:      jwtSecret,
		MigrationsPath: migrationsPath,
		RunMigrations:  runMigrations,
		Debug:          debug,
	}
}

// openDB opens a PostgreSQL connection pool.
func openDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection established")
	return db, nil
}

// runMigrations executes database migrations.
func runMigrations(db *sql.DB, migrationsPath string) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}