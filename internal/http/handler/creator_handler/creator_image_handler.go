// internal/http/handler/creator_handler/creator_image_handler.go
package creator_handler

import (
	"fmt"
	"net/http"
	"postmatic-api/internal/http/middleware"
	"postmatic-api/internal/module/creator/creator_image"
	"strconv"

	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type CreatorImageHandler struct {
	creatorImageSvc *creator_image.CreatorImageService
}

func NewCreatorImageHandler(creatorImageSvc *creator_image.CreatorImageService) *CreatorImageHandler {
	return &CreatorImageHandler{creatorImageSvc: creatorImageSvc}
}

func (h *CreatorImageHandler) CreatorImageRoutes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.CreateCreatorImage)
	r.Put("/{creatorImageId}", h.UpdateCreatorImage)
	r.Delete("/{creatorImageId}", h.SoftDeleteCreatorImage)

	return r
}

func (h *CreatorImageHandler) CreateCreatorImage(w http.ResponseWriter, r *http.Request) {
	var req creator_image.CreateUpdateCreatorImageInput

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}
	prof, _ := middleware.GetUserFromContext(r.Context())

	res, err := h.creatorImageSvc.CreateCreatorImage(r.Context(), req, prof.ID)

	if err != nil {
		fmt.Println(err)
		response.Error(w, r, err, res)
		return
	}

	response.OK(w, r, "CREATE_CREATOR_IMAGE_SUCCESS", res)
}

func (h *CreatorImageHandler) UpdateCreatorImage(w http.ResponseWriter, r *http.Request) {
	var req creator_image.CreateUpdateCreatorImageInput

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}
	prof, _ := middleware.GetUserFromContext(r.Context())

	creatorImageId := chi.URLParam(r, "creatorImageId")
	intCreatorImageId, err := strconv.Atoi(creatorImageId)
	if err != nil {
		response.ValidationFailed(w, r, map[string]string{"creatorImageId": "CREATOR_IMAGE_MUST_BE_INTEGER"})
		return
	}

	res, err := h.creatorImageSvc.UpdateCreatorImage(r.Context(), req, int64(intCreatorImageId), prof.ID)

	if err != nil {
		response.Error(w, r, err, res)
		return
	}

	response.OK(w, r, "UPDATE_CREATOR_IMAGE_SUCCESS", res)
}

func (h *CreatorImageHandler) SoftDeleteCreatorImage(w http.ResponseWriter, r *http.Request) {
	creatorImageId := chi.URLParam(r, "creatorImageId")
	intCreatorImageId, err := strconv.Atoi(creatorImageId)

	if err != nil {
		response.ValidationFailed(w, r, map[string]string{"creatorImageId": "CREATOR_IMAGE_MUST_BE_INTEGER"})
		return
	}

	prof, _ := middleware.GetUserFromContext(r.Context())

	res, err := h.creatorImageSvc.SoftDeleteCreatorImage(r.Context(), int64(intCreatorImageId), prof.ID)

	if err != nil {
		response.Error(w, r, err, res)
		return
	}

	response.OK(w, r, "DELETE_CREATOR_IMAGE_SUCCESS", res)
}
