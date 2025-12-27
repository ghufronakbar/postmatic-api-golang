// internal/http/router.go
package http

import (
	"database/sql"
	"net/http"
	"postmatic-api/config"
	"postmatic-api/internal/http/handler/account_handler"
	"postmatic-api/internal/http/handler/app_handler"
	"postmatic-api/internal/http/handler/business_handler"
	"postmatic-api/internal/http/handler/creator_handler"
	"postmatic-api/internal/http/middleware"
	"postmatic-api/internal/module/account/auth"
	"postmatic-api/internal/module/account/google_oauth"
	"postmatic-api/internal/module/account/profile"
	"postmatic-api/internal/module/account/session"
	"postmatic-api/internal/module/app/category_creator_image"
	"postmatic-api/internal/module/app/image_uploader"
	"postmatic-api/internal/module/app/rss"
	"postmatic-api/internal/module/app/timezone"
	"postmatic-api/internal/module/business/business_information"
	"postmatic-api/internal/module/business/business_knowledge"
	"postmatic-api/internal/module/business/business_product"
	"postmatic-api/internal/module/business/business_role"
	"postmatic-api/internal/module/business/business_rss_subscription"
	"postmatic-api/internal/module/business/business_timezone_pref"
	"postmatic-api/internal/module/creator/creator_image"
	"postmatic-api/internal/module/headless/cloudinary_uploader"
	"postmatic-api/internal/module/headless/mailer"
	"postmatic-api/internal/module/headless/s3_uploader"
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
	// HEADLESS
	mailerSvc := mailer.NewService(cfg)
	cldSvc, err := cloudinary_uploader.NewService(cfg)
	if err != nil {
		panic("Cannot connect to Cloudinary" + err.Error())
	}
	s3Svc, err := s3_uploader.NewService(cfg)
	if err != nil {
		panic("Cannot connect to S3" + err.Error())
	}
	// ACCOUNT
	authSvc := auth.NewService(store, *mailerSvc, *cfg, sessionRepo, emailLimiterRepo)
	sessSvc := session.NewService(sessionRepo)
	profSvc := profile.NewService(store, *mailerSvc, *cfg, emailLimiterRepo)
	googleSvc := google_oauth.NewService(store, *mailerSvc, *cfg, sessionRepo, emailLimiterRepo)
	// BUSINESS
	busInSvc := business_information.NewService(store, ownedRepo)
	busKnowledgeSvc := business_knowledge.NewService(store)
	busRoleSvc := business_role.NewService(store)
	busProductSvc := business_product.NewService(store)
	// APP
	imageUploaderSvc := image_uploader.NewImageUploaderService(cldSvc, s3Svc, store)
	rssSvc := rss.NewRSSService(store)
	rssSubscriptionSvc := business_rss_subscription.NewService(store, rssSvc)
	timezoneSvc := timezone.NewTimezoneService()
	busTimezonePrefSvc := business_timezone_pref.NewService(store, timezoneSvc)
	catCreatorImageSvc := category_creator_image.NewCategoryCreatorImageService(store)
	// CREATOR
	creatorImageSvc := creator_image.NewService(store, catCreatorImageSvc)

	// 3. =========== INITIAL HANDLER ===========
	// ACCOUNT
	authHandler := account_handler.NewAuthHandler(authSvc, cfg)
	sessHandler := account_handler.NewSessionHandler(sessSvc)
	profileHandler := account_handler.NewProfileHandler(profSvc)
	googleOauthHandler := account_handler.NewGoogleOAuthHandler(googleSvc, cfg)
	// BUSINESS
	busInHandler := business_handler.NewBusinessInformationHandler(busInSvc, ownedMw)
	busKnowledgeHandler := business_handler.NewBusinessKnowledgeHandler(busKnowledgeSvc, ownedMw)
	busRoleHandler := business_handler.NewBusinessRoleHandler(busRoleSvc, ownedMw)
	busProductHandler := business_handler.NewBusinessProductHandler(busProductSvc, ownedMw)
	busRssSubscriptionHandler := business_handler.NewBusinessRssSubscriptionHandler(rssSubscriptionSvc, ownedMw)
	busTimezonePrefHandler := business_handler.NewBusinessTimezonePrefHandler(busTimezonePrefSvc, ownedMw)
	// APP
	imageUploaderHandler := app_handler.NewImageUploaderHandler(imageUploaderSvc)
	rssHandler := app_handler.NewRSSHandler(rssSvc)
	timezoneHandler := app_handler.NewTimezoneHandler(timezoneSvc)
	catCreatorImageHandler := app_handler.NewCategoryCreatorImageHandler(catCreatorImageSvc)
	// CREATOR
	creatorImageHandler := creator_handler.NewCreatorImageHandler(creatorImageSvc)

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
		r.Mount("/product", busProductHandler.BusinessProductRoutes())
		r.Mount("/rss-subscription", busRssSubscriptionHandler.BusinessRssSubscriptionRoutes())
		r.Mount("/timezone-pref", busTimezonePrefHandler.BusinessTimezonePrefRoutes())
	})

	r.Route("/account", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Mount("/", authHandler.AuthRoutes())
		})
		r.Route("/google-oauth", func(r chi.Router) {
			r.Mount("/", googleOauthHandler.GoogleOAuthRoutes())
		})
		r.Route("/session", func(r chi.Router) {
			r.Use(middleware.AuthMiddleware)
			r.Mount("/", sessHandler.SessionRoutes())
		})
		r.Route("/profile", func(r chi.Router) {
			r.Use(middleware.AuthMiddleware)
			r.Mount("/", profileHandler.ProfileRoutes())
		})
	})

	r.Route("/app", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)
		r.Mount("/image-uploader", imageUploaderHandler.ImageUploaderRoutes())
		r.Route("/rss", func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				fil := append(rss.SORT_BY_RSS_CATEGORY, rss.SORT_BY_RSS_FEED...)
				return middleware.ReqFilterMiddleware(next, fil)
			})
			r.Mount("/", rssHandler.RSSRoutes())
		})
		r.Mount("/timezone", timezoneHandler.TimezoneRoutes())
		r.Route("/category-creator-image", func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				fil := append(category_creator_image.SORT_BY_CATEGORY_CREATOR_IMAGE_PRODUCT, category_creator_image.SORT_BY_CATEGORY_CREATOR_IMAGE_TYPE...)
				return middleware.ReqFilterMiddleware(next, fil)
			})
			r.Mount("/", catCreatorImageHandler.CategoryCreatorImageRoutes())
		})
	})

	r.Route("/creator", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)
		r.Use(func(next http.Handler) http.Handler {
			return middleware.ReqFilterMiddleware(next, creator_image.SORT_BY)
		})
		r.Mount("/image", creatorImageHandler.CreatorImageRoutes())
	})

	return r
}
