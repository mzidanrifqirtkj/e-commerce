package service

import (
	"context"
	"product-service/internal/adapter/repository"
	"product-service/internal/core/domain/entity"
)

type ProductServiceInterface interface {
	GetAll(ctx context.Context, req entity.QueryStringProduct) ([]entity.ProductEntity, int64, int64, error)
	GetByID(ctx context.Context, id uint) (*entity.ProductEntity, error)
	Create(ctx context.Context, req entity.ProductEntity) error
	Update(ctx context.Context, id uint, req entity.ProductEntity) error
	Delete(ctx context.Context, id uint) error
}

type productService struct {
	repo repository.ProductRepositoryInterface
}

// Create implements [ProductServiceInterface].
func (s *productService) Create(ctx context.Context, req entity.ProductEntity) error {
	return s.repo.Create(ctx, req)
}

// Delete implements [ProductServiceInterface].
func (s *productService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

// GetAll implements [ProductServiceInterface].
func (s *productService) GetAll(ctx context.Context, req entity.QueryStringProduct) ([]entity.ProductEntity, int64, int64, error) {
	return s.repo.GetAll(ctx, req)
}

// GetByID implements [ProductServiceInterface].
func (s *productService) GetByID(ctx context.Context, id uint) (*entity.ProductEntity, error) {
	return s.repo.GetByID(ctx, id)
}

// Update implements [ProductServiceInterface].
func (s *productService) Update(ctx context.Context, id uint, req entity.ProductEntity) error {
	return s.repo.Update(ctx, id, req)
}

func NewProductService(repo repository.ProductRepositoryInterface) ProductServiceInterface {
	return &productService{repo: repo}
}
