package handler

import (
	"errors"
	"log"
	"net/http"

	"product-service/internal/adapter"
	"product-service/internal/adapter/handler/request"
	"product-service/internal/adapter/handler/response"
	"product-service/internal/core/domain/apperror"
	"product-service/internal/core/domain/entity"
	"product-service/internal/core/service"
	"product-service/utils/conv"

	"github.com/labstack/echo/v4"
)

type CategoryHandlerInterface interface {
	GetAll(c echo.Context) error
	GetBySlug(c echo.Context) error
	Create(c echo.Context) error
	Update(c echo.Context) error
	Delete(c echo.Context) error
}

type categoryHandler struct {
	categoryService service.CategoryServiceInterface
}

// respondError logs one line for the failing branch and answers with the status
// carried by apperror, or 500 for anything that is not an apperror.
func respondError(c echo.Context, tag string, err error) error {
	log.Printf("[%s] %v", tag, err)

	if appErr, ok := errors.AsType[*apperror.Error](err); ok {
		return c.JSON(appErr.Code, response.DefaultResponse{Message: appErr.Message})
	}

	return c.JSON(http.StatusInternalServerError, response.DefaultResponse{Message: err.Error()})
}

// toCategoryResponse maps a list row, deriving TotalProduct from the preloaded
// products. It is only accurate when the repository preloaded them.
func toCategoryResponse(category entity.CategoryEntity) response.CategoryResponse {
	return response.CategoryResponse{
		ID:           category.ID,
		Name:         category.Name,
		Icon:         category.Icon,
		Slug:         category.Slug,
		Status:       conv.BoolToStatus(category.Status),
		TotalProduct: int64(len(category.Products)),
	}
}

// toCategoryDetailResponse maps a single category for the detail endpoint.
func toCategoryDetailResponse(category entity.CategoryEntity) response.CategoryDetailResponse {
	return response.CategoryDetailResponse{
		ID:          category.ID,
		Name:        category.Name,
		Icon:        category.Icon,
		Slug:        category.Slug,
		Status:      conv.BoolToStatus(category.Status),
		Description: category.Description,
	}
}

// Create implements [CategoryHandlerInterface].
func (h *categoryHandler) Create(c echo.Context) error {
	req := request.CategoryRequest{}

	if err := c.Bind(&req); err != nil {
		log.Printf("[CategoryHandler-1] Create: %v", err)
		return c.JSON(http.StatusBadRequest, response.DefaultResponse{Message: err.Error()})
	}

	if err := c.Validate(&req); err != nil {
		log.Printf("[CategoryHandler-2] Create: %v", err)
		return c.JSON(http.StatusUnprocessableEntity, response.DefaultResponse{Message: err.Error()})
	}

	// an omitted status keeps the DB default of published
	status := true
	if req.Status != nil {
		status, _ = conv.StatusToBool(*req.Status)
	}

	reqEntity := entity.CategoryEntity{
		ParentID:    req.ParentID,
		Name:        req.Name,
		Icon:        req.Icon,
		Slug:        req.Slug,
		Status:      status,
		Description: req.Description,
	}

	if err := h.categoryService.Create(c.Request().Context(), reqEntity); err != nil {
		return respondError(c, "CategoryHandler-3 Create", err)
	}

	return c.JSON(http.StatusCreated, response.DefaultResponse{Message: "category created successfully"})
}

// Delete implements [CategoryHandlerInterface].
func (h *categoryHandler) Delete(c echo.Context) error {
	slug := c.Param("slug")
	if slug == "" {
		log.Printf("[CategoryHandler-4] Delete: slug is required")
		return c.JSON(http.StatusBadRequest, response.DefaultResponse{Message: "slug is required"})
	}

	if err := h.categoryService.Delete(c.Request().Context(), slug); err != nil {
		return respondError(c, "CategoryHandler-5 Delete", err)
	}

	return c.JSON(http.StatusOK, response.DefaultResponse{Message: "category deleted successfully"})
}

