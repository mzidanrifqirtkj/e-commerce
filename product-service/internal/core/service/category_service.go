package service

import (
	"context"
	"product-service/internal/adapter/repository"
	"product-service/internal/core/domain/apperror"
	"product-service/internal/core/domain/entity"
	"product-service/utils/conv"
)

type CategoryServiceInterface interface {
	GetAll(ctx context.Context, query entity.QueryStringEntity) ([]entity.CategoryEntity, int64, int64, error)
	GetBySlug(ctx context.Context, slug string) (*entity.CategoryEntity, error)
	Create(ctx context.Context, req entity.CategoryEntity) error
	Update(ctx context.Context, slug string, req entity.CategoryEntity) error
	Delete(ctx context.Context, slug string) error
}

type categoryService struct {
	repo repository.CategoryRepositoryInterface
}

// Create implements [CategoryServiceInterface].
func (s *categoryService) Create(ctx context.Context, req entity.CategoryEntity) error {
	// client may leave slug empty; derive it from the name
	if req.Slug == "" {
		req.Slug = conv.Slugify(req.Name)
	}

	// a name made only of non-ASCII characters slugifies to "", and "" is a real
	// value in a UNIQUE column: the second such category would collide
	if req.Slug == "" {
		return apperror.ErrInvalidRequest
	}

	return s.repo.Create(ctx, req)
}

// Delete implements [CategoryServiceInterface].
func (s *categoryService) Delete(ctx context.Context, slug string) error {
	return s.repo.Delete(ctx, slug)
}

// GetAll implements [CategoryServiceInterface].
func (s *categoryService) GetAll(ctx context.Context, query entity.QueryStringEntity) ([]entity.CategoryEntity, int64, int64, error) {
	return s.repo.GetAll(ctx, query)
}

// GetBySlug implements [CategoryServiceInterface].
func (s *categoryService) GetBySlug(ctx context.Context, slug string) (*entity.CategoryEntity, error) {
	return s.repo.GetBySlug(ctx, slug)
}

// Update implements [CategoryServiceInterface].
func (s *categoryService) Update(ctx context.Context, slug string, req entity.CategoryEntity) error {
	return s.repo.Update(ctx, slug, req)
}

func NewCategoryService(repo repository.CategoryRepositoryInterface) CategoryServiceInterface {
	return &categoryService{repo: repo}
}
