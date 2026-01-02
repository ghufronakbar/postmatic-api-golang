// internal/http/router.go
package http

import (
	"database/sql"
	"net/http"
	"postmatic-api/config"
	"postmatic-api/internal/http/handler/account_handler"
	"postmatic-api/internal/http/handler/affiliator_handler"
	"postmatic-api/internal/http/handler/app_handler"
	"postmatic-api/internal/http/handler/business_handler"
	"postmatic-api/internal/http/handler/creator_handler"
	"postmatic-api/internal/http/middleware"
	"postmatic-api/internal/module/account/auth"
	"postmatic-api/internal/module/account/google_oauth"
	"postmatic-api/internal/module/account/profile"
	"postmatic-api/internal/module/account/session"
	"postmatic-api/internal/module/affiliator/referral_basic"
	"postmatic-api/internal/module/app/category_creator_image"
	"postmatic-api/internal/module/app/image_uploader"
	"postmatic-api/internal/module/app/referral_rule"
	"postmatic-api/internal/module/app/rss"
	"postmatic-api/internal/module/app/timezone"
	"postmatic-api/internal/module/business/business_image_content"
	"postmatic-api/internal/module/business/business_information"
	"postmatic-api/internal/module/business/business_knowledge"
	"postmatic-api/internal/module/business/business_member"
	"postmatic-api/internal/module/business/business_product"
	"postmatic-api/internal/module/business/business_role"
	"postmatic-api/internal/module/business/business_rss_subscription"
	"postmatic-api/internal/module/business/business_timezone_pref"
	"postmatic-api/internal/module/creator/creator_image"
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
)

