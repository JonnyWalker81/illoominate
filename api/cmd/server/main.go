package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/fulldisclosure/api/internal/auth"
	"github.com/fulldisclosure/api/internal/config"
	"github.com/fulldisclosure/api/internal/handler"
	"github.com/fulldisclosure/api/internal/repository"
)

func main() {
	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration
	cfg, err := config.Load(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Setup logging
	setupLogging(cfg)

	log.Info().
		Str("env", cfg.Env).
		Int("port", cfg.Port).
		Msg("Starting FullDisclosure API")

	// Connect to database
	dbPool, err := setupDatabase(ctx, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer dbPool.Close()

	// Initialize auth validators
	supabaseValidator := auth.NewSupabaseValidator(cfg.SupabaseURL)

	// Initialize repositories
	portalRepo := repository.NewPortalRepository(dbPool)

	// Initialize handlers
	sdkTokenHandler := handler.NewSDKTokenHandler(dbPool)
	portalHandlers := handler.NewPortalHandlers(portalRepo, log.Logger)

	// TODO: Initialize repositories (data layer)
	// feedbackRepo := repository.NewFeedbackRepository(dbPool)
	// voteRepo := repository.NewVoteRepository(dbPool)
	// commentRepo := repository.NewCommentRepository(dbPool)
	// membershipRepo := repository.NewMembershipRepository(dbPool)
	// projectRepo := repository.NewProjectRepository(dbPool)
	// inviteRepo := repository.NewInviteRepository(dbPool)
	// tagRepo := repository.NewTagRepository(dbPool)
	// attachmentRepo := repository.NewAttachmentRepository(dbPool)

	// TODO: Initialize storage
	// gcsClient, err := storage.NewGCSClient(ctx, cfg.GCSBucket)
	// if err != nil {
	//     log.Fatal().Err(err).Msg("Failed to initialize GCS client")
	// }

	// TODO: Initialize services (business layer)
	// feedbackSvc := service.NewFeedbackService(feedbackRepo, voteRepo, tagRepo)
	// voteSvc := service.NewVoteService(voteRepo, feedbackRepo)
	// commentSvc := service.NewCommentService(commentRepo, feedbackRepo)
	// membershipSvc := service.NewMembershipService(membershipRepo, inviteRepo)
	// projectSvc := service.NewProjectService(projectRepo)
	// inviteSvc := service.NewInviteService(inviteRepo, membershipRepo)
	// tagSvc := service.NewTagService(tagRepo)
	// attachmentSvc := service.NewAttachmentService(attachmentRepo, gcsClient)

	// TODO: Initialize auth
	// jwtValidator := auth.NewSupabaseValidator(cfg.SupabaseJWTSecret)
	// sdkTokenValidator := auth.NewSDKTokenValidator(dbPool)

	// TODO: Initialize handlers (HTTP layer)
	// sdkHandler := handler.NewSDKHandler(feedbackSvc, attachmentSvc, sdkTokenValidator)
	// communityHandler := handler.NewCommunityHandler(feedbackSvc, voteSvc, commentSvc)
	// creatorHandler := handler.NewCreatorHandler(feedbackSvc, voteSvc, commentSvc, tagSvc, membershipSvc)
	// inviteHandler := handler.NewInviteHandler(inviteSvc)

	// Setup router
	r := setupRouter(cfg)

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"fulldisclosure-api"}`))
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		// SDK upload endpoint (no auth - uses attachment ID as auth token)
		// This is outside the SDK auth middleware because presigned URLs don't carry tokens
		r.Put("/sdk/attachments/{attachmentId}/upload", sdkUploadFileHandler(dbPool))

		// SDK routes (SDK token auth)
		r.Route("/sdk", func(r chi.Router) {
			r.Use(sdkAuthMiddleware(dbPool))
			r.Post("/identify", sdkIdentifyHandler(dbPool))
			r.Post("/feedback", sdkSubmitFeedbackHandler(dbPool))
			r.Post("/attachments/init", sdkInitiateUploadHandler(dbPool))
			r.Post("/attachments/complete", sdkCompleteUploadHandler(dbPool))
		})

		// Community routes (Supabase JWT + membership required)
		r.Route("/community", func(r chi.Router) {
			// TODO: Add auth middleware
			// r.Use(supabaseAuthMiddleware)
			// r.Use(requireMembershipMiddleware)
			r.Route("/projects/{projectId}", func(r chi.Router) {
				r.Get("/feature-requests", listFeatureRequestsHandler(dbPool))
				r.Post("/feature-requests", createFeatureRequestHandler(dbPool))
				r.Get("/feature-requests/{feedbackId}", placeholderHandler("Get feature request"))
				r.Delete("/feature-requests/{feedbackId}", deleteFeatureRequestHandler(dbPool))
				r.Post("/feature-requests/{feedbackId}/vote", voteFeatureRequestHandler(dbPool))
				r.Delete("/feature-requests/{feedbackId}/vote", unvoteFeatureRequestHandler(dbPool))
				r.Get("/feature-requests/{feedbackId}/comments", listCommentsHandler(dbPool))
				r.Post("/feature-requests/{feedbackId}/comments", createCommentHandler(dbPool))
			})
		})

		// Creator routes (Supabase JWT + team role required)
		r.Route("/creator", func(r chi.Router) {
			// TODO: Add auth middleware
			// r.Use(supabaseAuthMiddleware)
			// r.Use(requireTeamRoleMiddleware)

			// Project management
			r.Get("/projects", listProjectsHandler(dbPool))
			r.Post("/projects", createProjectHandler(dbPool))

			r.Route("/projects/{projectId}", func(r chi.Router) {
				// Project CRUD
				r.Get("/", getProjectHandler(dbPool))
				r.Delete("/", deleteProjectHandler(dbPool))

				// Feedback
				r.Get("/feedback", listCreatorFeedbackHandler(dbPool))
				r.Post("/feedback", placeholderHandler("Create feedback"))
				r.Get("/feedback/{feedbackId}", placeholderHandler("Get feedback"))
				r.Patch("/feedback/{feedbackId}", placeholderHandler("Update feedback"))
				r.Post("/feedback/{feedbackId}/merge", placeholderHandler("Merge feedback"))
				r.Post("/feedback/{feedbackId}/notes", placeholderHandler("Add team note"))

				// Tags
				r.Get("/tags", placeholderHandler("List tags"))
				r.Post("/tags", placeholderHandler("Create tag"))
				r.Patch("/tags/{tagId}", placeholderHandler("Update tag"))
				r.Delete("/tags/{tagId}", placeholderHandler("Delete tag"))

				// Members
				r.Get("/members", placeholderHandler("List members"))
				r.Post("/members", placeholderHandler("Add member"))
				r.Patch("/members/{memberId}", placeholderHandler("Update member"))
				r.Delete("/members/{memberId}", placeholderHandler("Remove member"))
				r.Post("/members/invite", placeholderHandler("Send invite"))

				// Settings
				r.Get("/settings", placeholderHandler("Get settings"))
				r.Patch("/settings", placeholderHandler("Update settings"))

				// SDK Tokens
				r.Get("/sdk-tokens", sdkTokenHandler.List)
				r.Post("/sdk-tokens", sdkTokenHandler.Create)
				r.Delete("/sdk-tokens/{tokenId}", sdkTokenHandler.Revoke)

				// Users (identified feedback submitters)
				r.Get("/users", listProjectUsersHandler(dbPool))
				r.Get("/users/{userId}/feedback", listUserFeedbackHandler(dbPool))
			})
		})

		// Portal routes (for feedback users)
		r.Route("/portal/{projectId}", func(r chi.Router) {
			// Public endpoint - list public features (optional auth for has_voted tracking)
			r.Get("/feature-requests", portalHandlers.ListFeatures)

			// Protected endpoints (Supabase JWT required)
			r.Group(func(r chi.Router) {
				r.Use(auth.SupabaseAuthMiddleware(supabaseValidator))
				r.Use(handler.PortalAccessMiddleware(portalRepo, log.Logger))

				r.Get("/me", portalHandlers.GetProfile)
				r.Patch("/me/notifications", portalHandlers.UpdateNotificationPreferences)
				r.Get("/my-feedback", portalHandlers.ListMyFeedback)
				r.Post("/feature-requests/{feedbackId}/vote", portalHandlers.Vote)
				r.Delete("/feature-requests/{feedbackId}/vote", portalHandlers.Unvote)
			})
		})

		// Invite acceptance (public with Supabase JWT)
		r.Post("/invites/{token}/accept", placeholderHandler("Accept invite"))

		// User info
		r.Get("/me", placeholderHandler("Get current user"))
	})

	// Create server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info().Msgf("Server listening on :%d", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server stopped")
}

func setupLogging(cfg *config.Config) {
	// Set log level
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Pretty printing for development
	if cfg.IsDevelopment() {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}

func setupDatabase(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database URL: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.DBMaxOpenConns)
	poolConfig.MinConns = int32(cfg.DBMaxIdleConns)
	poolConfig.MaxConnLifetime = cfg.DBConnMaxLifetime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	log.Info().Msg("Connected to database")
	return pool, nil
}

func setupRouter(cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()

	// Middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-SDK-Token", "X-Request-ID"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	return r
}

// placeholderHandler returns a handler that returns a placeholder response
// This will be replaced with actual handlers once the service layer is implemented
func placeholderHandler(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(fmt.Sprintf(`{"message":"Not implemented: %s","status":"placeholder"}`, name)))
	}
}

// listProjectsHandler handles GET /creator/projects
func listProjectsHandler(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := dbPool.Query(r.Context(), `
			SELECT id, name, slug, project_key, primary_color, created_at, updated_at
			FROM projects
			WHERE archived_at IS NULL
			ORDER BY created_at DESC
		`)
		if err != nil {
			log.Error().Err(err).Msg("Failed to list projects")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Failed to list projects"}`))
			return
		}
		defer rows.Close()

		var projects []struct {
			ID           string    `json:"id"`
			Name         string    `json:"name"`
			Slug         string    `json:"slug"`
			ProjectKey   string    `json:"project_key"`
			PrimaryColor string    `json:"primary_color"`
			CreatedAt    time.Time `json:"created_at"`
			UpdatedAt    time.Time `json:"updated_at"`
		}

		for rows.Next() {
			var p struct {
				ID           string    `json:"id"`
				Name         string    `json:"name"`
				Slug         string    `json:"slug"`
				ProjectKey   string    `json:"project_key"`
				PrimaryColor string    `json:"primary_color"`
				CreatedAt    time.Time `json:"created_at"`
				UpdatedAt    time.Time `json:"updated_at"`
			}
			if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.ProjectKey, &p.PrimaryColor, &p.CreatedAt, &p.UpdatedAt); err != nil {
				log.Error().Err(err).Msg("Failed to scan project")
				continue
			}
			projects = append(projects, p)
		}

		if projects == nil {
			projects = []struct {
				ID           string    `json:"id"`
				Name         string    `json:"name"`
				Slug         string    `json:"slug"`
				ProjectKey   string    `json:"project_key"`
				PrimaryColor string    `json:"primary_color"`
				CreatedAt    time.Time `json:"created_at"`
				UpdatedAt    time.Time `json:"updated_at"`
			}{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(projects)
	}
}

