package controller

import (
	"voyagear/src/services"
	"voyagear/utils/apperror"
	"voyagear/utils/constant"

	"github.com/gin-gonic/gin"
)

type WishlistController struct {
	Service *services.WishlistService
}

func SetupWishlistController (service *services.WishlistService) *WishlistController {
	return &WishlistController{
		Service: service,
	}
}

type AddToWiahlistRequest struct {
	ProductID string `json:"product_id"`
}

func (h *WishlistController) GetWiahlist(c *gin.Context) {
	userId, exist := c.Get("user_id")
	if !exist {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error":"User id not found in the context"})
		return
	}

	wishlist, err := h.Service.GetWishlist(userId.(string))
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, appErr.Message)
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error":"Failed to fetch wishlist"})
		return
	}

	c.JSON(constant.SUCCESS, gin.H{
		"message":"Wishlist fetched successfully",
		"wishlist":wishlist,
	})
}

func (h *WishlistController) AddToWishlist(c *gin.Context) {
	userId, exist := c.Get("user_id")
	if !exist {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error":"User id not found in the context"})
		return
	}

	var req AddToWiahlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constant.BADREQUEST, gin.H{"error":"Invalid request body"})
		return
	}

	if req.ProductID == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error":"Product_id required"})
		return
	}

	if err := h.Service.AddToWishlist(userId.(string), req.ProductID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, appErr.Message)
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error":"Failed to add product in to wishlist"})
		return
	}

	c.JSON(constant.CREATED, gin.H{"message":"Product added to the wishlist"})
}

func (h *WishlistController) RemoveFromWishlist(c *gin.Context) {

	userId, exist := c.Get("user_id")
	if !exist {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error":"User id not found in the context"})
		return
	}

	productID := c.Param("id")
	if productID == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error":"product_id is required"})
		return
	}

	if err := h.Service.RemoveFromWishlist(userId.(string), productID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, appErr.Message)
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error":"Failed to remove product from wishlist"})
		return
	}

	c.JSON(constant.SUCCESS, gin.H{"message":"Product removed from wishlist successfully"})
}