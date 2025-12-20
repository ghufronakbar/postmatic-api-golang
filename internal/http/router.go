// internal/http/router.go
package http

import (
	"database/sql"
	"net/http"
	"postmatic-api/config"
	"postmatic-api/internal/http/handler/account_handler"
	"postmatic-api/internal/http/handler/business_handler"
	"postmatic-api/internal/http/middleware"
	"postmatic-api/internal/module/account/auth"
	"postmatic-api/internal/module/account/profile"
	"postmatic-api/internal/module/account/session"
	"postmatic-api/internal/module/business/business_information"
	"postmatic-api/internal/module/business/business_knowledge"
	"postmatic-api/internal/module/business/business_role"
	"postmatic-api/internal/module/headless/mailer"
	repository "postmatic-api/internal/repository/entity"
	emailLimiterRepo "postmatic-api/internal/repository/redis/email_limiter_repository"
	ownedBusinessRepo "postmatic-api/internal/repository/redis/owned_business_repository"
	sessionRepo "postmatic-api/internal/repository/redis/session_repository"

	"github.com/go-chi/chi/v5"
)

func NewRouter(db *sql.DB) chi.Router {
	// 1. =========== INITIAL REPOSITORY ===========
	store := repository.NewStore(db)
	cfg := config.Load()
	rdb, err := config.ConnectRedis(cfg)
	if err != nil {
		panic("Cannot connect to Redis" + err.Error())
	}
	sessionRepo := sessionRepo.NewSessionRepository(rdb)
	emailLimiterRepo := emailLimiterRepo.NewLimiterEmailRepository(rdb)
	ownedRepo := ownedBusinessRepo.NewOwnedBusinessRepository(rdb)

	ownedMw := middleware.NewOwnedBusiness(store, ownedRepo)

	// 2. =========== INITIAL SERVICE ===========
	mailerSvc := mailer.NewService(cfg)
	authSvc := auth.NewService(store, *mailerSvc, *cfg, sessionRepo, emailLimiterRepo)
	sessSvc := session.NewService(sessionRepo)
	profSvc := profile.NewService(store, *mailerSvc, *cfg, emailLimiterRepo)
	busInSvc := business_information.NewService(store, ownedRepo)
	busKnowledgeSvc := business_knowledge.NewService(store)
	busRoleSvc := business_role.NewService(store)

	// 3. =========== INITIAL HANDLER ===========
	authHandler := account_handler.NewAuthHandler(authSvc, sessSvc)
	profileHandler := account_handler.NewProfileHandler(profSvc)
	busInHandler := business_handler.NewBusinessInformationHandler(busInSvc, ownedMw)
	busKnowledgeHandler := business_handler.NewBusinessKnowledgeHandler(busKnowledgeSvc, ownedMw)
	busRoleHandler := business_handler.NewBusinessRoleHandler(busRoleSvc, ownedMw)

	// 4. =========== ROUTING ===========
	r := chi.NewRouter()

	r.Route("/business", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)
		r.Use(func(next http.Handler) http.Handler {
			return middleware.ReqFilterMiddleware(next, business_information.SORT_BY)
		})
		r.Mount("/information", busInHandler.BusinessInformationRoutes())
		r.Mount("/knowledge", busKnowledgeHandler.BusinessKnowledgeRoutes())
		r.Mount("/role", busRoleHandler.BusinessRoleRoutes())
	})

	r.Route("/account", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Mount("/", authHandler.AuthRoutes())
		})
		r.Route("/session", func(r chi.Router) {
			r.Use(middleware.AuthMiddleware)
			r.Mount("/", authHandler.SessionRoutes())
		})
		r.Route("/profile", func(r chi.Router) {
			r.Use(middleware.AuthMiddleware)
			r.Mount("/", profileHandler.ProfileRoutes())
		})
	})

	return r
}
