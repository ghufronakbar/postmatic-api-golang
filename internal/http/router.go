package http

import (
	"database/sql"
	"postmatic-api/config"
	"postmatic-api/internal/http/handler/account_handler"
	"postmatic-api/internal/http/handler/business_handler"
	"postmatic-api/internal/http/middleware"
	"postmatic-api/internal/module/account/auth"
	"postmatic-api/internal/module/account/profile"
	"postmatic-api/internal/module/account/session"
	"postmatic-api/internal/module/business/product"
	"postmatic-api/internal/module/headless/mailer"
	repository "postmatic-api/internal/repository/entity"
	emailLimiterRepo "postmatic-api/internal/repository/redis/email_limiter_repository"
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

	// 2. =========== INITIAL SERVICE ===========
	mailerSvc := mailer.NewService(cfg)
	productSvc := product.NewService(store)
	authSvc := auth.NewService(store, *mailerSvc, *cfg, sessionRepo, emailLimiterRepo)
	sessSvc := session.NewService(sessionRepo)
	profSvc := profile.NewService(store, *mailerSvc, *cfg, emailLimiterRepo)

	// 3. =========== INITIAL HANDLER ===========
	productHandler := business_handler.NewProductHandler(productSvc)
	authHandler := account_handler.NewAuthHandler(authSvc, sessSvc)
	profileHandler := account_handler.NewProfileHandler(profSvc)

	// 4. =========== ROUTING ===========
	r := chi.NewRouter()

	r.Route("/business", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)
		r.Mount("/product", productHandler.Routes())
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
