package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pina-colada-co/agent-go/internal/config"
	"github.com/pina-colada-co/agent-go/internal/controllers"
	"github.com/pina-colada-co/agent-go/internal/repositories"
	"github.com/pina-colada-co/agent-go/internal/routes"
	"github.com/pina-colada-co/agent-go/internal/services"
	"github.com/pina-colada-co/agent-go/pkg/db"
	"github.com/pina-colada-co/agent-go/pkg/s3"
)

func main() {
	ctx := context.Background()

	// Load configuration
	cfg := config.Load()
	log.Printf("Starting agent-go in %s mode on port %s", cfg.Env, cfg.Port)

	// Initialize database
	database, err := db.Connect(cfg.DatabaseURL, cfg.Env == "development")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	sqlDB, _ := database.DB()
	defer sqlDB.Close()

	// Initialize S3 client
	s3Client, err := s3.NewClient(ctx, cfg.AWSRegion, cfg.S3Bucket)
	if err != nil {
		log.Printf("Warning: Failed to initialize S3 client: %v", err)
	}
	_ = s3Client // Will be used by document services

	// Initialize repositories
	userRepo := repositories.NewUserRepository(database)
	jobRepo := repositories.NewJobRepository(database)
	orgRepo := repositories.NewOrganizationRepository(database)
	indRepo := repositories.NewIndividualRepository(database)
	taskRepo := repositories.NewTaskRepository(database)
	contactRepo := repositories.NewContactRepository(database)
	prefsRepo := repositories.NewPreferencesRepository(database)

	// Initialize services
	authService := services.NewAuthService(userRepo)
	jobService := services.NewJobService(jobRepo)
	orgService := services.NewOrganizationService(orgRepo)
	indService := services.NewIndividualService(indRepo)
	taskService := services.NewTaskService(taskRepo)
	contactService := services.NewContactService(contactRepo)
	prefsService := services.NewPreferencesService(prefsRepo)

	// Initialize controllers
	ctrls := &routes.Controllers{
		Auth:         controllers.NewAuthController(authService),
		Job:          controllers.NewJobController(jobService),
		Organization: controllers.NewOrganizationController(orgService),
		Individual:   controllers.NewIndividualController(indService),
		Task:         controllers.NewTaskController(taskService),
		Contact:      controllers.NewContactController(contactService),
		Preferences:  controllers.NewPreferencesController(prefsService),
	}

	// Initialize router and register routes
	router := routes.NewRouter()
	routes.RegisterRoutes(router, ctrls, authService)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server listening on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
