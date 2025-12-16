// internal/http/handler/business_handler/product_handler.go
package business_handler

import (
	"encoding/json"
	"net/http"
	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"postmatic-api/internal/module/business/product"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ProductHandler struct {
	svc *product.ProductService
}

func NewProductHandler(svc *product.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

func (h *ProductHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Get("/{id}", h.One)
	r.Post("/", h.Create)
	r.Delete("/{id}", h.Delete)
	return r
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	// 1. Gunakan Struct dari Module Product langsung!
	var req product.CreateProductInput

	// 2. Decode langsung ke struct tersebut
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, "Invalid JSON format", nil)
		return
	}

	// 3. Validasi struct tersebut
	// Validator akan membaca tag `validate` yang ada di product.CreateProductInput
	if errsMap := utils.ValidateStruct(req); errsMap != nil {
		response.JSON(w, http.StatusBadRequest, "Validation Failed", errsMap)
		return
	}

	// 4. Panggil Service
	// Tidak perlu mapping manual lagi! (req sudah bertipe product.CreateProductInput)
	res, err := h.svc.Create(r.Context(), req)

	if err != nil {
		response.Error(w, err, nil)
		return
	}

	response.JSON(w, http.StatusCreated, "Berhasil membuat produk", res)
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	res, err := h.svc.GetAll(r.Context())
	if err != nil {
		response.Error(w, err, nil)
		return
	}
	response.OK(w, "Berhasil mendapatkan list produk", res)
}

func (h *ProductHandler) One(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	uuidId, err := uuid.Parse(id)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}
	res, err := h.svc.GetOne(r.Context(), uuidId)

	if err != nil {
		response.Error(w, err, nil)
		return
	}

	response.OK(w, "Berhasil mendapatkan produk", res)
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	uuidId, err := uuid.Parse(id)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}
	res, err := h.svc.Delete(r.Context(), uuidId)

	if err != nil {
		response.Error(w, err, nil)
		return
	}

	response.JSON(w, http.StatusOK, "Berhasil menghapus produk", res)
}
