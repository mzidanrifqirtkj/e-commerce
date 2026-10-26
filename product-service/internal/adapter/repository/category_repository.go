package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"product-service/internal/core/domain/apperror"
	"product-service/internal/core/domain/entity"
	"product-service/internal/core/domain/model"

	"gorm.io/gorm"
)

type CategoryRepositoryInterface interface {
	GetAll(ctx context.Context, queryString entity.QueryStringEntity) ([]entity.CategoryEntity, int64, int64, error)
	GetByID(ctx context.Context, categoryID int64) (*entity.CategoryEntity, error)
	Create(ctx context.Context, req entity.CategoryEntity) error
	Update(ctx context.Context, slug string, req entity.CategoryEntity) error
	Delete(ctx context.Context, slug string) error
}

type categoryRepository struct {
	db *gorm.DB
}

// GetByID implements [CategoryRepositoryInterface].
func (c *categoryRepository) GetByID(ctx context.Context, categoryID int64) (*entity.CategoryEntity, error) {
	modelCategory := model.Category{}

	if err := c.db.WithContext(ctx).Preload("Products").
		Where("id = ?", categoryID).
		First(&modelCategory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[CategoryRepository-4] GetByID: category %d not found", categoryID)
			return nil, apperror.ErrNotFound
		}
		log.Printf("[CategoryRepository-5] GetByID: %v", err)
		return nil, err
	}

	category := toCategoryEntity(modelCategory)
	return &category, nil
}

var categoryOrderColumns = map[string]bool{
	"name": true, "status": true, "slug": true, "created_at": true, "updated_at": true,
}

// Create implements [CategoryRepositoryInterface].
func (c *categoryRepository) Create(ctx context.Context, req entity.CategoryEntity) error {
	modelCategory := model.Category{
		ParentID:    req.ParentID,
		Name:        req.Name,
		Icon:        req.Icon,
		Status:      req.Status,
		Slug:        req.Slug,
		Description: req.Description,
	}

	// ponytail: pre-check instead of parsing the unique-violation error; race-safe
	// only via the DB constraint, so map 23505 here later if that matters.
	var exists int64
	if err := c.db.WithContext(ctx).Model(&model.Category{}).
		Where("slug = ?", req.Slug).Count(&exists).Error; err != nil {
		log.Printf("[CategoryRepository-6] Create: %v", err)
		return err
	}
	if exists > 0 {
		log.Printf("[CategoryRepository-7] Create: slug %q already exists", req.Slug)
		return apperror.ErrConflict
	}

	if err := c.db.WithContext(ctx).Create(&modelCategory).Error; err != nil {
		log.Printf("[CategoryRepository-8] Create: %v", err)
		return err
	}

	return nil
}

// Update implements [CategoryRepositoryInterface].
func (c *categoryRepository) Update(ctx context.Context, slug string, req entity.CategoryEntity) error {
	modelCategory := model.Category{}

	if err := c.db.WithContext(ctx).Where("slug = ?", slug).First(&modelCategory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[CategoryRepository-9] Update: category %q not found", slug)
			return apperror.ErrNotFound
		}
		log.Printf("[CategoryRepository-10] Update: %v", err)
		return err
	}

	// map so zero values (empty description, status false) still get written
	updates := map[string]any{
		"parent_id":   req.ParentID,
		"name":        req.Name,
		"icon":        req.Icon,
		"status":      req.Status,
		"slug":        req.Slug,
		"description": req.Description,
	}

	if err := c.db.WithContext(ctx).Model(&modelCategory).Updates(updates).Error; err != nil {
		log.Printf("[CategoryRepository-11] Update: %v", err)
		return err
	}

	return nil
}