// createProjectHandler handles POST /creator/projects
func createProjectHandler(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name string `json:"name"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
			return
		}

		if req.Name == "" {
			http.Error(w, `{"error":"Project name is required"}`, http.StatusBadRequest)
			return
		}

		// Generate slug from name
		slug := strings.ToLower(strings.ReplaceAll(req.Name, " ", "-"))
		slug = strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
				return r
			}
			return -1
		}, slug)

		// Generate project key
		projectKey := fmt.Sprintf("pk_%s", generateRandomString(16))

		// Create project in database
		var project struct {
			ID           string    `json:"id"`
			Name         string    `json:"name"`
			Slug         string    `json:"slug"`
			ProjectKey   string    `json:"project_key"`
			PrimaryColor string    `json:"primary_color"`
			CreatedAt    time.Time `json:"created_at"`
			UpdatedAt    time.Time `json:"updated_at"`
		}

		err := dbPool.QueryRow(r.Context(), `
			INSERT INTO projects (name, slug, project_key, primary_color, settings)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, name, slug, project_key, primary_color, created_at, updated_at
		`, req.Name, slug, projectKey, "#6366f1", `{
			"default_visibility": {"bug": "TEAM_ONLY", "feature": "COMMUNITY", "general": "TEAM_ONLY"},
			"allow_anonymous_feedback": true,
			"require_email_for_anonymous": false,
			"voting_enabled": true,
			"community_comments_enabled": true,
			"auto_close_duplicates": true,
			"notification_preferences": {"new_feedback": true, "status_changes": true, "new_comments": true}
		}`).Scan(
			&project.ID, &project.Name, &project.Slug, &project.ProjectKey,
			&project.PrimaryColor, &project.CreatedAt, &project.UpdatedAt,
		)

		if err != nil {
			log.Error().Err(err).Msg("Failed to create project")
			// Check for unique constraint violation
			if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "duplicate key") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte(`{"error":"A project with this name already exists"}`))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Failed to create project"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(project)
	}
}

// listFeatureRequestsHandler handles GET /community/projects/:projectId/feature-requests
func listFeatureRequestsHandler(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectId := chi.URLParam(r, "projectId")
		userId := "00000000-0000-0000-0000-000000000001" // Hardcoded for now

		// Get query parameters
		sort := r.URL.Query().Get("sort")
		status := r.URL.Query().Get("status")

		// Build ORDER BY clause
		var orderBy string
		switch sort {
		case "newest":
			orderBy = "f.created_at DESC"
		case "oldest":
			orderBy = "f.created_at ASC"
		case "comments":
			orderBy = "f.comment_count DESC, f.created_at DESC"
		default: // "votes" or empty
			orderBy = "f.vote_count DESC, f.created_at DESC"
		}

		// Build status filter
		var statusFilter string
		var args []interface{}
		args = append(args, projectId, userId)

		if status != "" && status != "all" {
			statusFilter = " AND f.status = $3"
			args = append(args, status)
		}

		query := fmt.Sprintf(`
			SELECT f.id, f.project_id, f.author_id, f.title, f.description, f.type, f.status, f.visibility,
			       f.vote_count, f.comment_count, f.created_at, f.updated_at,
			       CASE WHEN v.user_id IS NOT NULL THEN true ELSE false END as has_voted,
			       CASE WHEN f.author_id = $2 OR m.role IN ('admin', 'owner') THEN true ELSE false END as can_delete
			FROM feedback f
			LEFT JOIN votes v ON f.id = v.feedback_id AND v.user_id = $2
			LEFT JOIN memberships m ON f.project_id = m.project_id AND m.user_id = $2
			WHERE f.project_id = $1
			  AND f.type = 'feature'
			  AND f.visibility = 'COMMUNITY'
			  AND f.canonical_id IS NULL
			  %s
			ORDER BY %s
		`, statusFilter, orderBy)

		rows, err := dbPool.Query(r.Context(), query, args...)
		if err != nil {
			log.Error().Err(err).Msg("Failed to list feature requests")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Failed to list feature requests"}`))
			return
		}
		defer rows.Close()

		var requests []struct {
			ID           string    `json:"id"`
			ProjectID    string    `json:"project_id"`
			AuthorID     *string   `json:"author_id,omitempty"`
			Title        string    `json:"title"`
			Description  string    `json:"description"`
			Type         string    `json:"type"`
			Status       string    `json:"status"`
			Visibility   string    `json:"visibility"`
			VoteCount    int       `json:"vote_count"`
			CommentCount int       `json:"comment_count"`
			CreatedAt    time.Time `json:"created_at"`
			UpdatedAt    time.Time `json:"updated_at"`
			HasVoted     bool      `json:"has_voted"`
			CanDelete    bool      `json:"can_delete"`
		}

		for rows.Next() {
			var req struct {
				ID           string    `json:"id"`
				ProjectID    string    `json:"project_id"`
				AuthorID     *string   `json:"author_id,omitempty"`
				Title        string    `json:"title"`
				Description  string    `json:"description"`
				Type         string    `json:"type"`
				Status       string    `json:"status"`
				Visibility   string    `json:"visibility"`
				VoteCount    int       `json:"vote_count"`
				CommentCount int       `json:"comment_count"`
				CreatedAt    time.Time `json:"created_at"`
				UpdatedAt    time.Time `json:"updated_at"`
				HasVoted     bool      `json:"has_voted"`
				CanDelete    bool      `json:"can_delete"`
			}
			if err := rows.Scan(&req.ID, &req.ProjectID, &req.AuthorID, &req.Title, &req.Description,
				&req.Type, &req.Status, &req.Visibility, &req.VoteCount,
				&req.CommentCount, &req.CreatedAt, &req.UpdatedAt, &req.HasVoted, &req.CanDelete); err != nil {
				log.Error().Err(err).Msg("Failed to scan feature request")
				continue
			}
			requests = append(requests, req)
		}

		if requests == nil {
			requests = []struct {
				ID           string    `json:"id"`
				ProjectID    string    `json:"project_id"`
				AuthorID     *string   `json:"author_id,omitempty"`
				Title        string    `json:"title"`
				Description  string    `json:"description"`
				Type         string    `json:"type"`
				Status       string    `json:"status"`
				Visibility   string    `json:"visibility"`
				VoteCount    int       `json:"vote_count"`
				CommentCount int       `json:"comment_count"`
				CreatedAt    time.Time `json:"created_at"`
				UpdatedAt    time.Time `json:"updated_at"`
				HasVoted     bool      `json:"has_voted"`
				CanDelete    bool      `json:"can_delete"`
			}{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(requests)
	}
}

