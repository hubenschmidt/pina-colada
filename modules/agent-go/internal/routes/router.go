package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/pina-colada-co/agent-go/internal/controllers"
	appMiddleware "github.com/pina-colada-co/agent-go/internal/middleware"
)

// Controllers holds all controller instances
type Controllers struct {
	Auth         *controllers.AuthController
	Job          *controllers.JobController
	Organization *controllers.OrganizationController
	Individual   *controllers.IndividualController
	Task         *controllers.TaskController
	Contact      *controllers.ContactController
	Preferences  *controllers.PreferenceController
	Notification *controllers.NotificationController
	Project      *controllers.ProjectController
	Conversation *controllers.ConversationController
	Lookup       *controllers.LookupController
	Note         *controllers.NoteController
	Comment      *controllers.CommentController
	Document     *controllers.DocumentController
	Account      *controllers.AccountController
	Lead         *controllers.LeadController
	Costs        *controllers.CostController
	Provenance   *controllers.ProvenanceController
	Technology   *controllers.TechnologyController
	Usage        *controllers.UsageController
	Report       *controllers.ReportController
	Agent        *controllers.AgentController
	WebSocket    *controllers.WebSocketController
}

// NewRouter creates and configures the Chi router
func NewRouter() *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Tenant-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check endpoints (no auth required)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"agent-go"}`))
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"version":"0.1.0","env":"development","service":"agent-go"}`))
	})

	return r
}

