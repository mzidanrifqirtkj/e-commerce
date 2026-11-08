package repository

import (
	"context"
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
}

type productcategoryRepository struct {
	db *gorm.DB
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

	type productWithCategory struct {
		model.Product
		CategoryName string
	}

	rows := []productWithCategory{}
	if err := p.db.WithContext(ctx).Model(&model.Product{}).
		Select("products.*, categories.name AS category_name").
		Joins("LEFT JOIN categories ON categories.slug = products.category_slug").
		Where(baseWhere, baseArgs...).
		Order(order).Limit(limit).Offset(offset).
		Scan(&rows).Error; err != nil {
		log.Printf("[ProductRepository-4] GetAll: %v", err)
		return nil, 0, 0, err
	}

	// Empty result is a valid response, not an error.
	if len(rows) < 1 {
		return []entity.ProductEntity{}, countData, 0, nil
	}

	totalPage := int(math.Ceil(float64(countData) / float64(limit)))

	products := make([]entity.ProductEntity, len(rows))
	for i, row := range rows {
		products[i] = entity.ProductEntity{
			ID:           row.ID,
			CategorySlug: row.CategorySlug,
			ParentID:     row.ParentID,
			Name:         row.Name,
			Image:        row.Image,
			Description:  row.Description,
			RegularPrice: row.RegularPrice,
			SalePrice:    row.SalePrice,
			Unit:         row.Unit,
			Weight:       row.Weight,
			Stock:        row.Stock,
			Variant:      row.Variant,
			Status:       row.Status,
			CategoryName: row.CategoryName,
		}
	}

	return products, countData, int64(totalPage), nil
}

func NewProductRepository(db *gorm.DB) ProductRepositoryInterface {
	return &productcategoryRepository{
		db: db,
	}
}
