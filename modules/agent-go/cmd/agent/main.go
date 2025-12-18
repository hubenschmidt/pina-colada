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

	"github.com/pina-colada-co/agent-go/internal/agent"
	"github.com/pina-colada-co/agent-go/internal/agent/utils"
	"github.com/pina-colada-co/agent-go/internal/config"
	"github.com/pina-colada-co/agent-go/internal/controllers"
	"github.com/pina-colada-co/agent-go/internal/repositories"
	"github.com/pina-colada-co/agent-go/internal/routes"
	"github.com/pina-colada-co/agent-go/internal/services"
	"github.com/pina-colada-co/agent-go/pkg/db"
)

func main() {
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

	// Initialize storage repository
	storageRepo := repositories.GetStorageRepository()
	log.Printf("Initialized storage repository")

	// Initialize repositories
	userRepo := repositories.NewUserRepository(database)
	jobRepo := repositories.NewJobRepository(database)
	orgRepo := repositories.NewOrganizationRepository(database)
	indRepo := repositories.NewIndividualRepository(database)
	taskRepo := repositories.NewTaskRepository(database)
	contactRepo := repositories.NewContactRepository(database)
	prefsRepo := repositories.NewPreferenceRepository(database)
	projectRepo := repositories.NewProjectRepository(database)
	convRepo := repositories.NewConversationRepository(database)
	lookupRepo := repositories.NewLookupRepository(database)
	noteRepo := repositories.NewNoteRepository(database)
	commentRepo := repositories.NewCommentRepository(database)
	docRepo := repositories.NewDocumentRepository(database)
	accountRepo := repositories.NewAccountRepository(database)
	leadRepo := repositories.NewLeadRepository(database)
	provenanceRepo := repositories.NewProvenanceRepository(database)
	techRepo := repositories.NewTechnologyRepository(database)
	usageRepo := repositories.NewUsageRepository(database)
	reportRepo := repositories.NewReportRepository(database)
	notifRepo := repositories.NewNotificationRepository(database)
	agentConfigRepo := repositories.NewAgentConfigRepository(database)
	metricRepo := repositories.NewMetricRepository(database)

	// Initialize services
	authService := services.NewAuthService(userRepo)
	jobService := services.NewJobService(jobRepo, orgRepo, indRepo, lookupRepo)
	orgService := services.NewOrganizationService(orgRepo)
	indService := services.NewIndividualService(indRepo)
	taskService := services.NewTaskService(taskRepo)
	contactService := services.NewContactService(contactRepo)
	prefsService := services.NewPreferenceService(prefsRepo)
	projectService := services.NewProjectService(projectRepo)
	convService := services.NewConversationService(convRepo)
	lookupService := services.NewLookupService(lookupRepo)
	noteService := services.NewNoteService(noteRepo)
	commentService := services.NewCommentService(commentRepo)
	docService := services.NewDocumentService(docRepo, storageRepo)
	accountService := services.NewAccountService(accountRepo)
	leadService := services.NewLeadService(leadRepo)
	costsService := services.NewCostService()
	provenanceService := services.NewProvenanceService(provenanceRepo)
	techService := services.NewTechnologyService(techRepo)
	usageService := services.NewUsageService(usageRepo, userRepo)
	reportService := services.NewReportService(reportRepo)
	notifService := services.NewNotificationService(notifRepo)
	agentConfigService := services.NewAgentConfigService(agentConfigRepo)
	metricService := services.NewMetricService(metricRepo, agentConfigService)

	// Initialize config cache for agent orchestrator
	configCache := utils.NewConfigCache(agentConfigService)

	// Initialize ADK agent orchestrator (only if OpenAI API key is configured)
	agentOrchestrator := initOrchestrator(cfg, indService, orgService, docService, jobService, convService, configCache, metricService)

	// Initialize controllers
	ctrls := &routes.Controllers{
		Auth:         controllers.NewAuthController(authService),
		Job:          controllers.NewJobController(jobService),
		Organization: controllers.NewOrganizationController(orgService),
		Individual:   controllers.NewIndividualController(indService),
		Task:         controllers.NewTaskController(taskService),
		Contact:      controllers.NewContactController(contactService),
		Preferences:  controllers.NewPreferenceController(prefsService),
		Notification: controllers.NewNotificationController(notifService),
		Project:      controllers.NewProjectController(projectService),
		Conversation: controllers.NewConversationController(convService),
		Lookup:       controllers.NewLookupController(lookupService),
		Note:         controllers.NewNoteController(noteService),
		Comment:      controllers.NewCommentController(commentService),
		Document:     controllers.NewDocumentController(docService),
		Account:      controllers.NewAccountController(accountService),
		Lead:         controllers.NewLeadController(leadService, jobService),
		Costs:        controllers.NewCostController(costsService),
		Provenance:   controllers.NewProvenanceController(provenanceService),
		Technology:   controllers.NewTechnologyController(techService),
		Usage:        controllers.NewUsageController(usageService),
		Report:       controllers.NewReportController(reportService),
		Agent:        controllers.NewAgentController(agentOrchestrator),
		AgentConfig:  controllers.NewAgentConfigController(agentConfigService, configCache),
		WebSocket:    controllers.NewWebSocketController(agentOrchestrator),
		Metric:       controllers.NewMetricController(metricService),
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

// initOrchestrator initializes the agent orchestrator if configured
func initOrchestrator(cfg *config.Config, indService *services.IndividualService, orgService *services.OrganizationService, docService *services.DocumentService, jobService *services.JobService, convService *services.ConversationService, configCache *utils.ConfigCache, metricService *services.MetricService) *agent.Orchestrator {
	if cfg.OpenAIAPIKey == "" {
		log.Println("OPENAI_API_KEY not configured - agent endpoints disabled")
		return nil
	}

	ctx := context.Background()
	orchestrator, err := agent.NewOrchestrator(ctx, cfg, indService, orgService, docService, jobService, convService, configCache, metricService)
	if err != nil {
		log.Printf("Warning: Failed to initialize agent orchestrator: %v", err)
		log.Printf("Agent endpoints will return errors until configured")
		return nil
	}

	log.Println("ADK agent orchestrator initialized successfully")
	return orchestrator
}
