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
	Create(ctx context.Context, req entity.ProductEntity) error
	Update(ctx context.Context, id uint, req entity.ProductEntity) error
	Delete(ctx context.Context, id uint) error
}

type productcategoryRepository struct {
	db *gorm.DB
}

// Create implements [ProductRepositoryInterface].
func (p *productcategoryRepository) Create(ctx context.Context, req entity.ProductEntity) error {
	modelProduct := toProductModel(req)

	// GORM auto-saves the Children association: it inserts the parent first,
	// then fills each child's parent_id with the parent's new ID and inserts it.
	if err := p.db.WithContext(ctx).Create(&modelProduct).Error; err != nil {
		log.Printf("[ProductRepository-7] Create: %v", err)
		return err
	}

	return nil
}

// Update implements [ProductRepositoryInterface].
func (p *productcategoryRepository) Update(ctx context.Context, id uint, req entity.ProductEntity) error {
	modelProduct := model.Product{}

	if err := p.db.WithContext(ctx).Where("id = ?", id).First(&modelProduct).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[ProductRepository-8] Update: product %d not found", id)
			return apperror.ErrNotFound
		}
		log.Printf("[ProductRepository-9] Update: %v", err)
		return err
	}

	// map so zero values (empty image, 0 stock) still get written
	updates := map[string]any{
		"category_slug": req.CategorySlug,
		"parent_id":     req.ParentID,
		"name":          req.Name,
		"image":         req.Image,
		"description":   req.Description,
		"regular_price": req.RegularPrice,
		"sale_price":    req.SalePrice,
		"unit":          req.Unit,
		"weight":        req.Weight,
		"stock":         req.Stock,
		"variant":       req.Variant,
		"status":        req.Status,
	}

	children := make([]model.Product, len(req.Child))
	for i, c := range req.Child {
		children[i] = toProductModel(c)
		children[i].ID = c.ID // existing child must keep its ID, else GORM inserts a duplicate
	}

	return p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&modelProduct).Updates(updates).Error; err != nil {
			log.Printf("[ProductRepository-10] Update: %v", err)
			return err
		}

		// nil = not sent, leave children untouched; empty slice = replace with none.
		if req.Child != nil {
			if err := tx.Model(&modelProduct).Association("Children").Replace(children); err != nil {
				log.Printf("[ProductRepository-16] Update: %v", err)
				return err
			}
		}

		return nil
	})
}

// Delete implements [ProductRepositoryInterface].
func (p *productcategoryRepository) Delete(ctx context.Context, id uint) error {
	modelProduct := model.Product{}

	if err := p.db.WithContext(ctx).Where("id = ?", id).First(&modelProduct).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[ProductRepository-11] Delete: product %d not found", id)
			return apperror.ErrNotFound
		}
		log.Printf("[ProductRepository-12] Delete: %v", err)
		return err
	}

	// ponytail: refuse instead of orphaning the children. If a product delete
	// should cascade to its variants, drop this guard and add ON DELETE CASCADE.
	var childCount int64
	if err := p.db.WithContext(ctx).Model(&model.Product{}).
		Where("parent_id = ?", id).Count(&childCount).Error; err != nil {
		log.Printf("[ProductRepository-13] Delete: %v", err)
		return err
	}
	if childCount > 0 {
		log.Printf("[ProductRepository-14] Delete: product %d still has %d children", id, childCount)
		return apperror.ErrConflict
	}

	if err := p.db.WithContext(ctx).Delete(&modelProduct).Error; err != nil {
		log.Printf("[ProductRepository-15] Delete: %v", err)
		return err
	}

	return nil
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

func toProductModel(e entity.ProductEntity) model.Product {
	children := make([]model.Product, len(e.Child))
	for i, c := range e.Child {
		children[i] = toProductModel(c)
	}

	return model.Product{
		CategorySlug: e.CategorySlug,
		ParentID:     e.ParentID,
		Name:         e.Name,
		Image:        e.Image,
		Description:  e.Description,
		RegularPrice: e.RegularPrice,
		SalePrice:    e.SalePrice,
		Unit:         e.Unit,
		Weight:       e.Weight,
		Stock:        e.Stock,
		Variant:      e.Variant,
		Status:       e.Status,
		Children:     children,
	}
}

func NewProductRepository(db *gorm.DB) ProductRepositoryInterface {
	return &productcategoryRepository{
		db: db,
	}
}