// createFeatureRequestHandler handles POST /community/projects/:projectId/feature-requests
func createFeatureRequestHandler(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectId := chi.URLParam(r, "projectId")
		userId := "00000000-0000-0000-0000-000000000001" // Hardcoded for now

		var req struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
			return
		}

		if req.Title == "" {
			http.Error(w, `{"error":"Title is required"}`, http.StatusBadRequest)
			return
		}

		if req.Description == "" {
			http.Error(w, `{"error":"Description is required"}`, http.StatusBadRequest)
			return
		}

		var feedback struct {
			ID           string    `json:"id"`
			ProjectID    string    `json:"project_id"`
			Title        string    `json:"title"`
			Description  string    `json:"description"`
			Type         string    `json:"type"`
			Status       string    `json:"status"`
			Visibility   string    `json:"visibility"`
			VoteCount    int       `json:"vote_count"`
			CommentCount int       `json:"comment_count"`
			CreatedAt    time.Time `json:"created_at"`
			UpdatedAt    time.Time `json:"updated_at"`
		}

		err := dbPool.QueryRow(r.Context(), `
			INSERT INTO feedback (project_id, author_id, title, description, type, status, visibility, source, submitter_identifier)
			VALUES ($1, $2, $3, $4, 'feature', 'new', 'COMMUNITY', 'web', 'anonymous')
			RETURNING id, project_id, title, description, type, status, visibility,
			          vote_count, comment_count, created_at, updated_at
		`, projectId, userId, req.Title, req.Description).Scan(
			&feedback.ID, &feedback.ProjectID, &feedback.Title, &feedback.Description,
			&feedback.Type, &feedback.Status, &feedback.Visibility, &feedback.VoteCount,
			&feedback.CommentCount, &feedback.CreatedAt, &feedback.UpdatedAt,
		)

		if err != nil {
			log.Error().Err(err).Msg("Failed to create feature request")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Failed to create feature request"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(feedback)
	}
}

