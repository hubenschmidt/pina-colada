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
			r.Delete("/{id}", c.Organization.DeleteOrganization)
		})

		// Individuals routes
		r.Route("/individuals", func(r chi.Router) {
			r.Get("/", c.Individual.GetIndividuals)
			r.Get("/search", c.Individual.SearchIndividuals)
			r.Get("/{id}", c.Individual.GetIndividual)
			r.Delete("/{id}", c.Individual.DeleteIndividual)
		})

		// Tasks routes
		r.Route("/tasks", func(r chi.Router) {
			r.Get("/", c.Task.GetTasks)
			r.Post("/", c.Task.CreateTask)
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
		})

		// Preferences routes
		r.Route("/preferences", func(r chi.Router) {
			r.Get("/user", c.Preferences.GetUserPreferences)
			r.Patch("/user", c.Preferences.UpdateUserPreferences)
		})
	})
}
