// internal/router.go
package internal

import (
	"database/sql"
	"net/http"
	"postmatic-api/config"
	"postmatic-api/internal/internal_middleware"

	// Module handlers
	auth_handler "postmatic-api/internal/module/account/auth/handler"
	google_oauth_handler "postmatic-api/internal/module/account/google_oauth/handler"
	profile_handler "postmatic-api/internal/module/account/profile/handler"
	session_handler "postmatic-api/internal/module/account/session/handler"

	referral_basic_handler "postmatic-api/internal/module/affiliator/referral_basic/handler"

	category_creator_image_handler "postmatic-api/internal/module/app/category_creator_image/handler"
	image_uploader_handler "postmatic-api/internal/module/app/image_uploader/handler"
	payment_method_handler "postmatic-api/internal/module/app/payment_method/handler"
	referral_rule_handler "postmatic-api/internal/module/app/referral_rule/handler"
	rss_handler "postmatic-api/internal/module/app/rss/handler"
	timezone_handler "postmatic-api/internal/module/app/timezone/handler"
	token_product_handler "postmatic-api/internal/module/app/token_product/handler"

	business_image_content_handler "postmatic-api/internal/module/business/business_image_content/handler"
	business_information_handler "postmatic-api/internal/module/business/business_information/handler"
	business_knowledge_handler "postmatic-api/internal/module/business/business_knowledge/handler"
	business_member_handler "postmatic-api/internal/module/business/business_member/handler"
	business_product_handler "postmatic-api/internal/module/business/business_product/handler"
	business_role_handler "postmatic-api/internal/module/business/business_role/handler"
	business_rss_subscription_handler "postmatic-api/internal/module/business/business_rss_subscription/handler"
	business_timezone_pref_handler "postmatic-api/internal/module/business/business_timezone_pref/handler"

	creator_image_handler "postmatic-api/internal/module/creator/creator_image/handler"

	// Module services
	auth_service "postmatic-api/internal/module/account/auth/service"
	google_oauth_service "postmatic-api/internal/module/account/google_oauth/service"
	profile_service "postmatic-api/internal/module/account/profile/service"
	session_service "postmatic-api/internal/module/account/session/service"
	referral_basic_service "postmatic-api/internal/module/affiliator/referral_basic/service"
	category_creator_image_service "postmatic-api/internal/module/app/category_creator_image/service"
	generative_image_model_handler "postmatic-api/internal/module/app/generative_image_model/handler"
	generative_image_model_service "postmatic-api/internal/module/app/generative_image_model/service"
	image_uploader_service "postmatic-api/internal/module/app/image_uploader/service"
	payment_method_service "postmatic-api/internal/module/app/payment_method/service"
	referral_rule_service "postmatic-api/internal/module/app/referral_rule/service"
	rss_service "postmatic-api/internal/module/app/rss/service"
	timezone_service "postmatic-api/internal/module/app/timezone/service"
	token_product_service "postmatic-api/internal/module/app/token_product/service"
	business_image_content_service "postmatic-api/internal/module/business/business_image_content/service"
	business_information_service "postmatic-api/internal/module/business/business_information/service"
	business_knowledge_service "postmatic-api/internal/module/business/business_knowledge/service"
	business_member_service "postmatic-api/internal/module/business/business_member/service"
	business_product_service "postmatic-api/internal/module/business/business_product/service"
	business_role_service "postmatic-api/internal/module/business/business_role/service"
	business_rss_subscription_service "postmatic-api/internal/module/business/business_rss_subscription/service"
	business_timezone_pref_service "postmatic-api/internal/module/business/business_timezone_pref/service"
	creator_image_service "postmatic-api/internal/module/creator/creator_image/service"
	"postmatic-api/internal/module/headless/cloudinary_uploader"
	"postmatic-api/internal/module/headless/queue"
	"postmatic-api/internal/module/headless/s3_uploader"
	"postmatic-api/internal/module/headless/token"
	"postmatic-api/internal/repository/entity"
	repository "postmatic-api/internal/repository/entity"
	emailLimiterRepo "postmatic-api/internal/repository/redis/email_limiter_repository"
	"postmatic-api/internal/repository/redis/invitation_limiter_repository"
	ownedBusinessRepo "postmatic-api/internal/repository/redis/owned_business_repository"
	sessionRepo "postmatic-api/internal/repository/redis/session_repository"

	"github.com/go-chi/chi/v5"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

func NewRouter(db *sql.DB, cfg *config.Config, asynqClient *asynq.Client, rdb *redis.Client) chi.Router {
	// 1. =========== INITIAL REPOSITORY ===========
	store := repository.NewStore(db)

	sessionRepo := sessionRepo.NewSessionRepository(rdb)
	emailLimiterRepo := emailLimiterRepo.NewLimiterEmailRepository(rdb)
	ownedRepo := ownedBusinessRepo.NewOwnedBusinessRepository(rdb)
	invitationLimiterRepo := invitation_limiter_repository.NewLimiterInvitationRepository(rdb)

	ownedMw := internal_middleware.NewOwnedBusiness(store, ownedRepo)

	// 2. =========== INITIAL SERVICE ===========
	// HEADLESS
	tokenSvc := token.NewTokenMaker(cfg)
	cldSvc := cloudinary_uploader.NewService(cfg)
	s3Svc := s3_uploader.NewService(cfg)

	// ✅ asynq client untuk enqueue
	queueProducer := queue.NewProducer(asynqClient)

	// ACCOUNT
	authSvc := auth_service.NewService(store, queueProducer, *cfg, sessionRepo, emailLimiterRepo, *tokenSvc)
	sessSvc := session_service.NewService(sessionRepo, *tokenSvc)
	profSvc := profile_service.NewService(store, queueProducer, *cfg, emailLimiterRepo, *tokenSvc)
	googleSvc := google_oauth_service.NewService(store, queueProducer, *cfg, sessionRepo, emailLimiterRepo, *tokenSvc)
	// BUSINESS
	busInSvc := business_information_service.NewService(store, ownedRepo, queueProducer)
	busKnowledgeSvc := business_knowledge_service.NewService(store)
	busRoleSvc := business_role_service.NewService(store)
	busProductSvc := business_product_service.NewService(store)
	busImageContentSvc := business_image_content_service.NewService(store)
	busMemberSvc := business_member_service.NewService(store, *cfg, queueProducer, tokenSvc, invitationLimiterRepo, ownedRepo)
	// APP
	imageUploaderSvc := image_uploader_service.NewImageUploaderService(cldSvc, s3Svc, store)
	rssSvc := rss_service.NewRSSService(store)
	rssSubscriptionSvc := business_rss_subscription_service.NewService(store, rssSvc)
	timezoneSvc := timezone_service.NewTimezoneService()
	busTimezonePrefSvc := business_timezone_pref_service.NewService(store, timezoneSvc)
	catCreatorImageSvc := category_creator_image_service.NewCategoryCreatorImageService(store)
	referralRuleSvc := referral_rule_service.NewReferralService(store)
	tokenProductSvc := token_product_service.NewTokenProductService(store)
	paymentMethodSvc := payment_method_service.NewService(store)
	generativeImageModelSvc := generative_image_model_service.NewService(store)
	// AFFILIATOR
	referralBasicSvc := referral_basic_service.NewService(store, referralRuleSvc)
	// CREATOR
	creatorImageSvc := creator_image_service.NewService(store, catCreatorImageSvc)

	// 3. =========== INITIAL HANDLER ===========
	// ACCOUNT
	authHandler := auth_handler.NewHandler(authSvc, cfg)
	sessHandler := session_handler.NewHandler(sessSvc)
	profileHandler := profile_handler.NewHandler(profSvc)
	googleOauthHandler := google_oauth_handler.NewHandler(googleSvc, cfg)
	// BUSINESS
	busInHandler := business_information_handler.NewHandler(busInSvc, ownedMw)
	busKnowledgeHandler := business_knowledge_handler.NewHandler(busKnowledgeSvc, ownedMw)
	busRoleHandler := business_role_handler.NewHandler(busRoleSvc, ownedMw)
	busProductHandler := business_product_handler.NewHandler(busProductSvc, ownedMw)
	busRssSubscriptionHandler := business_rss_subscription_handler.NewHandler(rssSubscriptionSvc, ownedMw)
	busTimezonePrefHandler := business_timezone_pref_handler.NewHandler(busTimezonePrefSvc, ownedMw)
	busImageContentHandler := business_image_content_handler.NewHandler(busImageContentSvc, ownedMw)
	busMemberHandler := business_member_handler.NewHandler(busMemberSvc, ownedMw)
	// APP
	imageUploaderHandler := image_uploader_handler.NewHandler(imageUploaderSvc)
	rssHandler := rss_handler.NewHandler(rssSvc)
	timezoneHandler := timezone_handler.NewHandler(timezoneSvc)
	catCreatorImageHandler := category_creator_image_handler.NewHandler(catCreatorImageSvc)
	ruleRefferralHandler := referral_rule_handler.NewHandler(referralRuleSvc)
	tokenProductHandler := token_product_handler.NewHandler(tokenProductSvc)
	paymentMethodHandler := payment_method_handler.NewHandler(paymentMethodSvc)
	generativeImageModelHandler := generative_image_model_handler.NewHandler(generativeImageModelSvc)
	// CREATOR
	creatorImageHandler := creator_image_handler.NewHandler(creatorImageSvc)
	// AFFILIATOR
	referralBasicHandler := referral_basic_handler.NewHandler(referralBasicSvc)

	// 4. =========== INITIAL MIDDLEWARE ===========
	allAllowed := internal_middleware.AuthMiddleware(*tokenSvc, []entity.AppRole{entity.AppRoleAdmin, entity.AppRoleUser})
	adminOnly := internal_middleware.AuthMiddleware(*tokenSvc, []entity.AppRole{entity.AppRoleAdmin})

	// 4. =========== ROUTING ===========
	r := chi.NewRouter()

	r.Route("/business", func(r chi.Router) {
		r.Use(allAllowed)
		r.Use(func(next http.Handler) http.Handler {
			return internal_middleware.ReqFilterMiddleware(next, business_information_service.SORT_BY)
		})
		r.Mount("/information", busInHandler.Routes())
		r.Mount("/knowledge", busKnowledgeHandler.Routes())
		r.Mount("/role", busRoleHandler.Routes())
		r.Mount("/product", busProductHandler.Routes())
		r.Mount("/rss-subscription", busRssSubscriptionHandler.Routes())
		r.Mount("/timezone-pref", busTimezonePrefHandler.Routes())
		r.Mount("/image-content", busImageContentHandler.Routes())
		r.Mount("/member", busMemberHandler.Routes())
	})

	r.Route("/account", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Mount("/", authHandler.Routes())
		})
		r.Route("/google-oauth", func(r chi.Router) {
			r.Mount("/", googleOauthHandler.Routes())
		})
		r.Route("/session", func(r chi.Router) {
			r.Use(allAllowed)
			r.Mount("/", sessHandler.Routes())
		})
		r.Route("/profile", func(r chi.Router) {
			r.Use(allAllowed)
			r.Mount("/", profileHandler.Routes())
		})
	})

	r.Route("/app", func(r chi.Router) {
		r.Use(allAllowed)
		r.Mount("/image-uploader", imageUploaderHandler.Routes())
		r.Route("/rss", func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				fil := append(rss_service.SORT_BY_RSS_CATEGORY, rss_service.SORT_BY_RSS_FEED...)
				return internal_middleware.ReqFilterMiddleware(next, fil)
			})
			r.Mount("/", rssHandler.Routes())
		})
		r.Mount("/timezone", timezoneHandler.Routes())
		r.Route("/category-creator-image", func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				fil := append(category_creator_image_service.SORT_BY_CATEGORY_CREATOR_IMAGE_PRODUCT, category_creator_image_service.SORT_BY_CATEGORY_CREATOR_IMAGE_TYPE...)
				return internal_middleware.ReqFilterMiddleware(next, fil)
			})
			r.Mount("/", catCreatorImageHandler.Routes())
		})

		r.Route("/referral-rule", func(r chi.Router) {
			r.Use(adminOnly)
			r.Mount("/", ruleRefferralHandler.Routes())
		})
		r.Mount("/token-product", tokenProductHandler.Routes())
		r.Mount("/payment-method", paymentMethodHandler.Routes(allAllowed, adminOnly))
		r.Mount("/generative-image-model", generativeImageModelHandler.Routes(allAllowed, adminOnly))
	})

	r.Route("/creator", func(r chi.Router) {
		r.Use(allAllowed)
		r.Use(func(next http.Handler) http.Handler {
			return internal_middleware.ReqFilterMiddleware(next, creator_image_service.SORT_BY)
		})
		r.Mount("/image", creatorImageHandler.Routes())
	})

	r.Route("/affiliator", func(r chi.Router) {
		r.Use(allAllowed)
		r.Mount("/referral-basic", referralBasicHandler.Routes())
	})

	return r
}
