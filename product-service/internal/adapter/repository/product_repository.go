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
	"strings"

	"gorm.io/gorm"
)

type ProductRepositoryInterface interface {
	GetAll(ctx context.Context, req entity.QueryStringProduct) ([]entity.ProductEntity, int64, int64, error)
	GetByID(ctx context.Context, id uint) (*entity.ProductEntity, error)
}

type productcategoryRepository struct {
	db *gorm.DB
}

// GetByID implements [ProductRepositoryInterface].
func (p *productcategoryRepository) GetByID(ctx context.Context, id uint) (*entity.ProductEntity, error) {
	var modelProduct model.Product
	if err := p.db.WithContext(ctx).
		Preload("Category").
		Preload("Children.Category").
		First(&modelProduct, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[ProductRepository-5] GetByID: product %d not found", id)
			return nil, apperror.ErrNotFound
		}
		log.Printf("[ProductRepository-6] GetByID: %v", err)
		return nil, err
	}

	product := toProductEntity(modelProduct)
	return &product, nil
}

var productOrderColumns = map[string]bool{
	"name": true, "regular_price": true, "sale_price": true,
	"stock": true, "created_at": true, "updated_at": true,
}

// GetAll implements [ProductRepositoryInterface].
func (p *productcategoryRepository) GetAll(ctx context.Context, req entity.QueryStringProduct) ([]entity.ProductEntity, int64, int64, error) {
	// Query params parse as int64, so a missing price becomes 0. Negative means
	// the client sent a malformed value; reject it instead of silently filtering
	// everything out.
	if req.StartPrice < 0 || req.EndPrice < 0 {
		log.Printf("[ProductRepository-1] GetAll: start_price %d or end_price %d is negative", req.StartPrice, req.EndPrice)
		return nil, 0, 0, apperror.ErrInvalidRequest
	}

	// 0 = filter not sent, so only compare the range when both sides are set.
	if req.StartPrice > 0 && req.EndPrice > 0 && req.StartPrice > req.EndPrice {
		log.Printf("[ProductRepository-2] GetAll: start_price %d greater than end_price %d", req.StartPrice, req.EndPrice)
		return nil, 0, 0, apperror.ErrInvalidRequest
	}

	categorySlug := strings.TrimSpace(req.CategorySlug)

	orderBy := "created_at"
	if productOrderColumns[req.OrderBy] {
		orderBy = req.OrderBy
	}
	orderType := "desc"
	if req.OrderType == "asc" {
		orderType = "asc"
	}
	order := fmt.Sprintf("%s %s", orderBy, orderType)

	limit := max(req.Limit, 1)
	page := max(req.Page, 1)
	offset := (page - 1) * limit

	search := "%" + req.Search + "%"
	baseWhere := "name ILIKE ? OR description ILIKE ? OR unit ILIKE ?"
	baseArgs := []any{search, search, search}

	if categorySlug != "" {
		baseWhere += " AND category_slug = ?"
		baseArgs = append(baseArgs, categorySlug)
	}
	if req.StartPrice > 0 {
		baseWhere += " AND sale_price >= ?"
		baseArgs = append(baseArgs, req.StartPrice)
	}
	if req.EndPrice > 0 {
		baseWhere += " AND sale_price <= ?"
		baseArgs = append(baseArgs, req.EndPrice)
	}

	var countData int64
	countMain := p.db.WithContext(ctx).Model(&model.Product{}).
		Where(baseWhere, baseArgs...)

	if err := countMain.Count(&countData).Error; err != nil {
		log.Printf("[ProductRepository-3] GetAll: %v", err)
		return nil, 0, 0, err
	}

	modelProducts := []model.Product{}
	if err := p.db.WithContext(ctx).
		Preload("Category").
		Preload("Children.Category").
		Where(baseWhere, baseArgs...).
		Order(order).Limit(limit).Offset(offset).
		Find(&modelProducts).Error; err != nil {
		log.Printf("[ProductRepository-4] GetAll: %v", err)
		return nil, 0, 0, err
	}

	// Empty result is a valid response, not an error.
	if len(modelProducts) < 1 {
		return []entity.ProductEntity{}, countData, 0, nil
	}

	totalPage := int(math.Ceil(float64(countData) / float64(limit)))

	products := make([]entity.ProductEntity, len(modelProducts))
	for i, modelProduct := range modelProducts {
		products[i] = toProductEntity(modelProduct)
	}

	return products, countData, int64(totalPage), nil
}

func toProductEntity(m model.Product) entity.ProductEntity {
	child := make([]entity.ProductEntity, len(m.Children))
	for i, c := range m.Children {
		child[i] = toProductEntity(c)
	}

	return entity.ProductEntity{
		ID:           m.ID,
		CategorySlug: m.CategorySlug,
		ParentID:     m.ParentID,
		Name:         m.Name,
		Image:        m.Image,
		Description:  m.Description,
		RegularPrice: m.RegularPrice,
		SalePrice:    m.SalePrice,
		Unit:         m.Unit,
		Weight:       m.Weight,
		Stock:        m.Stock,
		Variant:      m.Variant,
		Status:       m.Status,
		CategoryName: m.Category.Name,
		Child:        child,
	}
}

func NewProductRepository(db *gorm.DB) ProductRepositoryInterface {
	return &productcategoryRepository{
		db: db,
	}
}