// GetAll implements [CategoryHandlerInterface].
func (h *categoryHandler) GetAll(c echo.Context) error {
	page := 1
	if parsed, err := conv.StringToInt64(c.QueryParam("page")); err == nil && parsed > 0 {
		page = int(parsed)
	}

	limit := 10
	if parsed, err := conv.StringToInt64(c.QueryParam("limit")); err == nil && parsed > 0 {
		limit = int(parsed)
	}

	query := entity.QueryStringEntity{
		Search:    c.QueryParam("search"),
		Page:      page,
		Limit:     limit,
		OrderBy:   c.QueryParam("order_by"),
		OrderType: c.QueryParam("order_type"),
	}

	// status is optional; anything unparseable is ignored so the list keeps
	// showing all rows instead of silently returning nothing.
	if s := c.QueryParam("status"); s != "" {
		if parsed, ok := conv.StatusToBool(s); ok {
			query.Status = &parsed
		}
	}

	categories, countData, totalPage, err := h.categoryService.GetAll(c.Request().Context(), query)
	if err != nil {
		return respondError(c, "CategoryHandler-6 GetAll", err)
	}

	respCategories := make([]response.CategoryResponse, len(categories))
	for i, category := range categories {
		respCategories[i] = toCategoryResponse(category)
	}

	return c.JSON(http.StatusOK, response.DefaultResponseWithPagination{
		Message: "categories retrieved successfully",
		Data:    respCategories,
		Pagination: &response.PaginationResponse{
			Page:       int64(page),
			PerPage:    int64(limit),
			TotalCount: countData,
			TotalPage:  totalPage,
		},
	})
}

// GetBySlug implements [CategoryHandlerInterface].
func (h *categoryHandler) GetBySlug(c echo.Context) error {
	slug := c.Param("slug")
	if slug == "" {
		log.Printf("[CategoryHandler-7] GetBySlug: slug is required")
		return c.JSON(http.StatusBadRequest, response.DefaultResponse{Message: "slug is required"})
	}

	category, err := h.categoryService.GetBySlug(c.Request().Context(), slug)
	if err != nil {
		return respondError(c, "CategoryHandler-8 GetBySlug", err)
	}

	return c.JSON(http.StatusOK, response.DefaultResponse{
		Message: "category retrieved successfully",
		Data:    toCategoryDetailResponse(*category),
	})
}

// Update implements [CategoryHandlerInterface].
func (h *categoryHandler) Update(c echo.Context) error {
	slug := c.Param("slug")
	if slug == "" {
		log.Printf("[CategoryHandler-9] Update: slug is required")
		return c.JSON(http.StatusBadRequest, response.DefaultResponse{Message: "slug is required"})
	}

	req := request.CategoryRequest{}
	if err := c.Bind(&req); err != nil {
		log.Printf("[CategoryHandler-10] Update: %v", err)
		return c.JSON(http.StatusBadRequest, response.DefaultResponse{Message: err.Error()})
	}

	if err := c.Validate(&req); err != nil {
		log.Printf("[CategoryHandler-11] Update: %v", err)
		return c.JSON(http.StatusUnprocessableEntity, response.DefaultResponse{Message: err.Error()})
	}

	// the repository writes every column, so an omitted status would silently
	// become false; require it instead of guessing
	if req.Status == nil {
		log.Printf("[CategoryHandler-12] Update: status is required")
		return c.JSON(http.StatusUnprocessableEntity, response.DefaultResponse{Message: "status is required"})
	}

	status, ok := conv.StatusToBool(*req.Status)
	if !ok {
		log.Printf("[CategoryHandler-12] Update: invalid status %q", *req.Status)
		return c.JSON(http.StatusUnprocessableEntity, response.DefaultResponse{Message: "status must be published or unpublished"})
	}

	reqEntity := entity.CategoryEntity{
		ParentID:    req.ParentID,
		Name:        req.Name,
		Icon:        req.Icon,
		Slug:        req.Slug,
		Status:      status,
		Description: req.Description,
	}

	if err := h.categoryService.Update(c.Request().Context(), slug, reqEntity); err != nil {
		return respondError(c, "CategoryHandler-13 Update", err)
	}

	return c.JSON(http.StatusOK, response.DefaultResponse{Message: "category updated successfully"})
}

// NewCategoryHandler registers the category routes on e behind the token
// middleware.
//
// ponytail: CheckToken only bars the Customer role on /admin paths, so these
// /categories routes require any logged-in user (customer or admin). Add a
// separate role check here if create/update/delete must become admin-only.
func NewCategoryHandler(
	e *echo.Echo,
	categoryService service.CategoryServiceInterface,
	jwtService service.JWTServiceInterface,
) CategoryHandlerInterface {
	h := &categoryHandler{categoryService: categoryService}

	mid := adapter.NewMiddlewareAdapter(jwtService)
	categoryGroup := e.Group("/categories")
	categoryGroup.Use(mid.CheckToken())

	categoryGroup.GET("", h.GetAll)
	categoryGroup.GET("/:slug", h.GetBySlug)
	categoryGroup.POST("", h.Create)
	categoryGroup.PUT("/:slug", h.Update)
	categoryGroup.DELETE("/:slug", h.Delete)

	return h
}
