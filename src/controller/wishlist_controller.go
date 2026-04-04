package controller

import (
	"voyagear/src/services"
	"voyagear/utils/apperror"
	"voyagear/utils/constant"
	"voyagear/utils/logger"
	"voyagear/utils/validation"

	"github.com/gin-gonic/gin"
)

type WishlistController struct {
	Service *services.WishlistService
}

func SetupWishlistController(service *services.WishlistService) *WishlistController {
	return &WishlistController{
		Service: service,
	}
}

type AddToWishlistRequest struct {
	ProductID string `json:"product_id" validate:"required"`
}

func (h *WishlistController) GetWishlist(c *gin.Context) {
	userId, exist := c.Get("user_id")
	if !exist {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error": "User id not found in the context"})
		return
	}

	wishlist, err := h.Service.GetWishlist(userId.(string))
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("Wishlist structural extraction completely failed internally: %v", err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to fetch wishlist"})
		return
	}
	logger.Log.Infof("Wishlist memory extracted physically for User %s successfully", userId.(string))
	c.JSON(constant.SUCCESS, gin.H{
		"message":  "Wishlist fetched successfully",
		"wishlist": wishlist,
	})
}

func (h *WishlistController) AddToWishlist(c *gin.Context) {
	userId, exist := c.Get("user_id")
	if !exist {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error": "User id not found in the context"})
		return
	}

	var req AddToWishlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constant.BADREQUEST, validation.FormatValidationErrors(err))
		return
	}

	if req.ProductID == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error": "Product_id required"})
		return
	}

	if err := h.Service.AddToWishlist(userId.(string), req.ProductID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("Wishlist addition structurally aborted: %v", err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": err.Error()})
		return
	}
	logger.Log.Infof("Product explicitly mounted to Wishlist securely for User %s", userId.(string))
	c.JSON(constant.CREATED, gin.H{"message": "Product added to the wishlist"})
}

func (h *WishlistController) RemoveFromWishlist(c *gin.Context) {
	userId, exist := c.Get("user_id")
	if !exist {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error": "User id not found in the context"})
		return
	}

	productID := c.Param("id")
	if productID == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error": "product_id is required"})
		return
	}

	if err := h.Service.RemoveFromWishlist(userId.(string), productID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("Wishlist deletion functionally failed securely: %v", err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to remove product from wishlist"})
		return
	}
	logger.Log.Infof("Wiped Wishlist memory dynamically against Product %s internally", productID)
	c.JSON(constant.SUCCESS, gin.H{"message": "Product removed from wishlist successfully"})
}