// voteFeatureRequestHandler handles POST /community/projects/:projectId/feature-requests/:feedbackId/vote
func voteFeatureRequestHandler(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		feedbackId := chi.URLParam(r, "feedbackId")

		// For now, use a random user ID since auth isn't implemented
		// In production, this would come from the JWT
		userId := "00000000-0000-0000-0000-000000000001"

		// Try to insert vote (will fail if already voted due to unique constraint)
		_, err := dbPool.Exec(r.Context(), `
			INSERT INTO votes (feedback_id, user_id)
			VALUES ($1, $2)
			ON CONFLICT (feedback_id, user_id) DO NOTHING
		`, feedbackId, userId)

		if err != nil {
			log.Error().Err(err).Msg("Failed to vote")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Failed to vote"}`))
			return
		}

		// Get updated vote count
		var voteCount int
		err = dbPool.QueryRow(r.Context(), `
			SELECT vote_count FROM feedback WHERE id = $1
		`, feedbackId).Scan(&voteCount)

		if err != nil {
			log.Error().Err(err).Msg("Failed to get vote count")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"feedback_id": feedbackId,
			"vote_count":  voteCount,
			"has_voted":   true,
		})
	}
}

// unvoteFeatureRequestHandler handles DELETE /community/projects/:projectId/feature-requests/:feedbackId/vote
func unvoteFeatureRequestHandler(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		feedbackId := chi.URLParam(r, "feedbackId")

		// For now, use a random user ID since auth isn't implemented
		userId := "00000000-0000-0000-0000-000000000001"

		_, err := dbPool.Exec(r.Context(), `
			DELETE FROM votes WHERE feedback_id = $1 AND user_id = $2
		`, feedbackId, userId)

		if err != nil {
			log.Error().Err(err).Msg("Failed to unvote")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Failed to unvote"}`))
			return
		}

		// Get updated vote count
		var voteCount int
		err = dbPool.QueryRow(r.Context(), `
			SELECT vote_count FROM feedback WHERE id = $1
		`, feedbackId).Scan(&voteCount)

		if err != nil {
			log.Error().Err(err).Msg("Failed to get vote count")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"feedback_id": feedbackId,
			"vote_count":  voteCount,
			"has_voted":   false,
		})
	}
}

// getProjectHandler handles GET /creator/projects/:projectId
func getProjectHandler(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectId := chi.URLParam(r, "projectId")

		var project struct {
			ID           string    `json:"id"`
			Name         string    `json:"name"`
			Slug         string    `json:"slug"`
			ProjectKey   string    `json:"project_key"`
			PrimaryColor string    `json:"primary_color"`
			CreatedAt    time.Time `json:"created_at"`
			UpdatedAt    time.Time `json:"updated_at"`
		}

		err := dbPool.QueryRow(r.Context(), `
			SELECT id, name, slug, project_key, primary_color, created_at, updated_at
			FROM projects
			WHERE id = $1 AND archived_at IS NULL
		`, projectId).Scan(
			&project.ID, &project.Name, &project.Slug, &project.ProjectKey,
			&project.PrimaryColor, &project.CreatedAt, &project.UpdatedAt,
		)

		if err != nil {
			log.Error().Err(err).Str("projectId", projectId).Msg("Failed to get project")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"Project not found"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(project)
	}
}

// deleteProjectHandler handles DELETE /creator/projects/:projectId
func deleteProjectHandler(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectId := chi.URLParam(r, "projectId")

		result, err := dbPool.Exec(r.Context(), `
			DELETE FROM projects WHERE id = $1
		`, projectId)

		if err != nil {
			log.Error().Err(err).Str("projectId", projectId).Msg("Failed to delete project")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Failed to delete project"}`))
			return
		}

		if result.RowsAffected() == 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"Project not found"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Project deleted"}`))
	}
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.IntN(len(charset))]
	}
	return string(b)
}

