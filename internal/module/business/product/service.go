// internal/module/business/product/service.go
package product

import (
	"context"

	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"

	"github.com/google/uuid"
)

type ProductService struct {
	store entity.Store
}

func NewService(store entity.Store) *ProductService {
	return &ProductService{
		store: store,
	}
}

func (s *ProductService) Create(ctx context.Context, input CreateProductInput) (entity.Product, error) {
	if input.Price < 1000 {
		return entity.Product{}, errs.NewBadRequest("Harga tidak boleh dibawah modal (1000)")
	}

	res, err := s.store.CreateProduct(ctx, entity.CreateProductParams{
		Name:  input.Name,
		Price: int32(input.Price),
	})

	if err != nil {
		return entity.Product{}, errs.NewInternalServerError(err)
	}

	return res, nil
}

func (s *ProductService) GetAll(ctx context.Context) ([]entity.Product, error) {
	res, err := s.store.ListProducts(ctx)

	if err != nil {
		return nil, errs.NewInternalServerError(err)
	}

	if len(res) == 0 {
		return []entity.Product{}, nil
	}

	return res, nil
}

func (s *ProductService) GetOne(ctx context.Context, id uuid.UUID) (entity.Product, error) {
	res, err := s.store.GetProductById(ctx, id)

	if err != nil {
		return entity.Product{}, errs.NewNotFound("")
	}

	return res, nil
}

func (s *ProductService) Delete(ctx context.Context, id uuid.UUID) (entity.Product, error) {

	_, err := s.store.GetProductById(ctx, id)
	if err != nil {
		return entity.Product{}, errs.NewNotFound("")
	}

	res, err := s.store.DeleteProduct(ctx, id)

	if err != nil {
		return entity.Product{}, errs.NewInternalServerError(err)
	}

	return res, nil
}
