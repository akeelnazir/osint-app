// Package server wires the HTTP router, middleware, and handlers.
package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/akeelnazir/osint-app/backend/internal/auth"
	"github.com/akeelnazir/osint-app/backend/internal/config"
	"github.com/akeelnazir/osint-app/backend/internal/ent"
	"github.com/akeelnazir/osint-app/backend/internal/handler"
	appmw "github.com/akeelnazir/osint-app/backend/internal/middleware"
	"github.com/akeelnazir/osint-app/backend/internal/migrate"
	"github.com/akeelnazir/osint-app/backend/internal/storage"

	"entgo.io/ent/dialect/sql"
)

// Server bundles the HTTP server and its dependencies.
type Server struct {
	cfg     *config.Config
	client  *ent.Client
	drv     *sql.Driver
	authSvc *auth.Service
	storage storage.Storage
}

// New constructs the Server, opens the DB, runs migrations, and wires routes.
func New(cfg *config.Config) (*Server, error) {
	drv, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	client := ent.NewClient(ent.Driver(drv))

	// Run Ent auto-migration.
	if err := client.Schema.Create(context.Background()); err != nil {
		return nil, err
	}
	// Run PostGIS + FTS custom migrations.
	if err := migrate.Apply(context.Background(), drv); err != nil {
		log.Printf("postgis migration warning: %v", err)
	}

	authSvc := auth.New(cfg.JWTAccessSecret, cfg.JWTRefreshSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	store, err := storage.New(cfg)
	if err != nil {
		return nil, err
	}

	return &Server{cfg: cfg, client: client, drv: drv, authSvc: authSvc, storage: store}, nil
}

// Router builds the chi router with all routes mounted under /api/v1.
func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	// Global middleware.
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(appmw.Recover)
	r.Use(appmw.SecureHeaders)
	r.Use(appmw.RateLimit(s.cfg.RateLimitRPS))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   s.cfg.CORSAllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Content-Disposition"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		})

		// Auth (public).
		authH := handler.NewAuthHandler(s.client, s.authSvc)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authH.Register)
			r.Post("/login", authH.Login)
			r.Post("/refresh", authH.Refresh)
		})

		// Authenticated routes.
		r.Group(func(r chi.Router) {
			r.Use(appmw.RequireAuth(s.authSvc, s.client))

			userH := handler.NewUserHandler()
			r.Get("/users/me", userH.Me)

			caseH := handler.NewCaseHandler(s.client, s.drv)
			r.Route("/cases", func(r chi.Router) {
				r.Get("/", caseH.List)
				r.Post("/", caseH.Create)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", caseH.Get)
					r.Put("/", caseH.Update)
					r.Delete("/", caseH.Delete)
					r.Post("/members", caseH.AddMember)
					r.Delete("/members/{userId}", caseH.RemoveMember)
					r.Get("/members", caseH.ListMembers)

					// Evidence sub-resources.
					evH := handler.NewEvidenceHandler(s.client, s.drv, s.storage)
					r.Get("/evidence", evH.List)
					r.Post("/evidence", evH.Create)
					r.Get("/evidence/geojson", evH.GeoJSON)
					r.Get("/timeline", caseH.Timeline)
					r.Route("/evidence/{evidenceId}", func(r chi.Router) {
						r.Get("/", evH.Get)
						r.Put("/", evH.Update)
						r.Delete("/", evH.Delete)
						r.Post("/entities", evH.ExtractEntities)
						r.Get("/entities", evH.ListEntities)
						r.Get("/comments", evH.ListComments)
						r.Post("/comments", evH.CreateComment)
					})

					// Case comments.
					r.Get("/comments", caseH.ListComments)
					r.Post("/comments", caseH.CreateComment)

					// Activity.
					r.Get("/activity", caseH.ListActivity)
				})
			})

			searchH := handler.NewSearchHandler(s.client, s.drv)
			r.Get("/search", searchH.Search)

			exportH := handler.NewExportHandler(s.client)
			r.Get("/export/case/{id}/pdf", exportH.CasePDF)
			r.Get("/export/case/{id}/evidence.csv", exportH.EvidenceCSV)

			// Dashboard.
			dashH := handler.NewDashboardHandler(s.client)
			r.Get("/dashboard", dashH.Stats)
		})
	})

	return r
}

// Start runs the HTTP server until ctx is cancelled.
func (s *Server) Start(ctx context.Context) error {
	srv := &http.Server{
		Addr:              s.cfg.HTTPAddr,
		Handler:           s.Router(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		log.Printf("HTTP server listening on %s", s.cfg.HTTPAddr)
		errCh <- srv.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutCtx)
	case err := <-errCh:
		return err
	}
}

// Close releases DB resources.
func (s *Server) Close() error {
	return s.client.Close()
}
