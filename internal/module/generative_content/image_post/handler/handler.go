// internal/module/generative_content/image_post/handler/handler.go
package image_post_handler

import (
	"net/http"
	"strconv"

	"postmatic-api/internal/internal_middleware"
	image_post_service "postmatic-api/internal/module/generative_content/image_post/service"
	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc        *image_post_service.ImagePostService
	middleware *internal_middleware.OwnedBusiness
}

func NewHandler(svc *image_post_service.ImagePostService, ownedMw *internal_middleware.OwnedBusiness) *Handler {
	return &Handler{svc: svc, middleware: ownedMw}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Route("/{businessId}", func(r chi.Router) {
		r.Use(h.middleware.OwnedBusinessMiddleware)
		r.Use(func(next http.Handler) http.Handler {
			return internal_middleware.ReqFilterMiddleware(next, image_post_service.SORT_BY)
		})
		r.Get("/", h.GetAllImagePosts)
		r.Get("/{postId}", h.GetImagePostById)
		r.Post("/", h.CreateImagePost)
	})

	return r
}

func (h *Handler) GetAllImagePosts(w http.ResponseWriter, r *http.Request) {
	bus, err := internal_middleware.OwnedBusinessFromContext(r.Context())
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	filter := internal_middleware.GetFilterFromContext(r.Context())
	q := r.URL.Query()

	// Build filter
	filterData := image_post_service.GetImagePostsFilter{
		BusinessRootID: bus.BusinessRootID,
		SortBy:         filter.SortBy,
		SortDir:        filter.Sort,
		PageOffset:     filter.Offset(),
		PageLimit:      filter.Limit,
		Page:           filter.Page,
		DateStart:      filter.DateStart,
		DateEnd:        filter.DateEnd,
	}

	// Optional status filter
	if status := q.Get("status"); status != "" {
		filterData.Status = &status
	}

	// Optional mode filter
	if mode := q.Get("mode"); mode != "" {
		filterData.Mode = &mode
	}

	// Optional product filter
	if productIdStr := q.Get("productId"); productIdStr != "" {
		productId, err := strconv.ParseInt(productIdStr, 10, 64)
		if err != nil {
			response.ValidationFailed(w, r, map[string]string{"productId": "PRODUCT_ID_MUST_BE_INTEGER"})
			return
		}
		filterData.BusinessProductID = &productId
	}

	res, pag, err := h.svc.GetAllImagePosts(r.Context(), filterData)
	if err != nil {
		response.Error(w, r, err, res)
		return
	}

	response.LIST(w, r, "GET_IMAGE_POSTS_SUCCESS", res, &filter, pag)
}

func (h *Handler) GetImagePostById(w http.ResponseWriter, r *http.Request) {
	bus, err := internal_middleware.OwnedBusinessFromContext(r.Context())
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	postIdStr := chi.URLParam(r, "postId")
	postId, err := strconv.ParseInt(postIdStr, 10, 64)
	if err != nil {
		response.ValidationFailed(w, r, map[string]string{"postId": "POST_ID_MUST_BE_INTEGER"})
		return
	}

	res, err := h.svc.GetImagePostById(r.Context(), postId, bus.BusinessRootID)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "GET_IMAGE_POST_SUCCESS", res)
}

func (h *Handler) CreateImagePost(w http.ResponseWriter, r *http.Request) {
	bus, err := internal_middleware.OwnedBusinessFromContext(r.Context())
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	var req image_post_service.CreateImagePostInput
	req.BusinessRootID = bus.BusinessRootID

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	res, err := h.svc.CreateImagePost(r.Context(), req)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "CREATE_IMAGE_POST_SUCCESS", res)
}
