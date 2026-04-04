package controller

import (
	"math"
	"strconv"
	"voyagear/src/models"
	"voyagear/src/repository"
	"voyagear/src/services"
	"voyagear/utils/apperror"
	"voyagear/utils/constant"
	"voyagear/utils/logger"
	"voyagear/utils/validation"

	"github.com/gin-gonic/gin"
)

type ProductController struct {
	Service *services.ProductService
}

func SetupProductController(service *services.ProductService) *ProductController {
	return &ProductController{
		Service: service,
	}
}

type CreateProductRequest struct {
	Name          string `json:"name" validate:"required"`
	Description   string `json:"description" validate:"required"`
	Price         int    `json:"price" validate:"required"`
	OriginalPrice int    `json:"original_price"`
	Category      string `json:"category" validate:"required"`
	ImageURL      string `json:"image_url"`

	Variant []struct {
		Size     string `json:"size" validate:"required"`
		Quantity int    `json:"quantity" validate:"required"`
	} `json:"variant" validate:"required"`
}

type UpdateProductRequest struct {
	Name          *string `json:"name"`
	Description   *string `json:"description"`
	Price         *int    `json:"price"`
	OriginalPrice *int    `json:"original_price"`
	Category      *string `json:"category"`
	ImageURL      *string `json:"image_url"`
	IsActive      *bool   `json:"is_active"`
	Variant       *[]struct {
		ID       string `json:"id"`
		Size     string  `json:"size"`
		Quantity int     `json:"quantity"`
	} `json:"variant"`
}

func (h *ProductController) CreateProduct(c *gin.Context) {

	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constant.BADREQUEST, validation.FormatValidationErrors(err))
		return
	}

	if req.Name == "" || req.Price <= 0 || len(req.Variant) == 0 {
		c.JSON(constant.BADREQUEST, gin.H{"error": "name, price, sizes are required explicitly"})
		return
	}

	product := models.Product{
		Name:          req.Name,
		Description:   req.Description,
		Price:         req.Price,
		OriginalPrice: req.OriginalPrice,
		ImageURL:      req.ImageURL,
		Category:      req.Category,
		IsActive:      true,
	}

	for _, s := range req.Variant {
		product.Variants = append(product.Variants, models.Variant{
			Size:     s.Size,
			Quantity: s.Quantity,
		})
	}

	if err := h.Service.CreateProduct(&product); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("[ADMIN] Failed injecting physical structure payload into product catalog dynamically: %v", err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to create product"})
		return
	}
	logger.Log.Infof("[ADMIN] Successfully injected new mechanical Product %s locally internally", req.Name)
	c.JSON(constant.CREATED, product)
}

func (h *ProductController) UpdateProduct(c *gin.Context) {

	productID := c.Param("id")
	if productID == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error": "Product id is required"})
		return
	}

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constant.BADREQUEST, validation.FormatValidationErrors(err))
		return
	}

	var variant *[]services.UpdateProductVarientsInput
	if req.Variant != nil {
		tmp := make([]services.UpdateProductVarientsInput, 0, len(*req.Variant))
		for _, s := range *req.Variant {
			tmp = append(tmp, services.UpdateProductVarientsInput{
				ID:       s.ID,
				Size:     s.Size,
				Quantity: s.Quantity,
			})
		}
		variant = &tmp
	}

	input := services.UpdateProductInput{
		Name:          req.Name,
		Description:   req.Description,
		Price:         req.Price,
		OriginalPrice: req.OriginalPrice,
		ImageURL:      req.ImageURL,
		IsActive:      req.IsActive,
		Category:      req.Category,
		Variant:       variant,
	}

	product, err := h.Service.UpdateProduct(productID, &input)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("[ADMIN] Mechanical patch denied functionally against Product ID %s globally: %v", productID, err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to update product"})
		return
	}
	logger.Log.Infof("[ADMIN] Executed internal patch updating catalog natively securely for Product ID %s", productID)
	c.JSON(constant.SUCCESS, product)
}

func (h *ProductController) GetProductById(c *gin.Context) {

	productID := c.Param("id")
	if productID == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error": "Product ID is required"})
		return
	}

	product, err := h.Service.GetProductById(productID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("Internal system dynamically failed fetching specific Product ID %s explicitly: %v", productID, err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to fetch product"})
		return
	}
	logger.Log.Infof("Returned individual explicit details fetching Product ID %s natively", productID)
	c.JSON(constant.SUCCESS, product)
}

func (h *ProductController) GetAllProducts(c *gin.Context) {

	filter := repository.ProductFilter{
		Category: c.Query("category"),
		Search:   c.Query("q"),
		Size:     c.Query("size"),
	}

	 userID := ""

	uID, exist := c.Get("user_id")
	if exist {
		userID = uID.(string)
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	filter.MinPrice, _ = strconv.Atoi(c.Query("min_price"))
	filter.MaxPrice, _ = strconv.Atoi(c.Query("max_price"))

	sortBy := c.DefaultQuery("sort_by", "created_at")
	orderBy := c.DefaultQuery("sort_order", "desc")

	products, total, err := h.Service.GetAllProducts(filter, page, pageSize, sortBy, orderBy, userID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("Database parameters completely halted fetching core catalog array structurally: %v", err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to fetch products"})
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	hasNext := page < totalPages
	hasPrev := page > 1

	logger.Log.Infof("Securely routed memory mapping total block %d objects natively fetched strictly", len(products))
	c.JSON(constant.SUCCESS, gin.H{
		"data": products,
		"pagination": gin.H{
			"current_page":  page,
			"page_size":     pageSize,
			"total_items":   total,
			"has_next":      hasNext,
			"has_previous":  hasPrev,
			"next_page":     page + 1,
			"previous_page": page - 1,
		},
	})
}

func (h *ProductController) DeleteProduct(c *gin.Context) {

	productID := c.Param("id")
	if productID == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error": "Product ID is required"})
		return
	}

	if err := h.Service.DeleteProduct(productID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("[ADMIN] Safely stopped physically decoupling central mappings for Product %s: %v", productID, err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to delete product"})
		return
	}
	logger.Log.Infof("[ADMIN] Safely decoupled mapped product securely natively for ID %s internally", productID)
	c.JSON(constant.SUCCESS, gin.H{"message": "Product deleted successfully"})
}

func (h *ProductController) SearchProduct(c *gin.Context) {

	query := c.Query("q")
	category := c.Query("category")

	products, err := h.Service.SearchProduct(query, category)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("Structurally failed resolving fuzzy native queries internally: %v", err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to fetch products"})
		return
	}
	logger.Log.Infof("Strict search memory passed dynamically physically tracking %d exact mapping matches securely", len(products))
	c.JSON(constant.SUCCESS, products)
}
