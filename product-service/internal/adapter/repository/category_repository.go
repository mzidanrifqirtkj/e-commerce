package repository

import (
	"context"
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
}

type categoryRepository struct {
	db *gorm.DB
}

var categoryOrderColumns = map[string]bool{
	"name": true, "status": true, "slug": true, "created_at": true, "updated_at": true,
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
		Where("name ILIKE ? OR status ILIKE ? OR slug ILIKE ?", search, search, search)

	if err := countMain.Count(&countData).Error; err != nil {
		log.Printf("[CategoryRepository-1] GetCategoryAll: %v", err)
		return nil, 0, 0, err
	}

	if err := c.db.WithContext(ctx).Preload("Products").
		Where("name ILIKE ? OR status ILIKE ? OR slug ILIKE ?", search, search, search).
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