// listCommentsHandler handles GET /community/projects/:projectId/feature-requests/:feedbackId/comments
func listCommentsHandler(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		feedbackId := chi.URLParam(r, "feedbackId")

		rows, err := dbPool.Query(r.Context(), `
			SELECT id, feedback_id, author_id, body, is_edited, created_at, updated_at
			FROM comments
			WHERE feedback_id = $1
			  AND visibility = 'COMMUNITY'
			  AND deleted_at IS NULL
			ORDER BY created_at ASC
		`, feedbackId)
		if err != nil {
			log.Error().Err(err).Msg("Failed to list comments")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Failed to list comments"}`))
			return
		}
		defer rows.Close()

		var comments []struct {
			ID        string    `json:"id"`
			FeedbackID string   `json:"feedback_id"`
			AuthorID  string    `json:"author_id"`
			Body      string    `json:"body"`
			IsEdited  bool      `json:"is_edited"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
		}

		for rows.Next() {
			var c struct {
				ID        string    `json:"id"`
				FeedbackID string   `json:"feedback_id"`
				AuthorID  string    `json:"author_id"`
				Body      string    `json:"body"`
				IsEdited  bool      `json:"is_edited"`
				CreatedAt time.Time `json:"created_at"`
				UpdatedAt time.Time `json:"updated_at"`
			}
			if err := rows.Scan(&c.ID, &c.FeedbackID, &c.AuthorID, &c.Body, &c.IsEdited, &c.CreatedAt, &c.UpdatedAt); err != nil {
				log.Error().Err(err).Msg("Failed to scan comment")
				continue
			}
			comments = append(comments, c)
		}

		if comments == nil {
			comments = []struct {
				ID        string    `json:"id"`
				FeedbackID string   `json:"feedback_id"`
				AuthorID  string    `json:"author_id"`
				Body      string    `json:"body"`
				IsEdited  bool      `json:"is_edited"`
				CreatedAt time.Time `json:"created_at"`
				UpdatedAt time.Time `json:"updated_at"`
			}{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(comments)
	}
}

// createCommentHandler handles POST /community/projects/:projectId/feature-requests/:feedbackId/comments
func createCommentHandler(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		feedbackId := chi.URLParam(r, "feedbackId")
		userId := "00000000-0000-0000-0000-000000000001" // Hardcoded for now

		var input struct {
			Body string `json:"body"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"Invalid request body"}`))
			return
		}

		if input.Body == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"Comment body is required"}`))
			return
		}

		var comment struct {
			ID        string    `json:"id"`
			FeedbackID string   `json:"feedback_id"`
			AuthorID  string    `json:"author_id"`
			Body      string    `json:"body"`
			IsEdited  bool      `json:"is_edited"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
		}

		err := dbPool.QueryRow(r.Context(), `
			INSERT INTO comments (feedback_id, author_id, body, visibility)
			VALUES ($1, $2, $3, 'COMMUNITY')
			RETURNING id, feedback_id, author_id, body, is_edited, created_at, updated_at
		`, feedbackId, userId, input.Body).Scan(
			&comment.ID, &comment.FeedbackID, &comment.AuthorID,
			&comment.Body, &comment.IsEdited, &comment.CreatedAt, &comment.UpdatedAt,
		)

		if err != nil {
			log.Error().Err(err).Msg("Failed to create comment")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Failed to create comment"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(comment)
	}
}

// deleteFeatureRequestHandler handles DELETE /community/projects/:projectId/feature-requests/:feedbackId
func deleteFeatureRequestHandler(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectId := chi.URLParam(r, "projectId")
		feedbackId := chi.URLParam(r, "feedbackId")
		userId := "00000000-0000-0000-0000-000000000001" // Hardcoded for now

		// Check if user can delete (author or admin/owner)
		var canDelete bool
		err := dbPool.QueryRow(r.Context(), `
			SELECT
				CASE WHEN f.author_id = $3 OR m.role IN ('admin', 'owner') THEN true ELSE false END
			FROM feedback f
			LEFT JOIN memberships m ON f.project_id = m.project_id AND m.user_id = $3
			WHERE f.id = $2 AND f.project_id = $1
		`, projectId, feedbackId, userId).Scan(&canDelete)

		if err != nil {
			log.Error().Err(err).Msg("Failed to check delete permission")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"Feature request not found"}`))
			return
		}

		if !canDelete {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error":"You don't have permission to delete this feature request"}`))
			return
		}

		// Delete the feature request
		result, err := dbPool.Exec(r.Context(), `
			DELETE FROM feedback WHERE id = $1 AND project_id = $2
		`, feedbackId, projectId)

		if err != nil {
			log.Error().Err(err).Msg("Failed to delete feature request")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Failed to delete feature request"}`))
			return
		}

		if result.RowsAffected() == 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"Feature request not found"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Feature request deleted"}`))
	}
}

// sdkAuthMiddleware validates SDK tokens
func sdkAuthMiddleware(dbPool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("X-SDK-Token")
			if token == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"SDK token required","code":"UNAUTHORIZED"}`))
				return
			}

			// Hash the token to compare with stored hash
			tokenHash := hashToken(token)

			// Look up token in database
			var projectID string
			var isActive bool
			err := dbPool.QueryRow(r.Context(), `
				SELECT project_id, is_active
				FROM sdk_tokens
				WHERE token_hash = $1
			`, tokenHash).Scan(&projectID, &isActive)

			if err != nil {
				log.Debug().Err(err).Msg("SDK token lookup failed")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"Invalid SDK token","code":"INVALID_TOKEN"}`))
				return
			}

			if !isActive {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"SDK token is inactive","code":"TOKEN_INACTIVE"}`))
				return
			}

			// Update last_used_at
			_, _ = dbPool.Exec(r.Context(), `
				UPDATE sdk_tokens SET last_used_at = NOW() WHERE token_hash = $1
			`, tokenHash)

			// Store project ID in context
			ctx := context.WithValue(r.Context(), sdkProjectIDKey, projectID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type contextKey string

const sdkProjectIDKey contextKey = "sdkProjectID"

func hashToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}

// SDKUser represents an identified user from the SDK
type SDKUser struct {
	ID         string                 `json:"id"`
	ExternalID string                 `json:"external_id"`
	Email      *string                `json:"email,omitempty"`
	Name       *string                `json:"name,omitempty"`
	AvatarURL  *string                `json:"avatar_url,omitempty"`
	Traits     map[string]interface{} `json:"traits,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	LastSeenAt time.Time              `json:"last_seen_at"`
}

// findOrCreateSDKUser finds an existing SDK user or creates a new one
// This implements Canny-style auto-creation: if user doesn't exist, create them
func findOrCreateSDKUser(ctx context.Context, dbPool *pgxpool.Pool, projectID, externalID string, email, name *string, traits map[string]interface{}) (*SDKUser, error) {
	// Convert traits to JSON
	var traitsJSON []byte
	var err error
	if traits != nil {
		traitsJSON, err = json.Marshal(traits)
		if err != nil {
			traitsJSON = []byte("{}")
		}
	} else {
		traitsJSON = []byte("{}")
	}

	// Try to insert, on conflict update the user info and return
	var user SDKUser
	var traitsResult []byte
	err = dbPool.QueryRow(ctx, `
		INSERT INTO sdk_users (project_id, external_id, email, name, traits, last_seen_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (project_id, external_id) DO UPDATE SET
			email = COALESCE(EXCLUDED.email, sdk_users.email),
			name = COALESCE(EXCLUDED.name, sdk_users.name),
			traits = CASE
				WHEN EXCLUDED.traits != '{}'::jsonb THEN sdk_users.traits || EXCLUDED.traits
				ELSE sdk_users.traits
			END,
			last_seen_at = NOW(),
			updated_at = NOW()
		RETURNING id, external_id, email, name, avatar_url, traits, created_at, updated_at, last_seen_at
	`, projectID, externalID, email, name, traitsJSON).Scan(
		&user.ID, &user.ExternalID, &user.Email, &user.Name, &user.AvatarURL,
		&traitsResult, &user.CreatedAt, &user.UpdatedAt, &user.LastSeenAt,
	)

	if err != nil {
		return nil, err
	}

	if traitsResult != nil {
		json.Unmarshal(traitsResult, &user.Traits)
	}

	return &user, nil
}