// RegisterRoutes registers all API routes with controllers
func RegisterRoutes(r *chi.Mux, c *Controllers, userLoader appMiddleware.UserLoader) {
	// WebSocket route (no auth middleware - handles its own auth via messages)
	if c.WebSocket != nil {
		r.Get("/ws", c.WebSocket.HandleWS)
	}

	// Protected routes (require auth)
	r.Group(func(r chi.Router) {
		r.Use(appMiddleware.AuthMiddleware)
		r.Use(appMiddleware.UserLoaderMiddleware(userLoader))

		// Auth routes
		r.Route("/auth", func(r chi.Router) {
			r.Get("/me", c.Auth.GetMe)
			r.Post("/tenant/create", c.Auth.CreateTenant)
		})

		// Jobs routes
		r.Route("/jobs", func(r chi.Router) {
			r.Get("/", c.Job.GetJobs)
			r.Post("/", c.Job.CreateJob)
			r.Get("/recent-resume-date", c.Job.GetRecentResumeDate)
			r.Get("/{id}", c.Job.GetJob)
			r.Put("/{id}", c.Job.UpdateJob)
			r.Delete("/{id}", c.Job.DeleteJob)
		})

		// Organizations routes
		r.Route("/organizations", func(r chi.Router) {
			r.Get("/", c.Organization.GetOrganizations)
			r.Post("/", c.Organization.CreateOrganization)
			r.Get("/search", c.Organization.SearchOrganizations)
			r.Get("/{id}", c.Organization.GetOrganization)
			r.Put("/{id}", c.Organization.UpdateOrganization)
			r.Delete("/{id}", c.Organization.DeleteOrganization)
			r.Get("/{id}/contacts", c.Organization.GetOrganizationContacts)
			r.Post("/{id}/contacts", c.Organization.AddOrganizationContact)
			r.Put("/{id}/contacts/{contactId}", c.Organization.UpdateOrganizationContact)
			r.Delete("/{id}/contacts/{contactId}", c.Organization.DeleteOrganizationContact)
			r.Get("/{id}/technologies", c.Organization.GetOrganizationTechnologies)
			r.Post("/{id}/technologies", c.Organization.AddOrganizationTechnology)
			r.Delete("/{id}/technologies/{techId}", c.Organization.RemoveOrganizationTechnology)
			r.Get("/{id}/funding-rounds", c.Organization.GetOrganizationFundingRounds)
			r.Post("/{id}/funding-rounds", c.Organization.CreateOrganizationFundingRound)
			r.Delete("/{id}/funding-rounds/{roundId}", c.Organization.DeleteOrganizationFundingRound)
			r.Get("/{id}/signals", c.Organization.GetOrganizationSignals)
			r.Post("/{id}/signals", c.Organization.CreateOrganizationSignal)
			r.Delete("/{id}/signals/{signalId}", c.Organization.DeleteOrganizationSignal)
		})

		// Individuals routes
		r.Route("/individuals", func(r chi.Router) {
			r.Get("/", c.Individual.GetIndividuals)
			r.Post("/", c.Individual.CreateIndividual)
			r.Get("/search", c.Individual.SearchIndividuals)
			r.Get("/{id}", c.Individual.GetIndividual)
			r.Put("/{id}", c.Individual.UpdateIndividual)
			r.Delete("/{id}", c.Individual.DeleteIndividual)
			r.Get("/{id}/contacts", c.Individual.GetContacts)
			r.Post("/{id}/contacts", c.Individual.AddContact)
			r.Put("/{id}/contacts/{contactId}", c.Individual.UpdateContact)
			r.Delete("/{id}/contacts/{contactId}", c.Individual.DeleteContact)
			r.Get("/{id}/signals", c.Individual.GetSignals)
			r.Post("/{id}/signals", c.Individual.CreateSignal)
			r.Delete("/{id}/signals/{signalId}", c.Individual.DeleteSignal)
		})

		// Tasks routes
		r.Route("/tasks", func(r chi.Router) {
			r.Get("/", c.Task.GetTasks)
			r.Post("/", c.Task.CreateTask)
			r.Get("/statuses", c.Lookup.GetTaskStatuses)
			r.Get("/priorities", c.Lookup.GetTaskPriorities)
			r.Get("/entity/{entityType}/{entityID}", c.Task.GetTasksByEntity)
			r.Get("/{id}", c.Task.GetTask)
			r.Put("/{id}", c.Task.UpdateTask)
			r.Delete("/{id}", c.Task.DeleteTask)
		})

		// Contacts routes
		r.Route("/contacts", func(r chi.Router) {
			r.Get("/", c.Contact.GetContacts)
			r.Get("/search", c.Contact.SearchContacts)
			r.Post("/", c.Contact.CreateContact)
			r.Get("/{id}", c.Contact.GetContact)
			r.Put("/{id}", c.Contact.UpdateContact)
			r.Delete("/{id}", c.Contact.DeleteContact)
		})

		// Users routes
		r.Route("/users", func(r chi.Router) {
			r.Get("/", c.Auth.GetTenantUsers)
			r.Get("/{email}/tenant", c.Auth.GetUserTenant)
			r.Put("/me/selected-project", c.Auth.SetSelectedProject)
		})

		// Preferences routes
		r.Route("/preferences", func(r chi.Router) {
			r.Get("/user", c.Preferences.GetUserPreferences)
			r.Patch("/user", c.Preferences.UpdateUserPreferences)
			r.Get("/timezones", c.Preferences.GetTimezones)
			r.Get("/tenant", c.Preferences.GetTenantPreferences)
			r.Patch("/tenant", c.Preferences.UpdateTenantPreferences)
		})

		// Notifications routes
		r.Route("/notifications", func(r chi.Router) {
			r.Get("/", c.Notification.GetNotifications)
			r.Get("/count", c.Notification.GetNotificationCount)
			r.Post("/mark-read", c.Notification.MarkAsRead)
			r.Post("/mark-entity-read", c.Notification.MarkEntityAsRead)
		})

		// Projects routes
		r.Route("/projects", func(r chi.Router) {
			r.Get("/", c.Project.GetProjects)
			r.Post("/", c.Project.CreateProject)
			r.Get("/{id}", c.Project.GetProject)
			r.Put("/{id}", c.Project.UpdateProject)
			r.Delete("/{id}", c.Project.DeleteProject)
			r.Get("/{id}/leads", c.Project.GetProjectLeads)
			r.Get("/{id}/deals", c.Project.GetProjectDeals)
		})

		// Conversations routes
		r.Route("/conversations", func(r chi.Router) {
			r.Get("/", c.Conversation.GetConversations)
			r.Get("/all", c.Conversation.GetTenantConversations)
			r.Get("/{thread_id}", c.Conversation.GetConversation)
			r.Patch("/{thread_id}", c.Conversation.UpdateConversationTitle)
			r.Delete("/{thread_id}", c.Conversation.ArchiveConversation)
			r.Post("/{thread_id}/unarchive", c.Conversation.UnarchiveConversation)
			r.Delete("/{thread_id}/permanent", c.Conversation.DeleteConversationPermanent)
		})

		// Opportunities routes
		r.Route("/opportunities", func(r chi.Router) {
			r.Get("/", c.Lead.GetOpportunities)
			r.Get("/{id}", c.Lead.GetOpportunity)
		})

		// Partnerships routes
		r.Route("/partnerships", func(r chi.Router) {
			r.Get("/", c.Lead.GetPartnerships)
			r.Get("/{id}", c.Lead.GetPartnership)
		})

		// Accounts routes
		r.Route("/accounts", func(r chi.Router) {
			r.Get("/search", c.Account.SearchAccounts)
			r.Post("/{id}/relationships", c.Account.CreateRelationship)
			r.Delete("/{id}/relationships/{relationshipId}", c.Account.DeleteRelationship)
		})

		// Lookup routes
		r.Route("/industries", func(r chi.Router) {
			r.Get("/", c.Lookup.GetIndustries)
			r.Post("/", c.Lookup.CreateIndustry)
		})
		r.Get("/employee-count-ranges", c.Lookup.GetEmployeeCountRanges)
		r.Get("/revenue-ranges", c.Lookup.GetRevenueRanges)
		r.Get("/funding-stages", c.Lookup.GetFundingStages)
		r.Get("/salary-ranges", c.Lookup.GetSalaryRanges)

		// Costs routes
		r.Route("/costs", func(r chi.Router) {
			r.Get("/summary", c.Costs.GetCostsSummary)
			r.Get("/org", c.Costs.GetOrgCosts)
		})

		// Provenance routes
		r.Route("/provenance", func(r chi.Router) {
			r.Get("/{entityType}/{entityID}", c.Provenance.GetProvenance)
			r.Post("/", c.Provenance.CreateProvenance)
		})

		// Notes routes
		r.Route("/notes", func(r chi.Router) {
			r.Get("/", c.Note.GetNotes)
			r.Post("/", c.Note.CreateNote)
			r.Get("/{id}", c.Note.GetNote)
			r.Put("/{id}", c.Note.UpdateNote)
			r.Delete("/{id}", c.Note.DeleteNote)
		})

		// Comments routes
		r.Route("/comments", func(r chi.Router) {
			r.Get("/", c.Comment.GetComments)
			r.Post("/", c.Comment.CreateComment)
			r.Get("/{id}", c.Comment.GetComment)
			r.Put("/{id}", c.Comment.UpdateComment)
			r.Delete("/{id}", c.Comment.DeleteComment)
		})

		// Documents routes
		r.Route("/assets", func(r chi.Router) {
			r.Get("/documents", c.Document.GetDocuments)
			r.Get("/documents/check-filename", c.Document.CheckFilename)
			r.Get("/documents/{id}", c.Document.GetDocument)
			r.Get("/documents/{id}/download", c.Document.DownloadDocument)
			r.Post("/documents", c.Document.UploadDocument)
			r.Put("/documents/{id}", c.Document.UpdateDocument)
			r.Delete("/documents/{id}", c.Document.DeleteDocument)
			r.Post("/documents/{id}/link", c.Document.LinkDocument)
			r.Delete("/documents/{id}/link", c.Document.UnlinkDocument)
			r.Get("/documents/{id}/versions", c.Document.GetDocumentVersions)
			r.Post("/documents/{id}/versions", c.Document.CreateDocumentVersion)
			r.Patch("/documents/{id}/set-current", c.Document.SetCurrentVersion)
		})

		// Technologies routes
		r.Route("/technologies", func(r chi.Router) {
			r.Get("/", c.Technology.GetTechnologies)
			r.Get("/{id}", c.Technology.GetTechnology)
			r.Post("/", c.Technology.CreateTechnology)
		})

		// Usage routes
		r.Route("/usage", func(r chi.Router) {
			r.Get("/user", c.Usage.GetUserUsage)
			r.Get("/tenant", c.Usage.GetTenantUsage)
			r.Get("/timeseries", c.Usage.GetUsageTimeseries)
			r.Get("/analytics", c.Usage.GetDeveloperAnalytics)
			r.Get("/developer-access", c.Usage.CheckDeveloperAccess)
		})

		// Leads routes (job leads)
		r.Route("/leads", func(r chi.Router) {
			r.Get("/", c.Lead.GetLeads)
			r.Get("/statuses", c.Lead.GetStatuses)
			r.Post("/{job_id}/apply", c.Lead.MarkAsApplied)
			r.Post("/{job_id}/do-not-apply", c.Lead.MarkAsDoNotApply)
		})

		// Tags routes
		r.Get("/tags", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`))
		})

		// Reports routes
		r.Route("/reports", func(r chi.Router) {
			// Canned reports
			r.Get("/canned/lead-pipeline", c.Report.GetLeadPipeline)
			r.Get("/canned/account-overview", c.Report.GetAccountOverview)
			r.Get("/canned/contact-coverage", c.Report.GetContactCoverage)
			r.Get("/canned/notes-activity", c.Report.GetNotesActivity)
			r.Get("/canned/user-audit", c.Report.GetUserAudit)

			// Custom report execution
			r.Get("/fields/{entity}", c.Report.GetEntityFields)
			r.Post("/custom/preview", c.Report.PreviewCustomReport)
			r.Post("/custom/run", c.Report.RunCustomReport)
			r.Post("/custom/export", c.Report.ExportCustomReport)

			// Saved reports CRUD
			r.Get("/saved", c.Report.ListSavedReports)
			r.Post("/saved", c.Report.CreateSavedReport)
			r.Get("/saved/{report_id}", c.Report.GetSavedReport)
			r.Put("/saved/{report_id}", c.Report.UpdateSavedReport)
			r.Delete("/saved/{report_id}", c.Report.DeleteSavedReport)
		})

		// Agent routes (ADK-powered chat)
		r.Route("/agent", func(r chi.Router) {
			r.Post("/chat", c.Agent.Chat)
			r.Get("/health", c.Agent.HealthCheck)
		})
	})
}
