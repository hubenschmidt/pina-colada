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
	Preferences  *controllers.PreferencesController
	Notification *controllers.NotificationController
	Project      *controllers.ProjectController
	Conversation *controllers.ConversationController
	Lookup       *controllers.LookupController
	Note         *controllers.NoteController
	Comment      *controllers.CommentController
	Document     *controllers.DocumentController
	Account      *controllers.AccountController
	Lead         *controllers.LeadController
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
	// Protected routes (require auth)
	r.Group(func(r chi.Router) {
		r.Use(appMiddleware.AuthMiddleware)
		r.Use(appMiddleware.UserLoaderMiddleware(userLoader))

		// Auth routes
		r.Route("/auth", func(r chi.Router) {
			r.Get("/me", c.Auth.GetMe)
		})

		// Jobs routes
		r.Route("/jobs", func(r chi.Router) {
			r.Get("/", c.Job.GetJobs)
			r.Post("/", c.Job.CreateJob)
			r.Get("/{id}", c.Job.GetJob)
			r.Put("/{id}", c.Job.UpdateJob)
			r.Delete("/{id}", c.Job.DeleteJob)
		})

		// Organizations routes
		r.Route("/organizations", func(r chi.Router) {
			r.Get("/", c.Organization.GetOrganizations)
			r.Get("/search", c.Organization.SearchOrganizations)
			r.Get("/{id}", c.Organization.GetOrganization)
			r.Put("/{id}", c.Organization.UpdateOrganization)
			r.Delete("/{id}", c.Organization.DeleteOrganization)
			r.Post("/{id}/contacts", c.Organization.AddOrganizationContact)
		})

		// Individuals routes
		r.Route("/individuals", func(r chi.Router) {
			r.Get("/", c.Individual.GetIndividuals)
			r.Get("/search", c.Individual.SearchIndividuals)
			r.Get("/{id}", c.Individual.GetIndividual)
			r.Put("/{id}", c.Individual.UpdateIndividual)
			r.Delete("/{id}", c.Individual.DeleteIndividual)
			r.Post("/{id}/contacts", c.Individual.AddContact)
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
			r.Get("/{email}/tenant", c.Auth.GetUserTenant)
			r.Put("/me/selected-project", c.Auth.SetSelectedProject)
		})

		// Preferences routes
		r.Route("/preferences", func(r chi.Router) {
			r.Get("/user", c.Preferences.GetUserPreferences)
			r.Patch("/user", c.Preferences.UpdateUserPreferences)
			r.Get("/timezones", c.Preferences.GetTimezones)
		})

		// Notifications routes
		r.Route("/notifications", func(r chi.Router) {
			r.Get("/count", c.Notification.GetNotificationCount)
		})

		// Projects routes
		r.Route("/projects", func(r chi.Router) {
			r.Get("/", c.Project.GetProjects)
		})

		// Conversations routes
		r.Route("/conversations", func(r chi.Router) {
			r.Get("/", c.Conversation.GetConversations)
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
		r.Get("/industries", c.Lookup.GetIndustries)
		r.Get("/employee-count-ranges", c.Lookup.GetEmployeeCountRanges)
		r.Get("/revenue-ranges", c.Lookup.GetRevenueRanges)
		r.Get("/funding-stages", c.Lookup.GetFundingStages)
		r.Get("/salary-ranges", c.Lookup.GetSalaryRanges)

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
			r.Post("/documents", c.Document.UploadDocument)
			r.Post("/documents/{id}/link", c.Document.LinkDocument)
			r.Delete("/documents/{id}/link", c.Document.UnlinkDocument)
			r.Get("/documents/{id}/versions", c.Document.GetDocumentVersions)
			r.Post("/documents/{id}/versions", c.Document.CreateDocumentVersion)
			r.Patch("/documents/{id}/set-current", c.Document.SetCurrentVersion)
		})

		// Tags routes
		r.Get("/tags", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`))
		})
	})
}
