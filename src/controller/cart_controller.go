package controller

import (
	"voyagear/src/services"
	"voyagear/utils/apperror"
	"voyagear/utils/constant"
	"voyagear/utils/logger"
	"voyagear/utils/validation"

	"github.com/gin-gonic/gin"
)

type CartController struct {
	Service *services.CartService
}

func SetupCartController(service *services.CartService) *CartController {
	return &CartController{
		Service: service,
	}
}

type AddToCartRequest struct {
	ProductID string `json:"product_id" validate:"required"`
	Size      string `json:"size" validate:"required"`
	Quantity  int    `json:"quantity" validate:"required"`
}

type UpdateCartItemRequest struct {
	Size     *string `json:"size"`
	Quantity *int    `json:"quantity"`
}

func (h *CartController) AddToCart(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error": "User id not found in the context"})
		return
	}

	var req AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constant.BADREQUEST, validation.FormatValidationErrors(err))
		return
	}

	if err := h.Service.AddToCart(userID.(string), req.ProductID, req.Size, req.Quantity); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("Failed explicitly to add item to Cart for User %s: %v", userID.(string), err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to add item to cart"})
		return
	}

	logger.Log.Infof("User %s successfully added Product %s explicitly into Cart", userID.(string), req.ProductID)
	c.JSON(constant.CREATED, gin.H{"message": "Product added to cart"})
}

func (h *CartController) GetCart(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error": "User id not found in the context"})
		return
	}

	cart, err := h.Service.GetCart(userID.(string))
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("Failed extracting cart nodes structurally for User %s: %v", userID.(string), err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to fetch cart"})
		return
	}

	logger.Log.Infof("Successfully fetched cart structure mathematically for User %s", userID.(string))
	c.JSON(constant.SUCCESS, gin.H{
		"message": "Cart fetched successfully",
		"cart":    cart,
	})
}

func (h *CartController) UpdateCartItem(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error": "User id not found in the context"})
		return
	}

	itemID := c.Param("id")
	if itemID == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error": "cart item id required"})
		return
	}

	var req UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constant.BADREQUEST, validation.FormatValidationErrors(err))
		return
	}

	if req.Size == nil && req.Quantity == nil {
		c.JSON(constant.BADREQUEST, gin.H{"error": "Nothing to update"})
		return
	}

	if err := h.Service.UpdateCart(userID.(string), itemID, req.Size, req.Quantity); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("Failed modifying specific Cart sequence for User %s globally: %v", userID.(string), err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to update cart item"})
		return
	}

	logger.Log.Infof("Successfully overrode explicit Cart variant specifically for User %s", userID.(string))
	c.JSON(constant.SUCCESS, gin.H{"message": "Cart item updated successfully"})
}

func (h *CartController) RemoveCartItem(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error": "User id not found in the context"})
		return
	}

	itemID := c.Param("id")
	if itemID == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error": "cart item id required"})
		return
	}

	if err := h.Service.RemoveItemFromCart(itemID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("Failed explicitly wiping payload variant %s from central logic globally: %v", itemID, err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to remove cart item"})
		return
	}

	logger.Log.Infof("Deductively removed physical internal memory payload %s permanently for User %s", itemID, userID.(string))
	c.JSON(constant.SUCCESS, gin.H{"message": "Cart item removed successfully"})
}