// Delete implements [CategoryRepositoryInterface].
func (c *categoryRepository) Delete(ctx context.Context, slug string) error {
	modelCategory := model.Category{}

	if err := c.db.WithContext(ctx).Preload("Products").
		Where("slug = ?", slug).First(&modelCategory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[CategoryRepository-12] Delete: category %q not found", slug)
			return apperror.ErrNotFound
		}
		log.Printf("[CategoryRepository-13] Delete: %v", err)
		return err
	}

	// ponytail: refuse instead of letting ON DELETE CASCADE wipe the products;
	// drop this guard if a category delete should take its products with it.
	if len(modelCategory.Products) > 0 {
		log.Printf("[CategoryRepository-14] Delete: category %q still has %d products", slug, len(modelCategory.Products))
		return apperror.ErrConflict
	}

	// products.category_slug is ON DELETE CASCADE, so the rows go with it
	if err := c.db.WithContext(ctx).Delete(&modelCategory).Error; err != nil {
		log.Printf("[CategoryRepository-15] Delete: %v", err)
		return err
	}

	return nil
}

// GetAll implements [CategoryRepositoryInterface].
func (c *categoryRepository) GetAll(ctx context.Context, query entity.QueryStringEntity) ([]entity.CategoryEntity, int64, int64, error) {
	modelCategories := []model.Category{}
	var countData int64

	// ponytail: order column comes from client, whitelist it instead of raw SQL
	orderBy := "created_at"
	if categoryOrderColumns[query.OrderBy] {
		orderBy = query.OrderBy
	}
	orderType := "desc"
	if query.OrderType == "asc" {
		orderType = "asc"
	}
	order := fmt.Sprintf("%s %s", orderBy, orderType)

	limit := max(query.Limit, 1)
	page := max(query.Page, 1)
	offset := (page - 1) * limit

	search := "%" + query.Search + "%"
	// separate chains: GORM statement cannot be reused after Count
	countMain := c.db.WithContext(ctx).Model(&model.Category{}).
		Where("name ILIKE ? OR status::text ILIKE ? OR slug ILIKE ?", search, search, search)

	if err := countMain.Count(&countData).Error; err != nil {
		log.Printf("[CategoryRepository-1] GetCategoryAll: %v", err)
		return nil, 0, 0, err
	}

	if err := c.db.WithContext(ctx).Preload("Products").
		Where("name ILIKE ? OR status::text ILIKE ? OR slug ILIKE ?", search, search, search).
		Order(order).Limit(limit).Offset(offset).
		Find(&modelCategories).Error; err != nil {
		log.Printf("[CategoryRepository-2] GetCategoryAll: %v", err)
		return nil, 0, 0, err
	}

	if len(modelCategories) < 1 {
		log.Printf("[CategoryRepository-3] GetCategoryAll: no category found")
		return nil, 0, 0, apperror.ErrNotFound
	}

	totalPage := int(math.Ceil(float64(countData) / float64(limit)))

	categories := make([]entity.CategoryEntity, len(modelCategories))
	for i, mc := range modelCategories {
		categories[i] = toCategoryEntity(mc)
	}

	return categories, countData, int64(totalPage), nil
}

func toCategoryEntity(m model.Category) entity.CategoryEntity {
	products := make([]entity.ProductEntity, len(m.Products))
	for i, p := range m.Products {
		products[i] = entity.ProductEntity{
			ID:           p.ID,
			CategorySlug: p.CategorySlug,
			ParentID:     p.ParentID,
			Name:         p.Name,
			Image:        p.Image,
			Description:  p.Description,
			RegularPrice: p.RegularPrice,
			SalePrice:    p.SalePrice,
			Unit:         p.Unit,
			Weight:       p.Weight,
			Stock:        p.Stock,
			Variant:      p.Variant,
			Status:       p.Status,
			CategoryName: m.Name,
		}
	}

	return entity.CategoryEntity{
		ID:          m.ID,
		ParentID:    m.ParentID,
		Name:        m.Name,
		Icon:        m.Icon,
		Status:      m.Status,
		Slug:        m.Slug,
		Description: m.Description,
		Products:    products,
	}
}

func NewCategoryRepository(db *gorm.DB) CategoryRepositoryInterface {
	return &categoryRepository{db: db}
}
