// internal/module/app/image_uploader/handler/handler.go
package image_uploader_handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	image_uploader_service "postmatic-api/internal/module/app/image_uploader/service"
	"postmatic-api/internal/module/headless/s3_uploader"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	imageUploaderService *image_uploader_service.ImageUploaderService
}

func NewHandler(imageUploaderService *image_uploader_service.ImageUploaderService) *Handler {
	return &Handler{imageUploaderService: imageUploaderService}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/upload-single-image", h.UploadSingleImage)
	r.Post("/presign-upload-image", h.PresignUploadImage)

	return r
}

func (h *Handler) UploadSingleImage(w http.ResponseWriter, r *http.Request) {
	// 1) Limit ukuran upload (mis. 10MB)
	const maxUpload = 10 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)

	// 2) Parse multipart
	if err := r.ParseMultipartForm(maxUpload); err != nil {
		response.Error(w, r, errs.NewBadRequest("INVALID_MULTIPART_FORM"), nil)
		return
	}

	// 3) Ambil file
	f, _, err := r.FormFile("image")
	if err != nil {
		response.Error(w, r, errs.NewValidationFailed(map[string]string{
			"image": "INVALID_FILE",
		}), nil)
		return
	}
	defer f.Close()

	rs, ok := f.(io.ReadSeeker)
	if !ok {
		response.Error(w, r, errs.NewInternalServerError(errors.New("FILE_NOT_SEEKABLE")), nil)
		return
	}

	// 4) Upload file
	head := make([]byte, 512)
	n, _ := io.ReadFull(rs, head)
	contentType := http.DetectContentType(head[:n])

	// multipart.File itu io.Seeker, jadi bisa balik ke awal
	if seeker, ok := f.(io.Seeker); ok {
		_, _ = seeker.Seek(0, io.SeekStart)
	}

	if !strings.HasPrefix(contentType, "image/") {
		response.Error(w, r, errs.NewValidationFailed(map[string]string{
			"image": "INVALID_FILE_TYPE",
		}), nil)
		return
	}

	// 5) Upload ke Cloudinary (beri timeout)
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	upRes, err := h.imageUploaderService.UploadSingleImage(ctx, f)
	if err != nil {
		response.Error(w, r, errs.NewInternalServerError(errors.New(err.Error())), nil)
		return
	}
	if upRes.ImageUrl == "" || upRes.PublicId == "" || upRes.Hashkey == "" {
		response.Error(w, r, errs.NewInternalServerError(errors.New("FAILED_UPLOAD_IMAGE")), nil)
		return
	}

	response.OK(w, r, "SUCCESS_UPLOAD_IMAGE", upRes)

}

func (h *Handler) PresignUploadImage(w http.ResponseWriter, r *http.Request) {
	var req s3_uploader.PresignUploadImageInput
	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	res, err := h.imageUploaderService.PresignUploadImage(r.Context(), req)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}
	response.OK(w, r, "SUCCESS_PRESIGN_UPLOAD_IMAGE", res)
}