// listCreatorFeedbackHandler handles GET /creator/projects/{projectId}/feedback
func listCreatorFeedbackHandler(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "projectId")
		userFilter := r.URL.Query().Get("user")

		query := `
			SELECT
				f.id, f.title, f.description, f.type, f.status,
				f.vote_count, f.comment_count, f.source,
				f.submitter_email, f.submitter_name, f.submitter_identifier,
				f.source_metadata, f.created_at, f.updated_at,
				su.id as sdk_user_id, su.external_id, su.email as sdk_user_email,
				su.name as sdk_user_name, su.traits as sdk_user_traits
			FROM feedback f
			LEFT JOIN sdk_users su ON f.sdk_user_id = su.id
			WHERE f.project_id = $1 AND f.canonical_id IS NULL
		`
		args := []interface{}{projectID}

		if userFilter != "" {
			query += " AND (su.external_id = $2 OR f.submitter_identifier = $2)"
			args = append(args, userFilter)
		}

		query += " ORDER BY f.created_at DESC LIMIT 100"

		rows, err := dbPool.Query(r.Context(), query, args...)

		if err != nil {
			log.Error().Err(err).Msg("Failed to list feedback")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Failed to list feedback","code":"INTERNAL_ERROR"}`))
			return
		}
		defer rows.Close()

		type SDKUserInfo struct {
			ID         string                 `json:"id"`
			ExternalID string                 `json:"external_id"`
			Email      *string                `json:"email,omitempty"`
			Name       *string                `json:"name,omitempty"`
			Traits     map[string]interface{} `json:"traits,omitempty"`
		}

		type FeedbackItem struct {
			ID                  string                 `json:"id"`
			Title               string                 `json:"title"`
			Description         string                 `json:"description"`
			Type                string                 `json:"type"`
			Status              string                 `json:"status"`
			VoteCount           int                    `json:"vote_count"`
			CommentCount        int                    `json:"comment_count"`
			Source              string                 `json:"source"`
			SubmitterEmail      *string                `json:"submitter_email,omitempty"`
			SubmitterName       *string                `json:"submitter_name,omitempty"`
			SubmitterIdentifier *string                `json:"submitter_identifier,omitempty"`
			SourceMetadata      map[string]interface{} `json:"source_metadata,omitempty"`
			SDKUser             *SDKUserInfo           `json:"sdk_user,omitempty"`
			CreatedAt           string                 `json:"created_at"`
			UpdatedAt           string                 `json:"updated_at"`
		}

		var items []FeedbackItem
		for rows.Next() {
			var item FeedbackItem
			var createdAt, updatedAt time.Time
			var submitterEmail, submitterName, submitterIdentifier *string
			var sourceMetadataJSON []byte
			var sdkUserID, sdkExternalID, sdkUserEmail, sdkUserName *string
			var sdkUserTraitsJSON []byte

			err := rows.Scan(
				&item.ID, &item.Title, &item.Description,
				&item.Type, &item.Status, &item.VoteCount,
				&item.CommentCount, &item.Source,
				&submitterEmail, &submitterName, &submitterIdentifier,
				&sourceMetadataJSON, &createdAt, &updatedAt,
				&sdkUserID, &sdkExternalID, &sdkUserEmail, &sdkUserName, &sdkUserTraitsJSON,
			)
			if err != nil {
				log.Error().Err(err).Msg("Failed to scan feedback row")
				continue
			}

			item.SubmitterEmail = submitterEmail
			item.SubmitterName = submitterName
			item.SubmitterIdentifier = submitterIdentifier
			if sourceMetadataJSON != nil {
				json.Unmarshal(sourceMetadataJSON, &item.SourceMetadata)
			}

			// Add SDK user info if present
			if sdkUserID != nil && sdkExternalID != nil {
				item.SDKUser = &SDKUserInfo{
					ID:         *sdkUserID,
					ExternalID: *sdkExternalID,
					Email:      sdkUserEmail,
					Name:       sdkUserName,
				}
				if sdkUserTraitsJSON != nil {
					json.Unmarshal(sdkUserTraitsJSON, &item.SDKUser.Traits)
				}
			}

			item.CreatedAt = createdAt.Format(time.RFC3339)
			item.UpdatedAt = updatedAt.Format(time.RFC3339)
			items = append(items, item)
		}

		if items == nil {
			items = []FeedbackItem{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(items)
	}
}

// listProjectUsersHandler handles GET /creator/projects/{projectId}/users
func listProjectUsersHandler(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "projectId")
		search := r.URL.Query().Get("search")

		// Query SDK users directly with feedback count
		query := `
			SELECT
				su.id,
				su.external_id,
				su.email,
				su.name,
				su.traits,
				su.created_at,
				su.last_seen_at,
				COALESCE((
					SELECT COUNT(*)
					FROM feedback f
					WHERE f.sdk_user_id = su.id AND f.canonical_id IS NULL
				), 0) as feedback_count
			FROM sdk_users su
			WHERE su.project_id = $1
		`
		args := []interface{}{projectID}

		if search != "" {
			query += " AND (su.external_id ILIKE $2 OR su.email ILIKE $2 OR su.name ILIKE $2)"
			args = append(args, "%"+search+"%")
		}

		query += `
			ORDER BY su.last_seen_at DESC
			LIMIT 100
		`

		rows, err := dbPool.Query(r.Context(), query, args...)
		if err != nil {
			log.Error().Err(err).Msg("Failed to list project users")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Failed to list users","code":"INTERNAL_ERROR"}`))
			return
		}
		defer rows.Close()

		type IdentifiedUser struct {
			ID            string                 `json:"id"`
			Identifier    string                 `json:"identifier"`
			Email         *string                `json:"email,omitempty"`
			Name          *string                `json:"name,omitempty"`
			Traits        map[string]interface{} `json:"traits,omitempty"`
			FeedbackCount int                    `json:"feedback_count"`
			FirstSeen     string                 `json:"first_seen"`
			LastSeen      string                 `json:"last_seen"`
		}

		var users []IdentifiedUser
		for rows.Next() {
			var user IdentifiedUser
			var email, name *string
			var createdAt, lastSeenAt time.Time
			var traitsJSON []byte

			err := rows.Scan(
				&user.ID, &user.Identifier, &email, &name,
				&traitsJSON, &createdAt, &lastSeenAt, &user.FeedbackCount,
			)
			if err != nil {
				log.Error().Err(err).Msg("Failed to scan user row")
				continue
			}

			user.Email = email
			user.Name = name
			user.FirstSeen = createdAt.Format(time.RFC3339)
			user.LastSeen = lastSeenAt.Format(time.RFC3339)
			if traitsJSON != nil {
				json.Unmarshal(traitsJSON, &user.Traits)
			}
			users = append(users, user)
		}

		if users == nil {
			users = []IdentifiedUser{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}
}