func NewRouter(db *sql.DB, cfg *config.Config, asynqClient *asynq.Client) chi.Router {
	// 1. =========== INITIAL REPOSITORY ===========
	store := repository.NewStore(db)

	rdb, err := config.ConnectRedis(cfg)
	if err != nil {
		panic("Cannot connect to Redis" + err.Error())
	}
	sessionRepo := sessionRepo.NewSessionRepository(rdb)
	emailLimiterRepo := emailLimiterRepo.NewLimiterEmailRepository(rdb)
	ownedRepo := ownedBusinessRepo.NewOwnedBusinessRepository(rdb)
	invitationLimiterRepo := invitation_limiter_repository.NewLimiterInvitationRepository(rdb)

	ownedMw := middleware.NewOwnedBusiness(store, ownedRepo)

	// 2. =========== INITIAL SERVICE ===========
	// HEADLESS
	tokenSvc := token.NewTokenMaker(cfg)
	cldSvc, err := cloudinary_uploader.NewService(cfg)
	if err != nil {
		panic("Cannot connect to Cloudinary" + err.Error())
	}
	s3Svc, err := s3_uploader.NewService(cfg)
	if err != nil {
		panic("Cannot connect to S3" + err.Error())
	}
	// ✅ asynq client untuk enqueue
	queueProducer := queue.NewProducer(asynqClient)

	// ACCOUNT
	authSvc := auth.NewService(store, queueProducer, *cfg, sessionRepo, emailLimiterRepo, *tokenSvc)
	sessSvc := session.NewService(sessionRepo, *tokenSvc)
	profSvc := profile.NewService(store, queueProducer, *cfg, emailLimiterRepo, *tokenSvc)
	googleSvc := google_oauth.NewService(store, queueProducer, *cfg, sessionRepo, emailLimiterRepo, *tokenSvc)
	// BUSINESS
	busInSvc := business_information.NewService(store, ownedRepo, queueProducer)
	busKnowledgeSvc := business_knowledge.NewService(store)
	busRoleSvc := business_role.NewService(store)
	busProductSvc := business_product.NewService(store)
	busImageContentSvc := business_image_content.NewService(store)
	busMemberSvc := business_member.NewService(store, *cfg, queueProducer, tokenSvc, invitationLimiterRepo, ownedRepo)
	// APP
	imageUploaderSvc := image_uploader.NewImageUploaderService(cldSvc, s3Svc, store)
	rssSvc := rss.NewRSSService(store)
	rssSubscriptionSvc := business_rss_subscription.NewService(store, rssSvc)
	timezoneSvc := timezone.NewTimezoneService()
	busTimezonePrefSvc := business_timezone_pref.NewService(store, timezoneSvc)
	catCreatorImageSvc := category_creator_image.NewCategoryCreatorImageService(store)
	referralRuleSvc := referral_rule.NewReferralService(store)
	// AFFILIATOR
	referralBasicSvc := referral_basic.NewService(store, referralRuleSvc)
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
	busImageContentHandler := business_handler.NewBusinessImageContentHandler(busImageContentSvc, ownedMw)
	busMemberHandler := business_handler.NewBusinessMemberHandler(busMemberSvc, ownedMw)
	// APP
	imageUploaderHandler := app_handler.NewImageUploaderHandler(imageUploaderSvc)
	rssHandler := app_handler.NewRSSHandler(rssSvc)
	timezoneHandler := app_handler.NewTimezoneHandler(timezoneSvc)
	catCreatorImageHandler := app_handler.NewCategoryCreatorImageHandler(catCreatorImageSvc)
	ruleRefferralHandler := app_handler.NewReferralRuleHandler(referralRuleSvc)
	// CREATOR
	creatorImageHandler := creator_handler.NewCreatorImageHandler(creatorImageSvc)
	// AFFILIATOR
	referralBasicHandler := affiliator_handler.NewReferralBasicHandler(referralBasicSvc)

	// 4. =========== INITIAL MIDDLEWARE ===========
	allAllowed := middleware.AuthMiddleware(*tokenSvc, []entity.AppRole{entity.AppRoleAdmin, entity.AppRoleUser})
	adminOnly := middleware.AuthMiddleware(*tokenSvc, []entity.AppRole{entity.AppRoleAdmin})

	// 4. =========== ROUTING ===========
	r := chi.NewRouter()

	r.Route("/business", func(r chi.Router) {
		r.Use(allAllowed)
		r.Use(func(next http.Handler) http.Handler {
			return middleware.ReqFilterMiddleware(next, business_information.SORT_BY)
		})
		r.Mount("/information", busInHandler.BusinessInformationRoutes())
		r.Mount("/knowledge", busKnowledgeHandler.BusinessKnowledgeRoutes())
		r.Mount("/role", busRoleHandler.BusinessRoleRoutes())
		r.Mount("/product", busProductHandler.BusinessProductRoutes())
		r.Mount("/rss-subscription", busRssSubscriptionHandler.BusinessRssSubscriptionRoutes())
		r.Mount("/timezone-pref", busTimezonePrefHandler.BusinessTimezonePrefRoutes())
		r.Mount("/image-content", busImageContentHandler.BusinessImageContentRoutes())
		r.Mount("/member", busMemberHandler.BusinessMemberRoutes())
	})

	r.Route("/account", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Mount("/", authHandler.AuthRoutes())
		})
		r.Route("/google-oauth", func(r chi.Router) {
			r.Mount("/", googleOauthHandler.GoogleOAuthRoutes())
		})
		r.Route("/session", func(r chi.Router) {
			r.Use(allAllowed)
			r.Mount("/", sessHandler.SessionRoutes())
		})
		r.Route("/profile", func(r chi.Router) {
			r.Use(allAllowed)
			r.Mount("/", profileHandler.ProfileRoutes())
		})
	})

	r.Route("/app", func(r chi.Router) {
		r.Use(allAllowed)
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

		r.Route("/referral-rule", func(r chi.Router) {
			r.Use(adminOnly)
			r.Mount("/", ruleRefferralHandler.ReferralRuleRoutes())
		})
	})

	r.Route("/creator", func(r chi.Router) {
		r.Use(allAllowed)
		r.Use(func(next http.Handler) http.Handler {
			return middleware.ReqFilterMiddleware(next, creator_image.SORT_BY)
		})
		r.Mount("/image", creatorImageHandler.CreatorImageRoutes())
	})

	r.Route("/affiliator", func(r chi.Router) {
		r.Use(allAllowed)
		r.Mount("/referral-basic", referralBasicHandler.ReferralBasicRoutes())
	})

	return r
}
