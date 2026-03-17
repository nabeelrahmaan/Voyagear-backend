package controller

import (
	"math"
	"strconv"
	"voyagear/src/models"
	"voyagear/src/services"
	"voyagear/utils/apperror"
	"voyagear/utils/constant"

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
	Name          string `json:"name"`
	Description   string `json:"description"`
	Price         int    `json:"price"`
	OriginalPrice int    `json:"original_price"`
	Category      string `json:"category"`
	ImageURL      string `json:"image_url"`

	Variants []struct {
		Size     string `json:"size"`
		Quantity int    `json:"quantity"`
	} `json:"variants"`
}

type UpdateProductRequest struct {
	Name          *string `json:"name"`
	Description   *string `json:"description"`
	Price         *int    `json:"price"`
	OriginalPrice *int    `json:"original_price"`
	Category      *string `json:"category"`
	ImageURL      *string `json:"image_url"`
	IsActive      *bool   `json:"is_active"`
	Variants         *[]struct {
		ID       *string `json:"id"`
		Size     string  `json:"size"`
		Quantity int     `json:"quantity"`
	} `json:"sizes"`
}

func (h *ProductController) CreateProduct(c *gin.Context) {

	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constant.BADREQUEST, gin.H{"error": constant.INVALID_REQ})
		return
	}

	if req.Name == "" || req.Price <= 0 || len(req.Variants) == 0 {
		c.JSON(constant.BADREQUEST, gin.H{"error": "name, price, sizes are required"})
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

	for _, s := range req.Variants {
		product.Variants = append(product.Variants, models.Variants{
			Size:     s.Size,
			Quantity: s.Quantity,
		})
	}

	if err := h.Service.CreateProduct(&product); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, appErr.Message)
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to create product"})
		return
	}

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
		c.JSON(constant.BADREQUEST, gin.H{"error": constant.INVALID_REQ})
		return
	}

	var variants *[]services.UpdateProductVarientsInput
	if req.Variants != nil {
		tmp := make([]services.UpdateProductVarientsInput, 0, len(*req.Variants))
		for _, s := range *req.Variants {
			tmp = append(tmp, services.UpdateProductVarientsInput{
				ID:       s.ID,
				Size:     s.Size,
				Quantity: s.Quantity,
			})
		}
		variants = &tmp
	}

	input := services.UpdateProductInput{
		Name:          req.Name,
		Description:   req.Description,
		Price:         req.Price,
		OriginalPrice: req.OriginalPrice,
		ImageURL:      req.ImageURL,
		IsActive:      req.IsActive,
		Category:      req.Category,
		Variants:      variants,
	}

	product, err := h.Service.UpdateProduct(productID, &input)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, appErr.Message)
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to update product"})
		return
	}

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
			c.JSON(appErr.Code, appErr.Message)
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to fetch product"})
		return
	}

	c.JSON(constant.SUCCESS, product)
}

func (h *ProductController) GetAllProducts(c *gin.Context) {

	filter := services.ProductFilter{
		Category: c.Query("category"),
		Search:   c.Query("q"),
		Size:     c.Query("size"),
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

	products, total, err := h.Service.GetAllProducts(filter, page, pageSize, sortBy, orderBy)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, appErr.Message)
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to fetch products"})
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	hasNext := page < totalPages
	hasPrev := page > 1

	c.JSON(constant.SUCCESS, gin.H{
		"data": products,
		"pagination": gin.H{
			"current_page":  page,
			"page_size":     pageSize,
			"total_items":   total,
			"has_next":      hasNext,
			"has_previous":      hasPrev,
			"next_page":     page + 1,
			"previous_page": page - 1,
		},
	})
}

func (h *ProductController) DeleteProduct(c *gin.Context) {

	productID := c.Param("id")
	if productID == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error":"Product ID is required"})
		return
	}

	if err := h.Service.DeleteProduct(productID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, appErr.Message)
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error":"Failed to delete product"})
		return
	}

	c.JSON(constant.SUCCESS, gin.H{"message":"Product deleted successfully"})
}

func (h *ProductController) SearchProduct (c *gin.Context) {

	query := c.Query("q")
	category := c.Query("category")

	products, err := h.Service.SearchProduct(query, category)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, appErr.Message)
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to fetch products"})
		return
	}

	c.JSON(constant.SUCCESS, products)
}