// listUserFeedbackHandler handles GET /creator/projects/{projectId}/users/{userId}/feedback
func listUserFeedbackHandler(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "projectId")
		userID := chi.URLParam(r, "userId") // This is the external_id

		// Query feedback via SDK user join using external_id
		rows, err := dbPool.Query(r.Context(), `
			SELECT
				f.id, f.title, f.description, f.type, f.status,
				f.vote_count, f.comment_count, f.source,
				f.submitter_email, f.submitter_name, f.submitter_identifier,
				f.source_metadata, f.created_at, f.updated_at
			FROM feedback f
			INNER JOIN sdk_users su ON f.sdk_user_id = su.id
			WHERE f.project_id = $1
				AND su.external_id = $2
				AND f.canonical_id IS NULL
			ORDER BY f.created_at DESC
			LIMIT 100
		`, projectID, userID)

		if err != nil {
			log.Error().Err(err).Msg("Failed to list user feedback")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Failed to list user feedback","code":"INTERNAL_ERROR"}`))
			return
		}
		defer rows.Close()

		type FeedbackItem struct {
			ID                  string                 `json:"id"`
			Title               string                 `json:"title"`
			Description         string                 `json:"description"`
			Type                string                 `json:"type"`
			Status              string                 `json:"status"`
			VoteCount           int                    `json:"vote_count"`
			CommentCount        int                    `json:"comment_count"`
			Source              string                 `json:"source"`
			SubmitterEmail      *string                `json:"submitter_email,omitempty"`
			SubmitterName       *string                `json:"submitter_name,omitempty"`
			SubmitterIdentifier *string                `json:"submitter_identifier,omitempty"`
			SourceMetadata      map[string]interface{} `json:"source_metadata,omitempty"`
			CreatedAt           string                 `json:"created_at"`
			UpdatedAt           string                 `json:"updated_at"`
		}

		var items []FeedbackItem
		for rows.Next() {
			var item FeedbackItem
			var createdAt, updatedAt time.Time
			var submitterEmail, submitterName, submitterIdentifier *string
			var sourceMetadataJSON []byte

			err := rows.Scan(
				&item.ID, &item.Title, &item.Description,
				&item.Type, &item.Status, &item.VoteCount,
				&item.CommentCount, &item.Source,
				&submitterEmail, &submitterName, &submitterIdentifier,
				&sourceMetadataJSON, &createdAt, &updatedAt,
			)
			if err != nil {
				log.Error().Err(err).Msg("Failed to scan feedback row")
				continue
			}

			item.SubmitterEmail = submitterEmail
			item.SubmitterName = submitterName
			item.SubmitterIdentifier = submitterIdentifier
			if sourceMetadataJSON != nil {
				json.Unmarshal(sourceMetadataJSON, &item.SourceMetadata)
			}
			item.CreatedAt = createdAt.Format(time.RFC3339)
			item.UpdatedAt = updatedAt.Format(time.RFC3339)
			items = append(items, item)
		}

		if items == nil {
			items = []FeedbackItem{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(items)
	}
}

// sdkSubmitFeedbackHandler handles POST /sdk/feedback
func sdkSubmitFeedbackHandler(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get project ID from context (set by auth middleware)
		projectID, ok := r.Context().Value(sdkProjectIDKey).(string)
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"SDK authentication required","code":"UNAUTHORIZED"}`))
			return
		}

		var req struct {
			Title               string                 `json:"title"`
			Description         string                 `json:"description"`
			Type                string                 `json:"type"`
			SubmitterEmail      *string                `json:"submitter_email"`
			SubmitterName       *string                `json:"submitter_name"`
			SubmitterIdentifier *string                `json:"submitter_identifier"`
			SourceMetadata      map[string]interface{} `json:"source_metadata"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"Invalid JSON","code":"INVALID_JSON"}`))
			return
		}

		// Validation
		if req.Title == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"Title is required","code":"VALIDATION_ERROR"}`))
			return
		}
		if req.Description == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"Description is required","code":"VALIDATION_ERROR"}`))
			return
		}

		// Default type
		feedbackType := req.Type
		if feedbackType == "" {
			feedbackType = "general"
		}

		// Determine source from header
		source := "sdk"
		if sdkSource := r.Header.Get("X-SDK-Source"); sdkSource != "" {
			source = "sdk-" + sdkSource
		}

		// Convert metadata to JSON
		var metadataJSON []byte
		var err error
		if req.SourceMetadata != nil {
			metadataJSON, err = json.Marshal(req.SourceMetadata)
			if err != nil {
				metadataJSON = nil
			}
		}

		// Extract user traits from source_metadata if present
		var userTraits map[string]interface{}
		if req.SourceMetadata != nil {
			if traits, ok := req.SourceMetadata["user_traits"].(map[string]interface{}); ok {
				userTraits = traits
			}
		}

		// Ensure we have at least one identifier for the constraint
		// If no identifier and no email, use "anonymous" as fallback
		submitterIdentifier := req.SubmitterIdentifier
		if (submitterIdentifier == nil || *submitterIdentifier == "") && (req.SubmitterEmail == nil || *req.SubmitterEmail == "") {
			anon := "anonymous"
			submitterIdentifier = &anon
		}

		// Find or create SDK user if identifier is provided
		var sdkUserID *string
		if submitterIdentifier != nil && *submitterIdentifier != "" && *submitterIdentifier != "anonymous" {
			sdkUser, err := findOrCreateSDKUser(
				r.Context(),
				dbPool,
				projectID,
				*submitterIdentifier,
				req.SubmitterEmail,
				req.SubmitterName,
				userTraits,
			)
			if err != nil {
				log.Error().Err(err).Str("external_id", *submitterIdentifier).Msg("Failed to find/create SDK user")
				// Continue without SDK user - don't fail the feedback submission
			} else {
				sdkUserID = &sdkUser.ID
				log.Info().
					Str("sdk_user_id", sdkUser.ID).
					Str("external_id", sdkUser.ExternalID).
					Msg("SDK user found/created")
			}
		}

		// Insert feedback
		var feedbackID string
		var createdAt time.Time
		err = dbPool.QueryRow(r.Context(), `
			INSERT INTO feedback (
				project_id, title, description, type, status, visibility,
				source, submitter_email, submitter_name, source_metadata, submitter_identifier, sdk_user_id
			)
			VALUES ($1, $2, $3, $4, 'new', 'TEAM_ONLY', $5, $6, $7, $8, $9, $10)
			RETURNING id, created_at
		`, projectID, req.Title, req.Description, feedbackType, source,
			req.SubmitterEmail, req.SubmitterName, metadataJSON, submitterIdentifier, sdkUserID).Scan(&feedbackID, &createdAt)

		if err != nil {
			log.Error().Err(err).Msg("Failed to create feedback from SDK")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Failed to create feedback","code":"INTERNAL_ERROR"}`))
			return
		}

		log.Info().
			Str("feedback_id", feedbackID).
			Str("project_id", projectID).
			Str("source", source).
			Msg("Feedback submitted via SDK")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         feedbackID,
			"created_at": createdAt,
		})
	}
}

// sdkIdentifyHandler handles POST /sdk/identify
func sdkIdentifyHandler(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get project ID from context (set by auth middleware)
		projectID, ok := r.Context().Value(sdkProjectIDKey).(string)
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"SDK authentication required","code":"UNAUTHORIZED"}`))
			return
		}

		var req struct {
			UserID string            `json:"user_id"`
			Email  *string           `json:"email"`
			Name   *string           `json:"name"`
			Traits map[string]string `json:"traits"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"Invalid JSON","code":"INVALID_JSON"}`))
			return
		}

		// Validation
		if req.UserID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"user_id is required","code":"VALIDATION_ERROR"}`))
			return
		}

		// Convert traits to map[string]interface{}
		var traits map[string]interface{}
		if req.Traits != nil {
			traits = make(map[string]interface{})
			for k, v := range req.Traits {
				traits[k] = v
			}
		}

		// Find or create SDK user
		sdkUser, err := findOrCreateSDKUser(
			r.Context(),
			dbPool,
			projectID,
			req.UserID,
			req.Email,
			req.Name,
			traits,
		)
		if err != nil {
			log.Error().Err(err).Str("user_id", req.UserID).Msg("Failed to identify SDK user")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Failed to identify user","code":"INTERNAL_ERROR"}`))
			return
		}

		log.Info().
			Str("sdk_user_id", sdkUser.ID).
			Str("external_id", sdkUser.ExternalID).
			Msg("SDK user identified")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":          sdkUser.ID,
			"external_id": sdkUser.ExternalID,
			"email":       sdkUser.Email,
			"name":        sdkUser.Name,
			"traits":      sdkUser.Traits,
		})
	}
}

// sdkInitiateUploadHandler handles POST /sdk/attachments/init
func sdkInitiateUploadHandler(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			FeedbackID  string `json:"feedback_id"`
			Filename    string `json:"filename"`
			ContentType string `json:"content_type"`
			SizeBytes   int64  `json:"size_bytes"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"Invalid JSON","code":"INVALID_JSON"}`))
			return
		}

		// Validation
		if req.FeedbackID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"feedback_id is required","code":"VALIDATION_ERROR"}`))
			return
		}
		if req.Filename == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"filename is required","code":"VALIDATION_ERROR"}`))
			return
		}
		if req.ContentType == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"content_type is required","code":"VALIDATION_ERROR"}`))
			return
		}

		// Generate storage path for local development
		storagePath := fmt.Sprintf("uploads/%s/%s", req.FeedbackID, req.Filename)

		// Create attachment record
		var id string
		err := dbPool.QueryRow(r.Context(), `
			INSERT INTO attachments (feedback_id, filename, content_type, size_bytes, status, gcs_bucket, gcs_path)
			VALUES ($1, $2, $3, $4, 'pending', 'local', $5)
			RETURNING id
		`, req.FeedbackID, req.Filename, req.ContentType, req.SizeBytes, storagePath).Scan(&id)

		if err != nil {
			log.Error().Err(err).Msg("Failed to create attachment record")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Failed to initiate upload","code":"INTERNAL_ERROR"}`))
			return
		}

		// For local development, we'll use a simple upload endpoint
		// In production, this would be a presigned cloud storage URL
		uploadURL := fmt.Sprintf("http://localhost:8080/api/sdk/attachments/%s/upload", id)
		expiresAt := time.Now().Add(1 * time.Hour).Format(time.RFC3339)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"attachment_id": id,
			"upload_url":    uploadURL,
			"expires_at":    expiresAt,
		})
	}
}

// sdkCompleteUploadHandler handles POST /sdk/attachments/complete
func sdkCompleteUploadHandler(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			AttachmentID string `json:"attachment_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"Invalid JSON","code":"INVALID_JSON"}`))
			return
		}

		if req.AttachmentID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"attachment_id is required","code":"VALIDATION_ERROR"}`))
			return
		}

		// Update attachment status to completed
		var attachment struct {
			ID          string `json:"id"`
			Filename    string `json:"filename"`
			ContentType string `json:"content_type"`
			SizeBytes   int64  `json:"size_bytes"`
			Status      string `json:"status"`
		}

		err := dbPool.QueryRow(r.Context(), `
			UPDATE attachments
			SET status = 'uploaded', uploaded_at = NOW()
			WHERE id = $1
			RETURNING id, filename, content_type, size_bytes, status
		`, req.AttachmentID).Scan(
			&attachment.ID, &attachment.Filename, &attachment.ContentType,
			&attachment.SizeBytes, &attachment.Status,
		)

		if err != nil {
			log.Error().Err(err).Msg("Failed to complete upload")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"Attachment not found","code":"NOT_FOUND"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(attachment)
	}
}

// sdkUploadFileHandler handles PUT /sdk/attachments/{attachmentId}/upload
func sdkUploadFileHandler(dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		attachmentID := chi.URLParam(r, "attachmentId")

		// Verify attachment exists and is pending
		var gcsPath string
		err := dbPool.QueryRow(r.Context(), `
			SELECT gcs_path FROM attachments WHERE id = $1 AND status = 'pending'
		`, attachmentID).Scan(&gcsPath)

		if err != nil {
			log.Error().Err(err).Msg("Attachment not found or not pending")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"Attachment not found or upload already completed","code":"NOT_FOUND"}`))
			return
		}

		// For local development, just read and discard the body
		// In production, this would upload to cloud storage
		body, err := io.ReadAll(io.LimitReader(r.Body, 25*1024*1024)) // 25MB limit
		if err != nil {
			log.Error().Err(err).Msg("Failed to read upload body")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"Failed to read upload data","code":"UPLOAD_FAILED"}`))
			return
		}

		log.Info().
			Str("attachment_id", attachmentID).
			Int("size_bytes", len(body)).
			Msg("Received file upload (local dev - file not persisted)")

		// Update attachment status
		_, err = dbPool.Exec(r.Context(), `
			UPDATE attachments SET status = 'uploaded', uploaded_at = NOW() WHERE id = $1
		`, attachmentID)

		if err != nil {
			log.Error().Err(err).Msg("Failed to update attachment status")
		}

		w.WriteHeader(http.StatusOK)
	}
}
