package controller

import (
	"voyagear/src/services"
	"voyagear/utils/apperror"
	"voyagear/utils/constant"

	"github.com/gin-gonic/gin"
)

type CartController struct {
	Service services.CartService
}

type AddToCartRequest struct {
	ProductID string `json:"product_id"`
	Size string `json:"size"`
	Quantity int `json:"quantity"`
}

type UpdateCartItemRequest struct {
	Size *string `json:"size"`
	Quantity *int `json:"quantity"`
}


func (h *CartController) AddToCart(c *gin.Context) {

	userID, exist := c.Get("user_id")
	if !exist {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error":"User id not found in the context"})
		return
	}

	var req AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		 c.JSON(constant.BADREQUEST, gin.H{"error":"Invalid request body"})
		 return
	}

	if req.ProductID == "" || req.Quantity == 0 || req.Size == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error":"Product_id, size, quantity required"})
		return
	}

	if err := h.Service.AddToCart(userID.(string), req.ProductID, req.Size, req.Quantity); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, appErr.Message)
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error":"Failed to add item to cart"})
		return
	}

	c.JSON(constant.CREATED, gin.H{"message":"Product added to cart"})
}

func (h *CartController) GetCart(c *gin.Context) {

	userID, exist := c.Get("user_id")
	if !exist {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error":"User id not found in the context"})
		return
	}

	cart, err := h.Service.GetCart(userID.(string))
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, appErr.Message)
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error":"Failed to fetch cart"})
		return
	}

	c.JSON(constant.SUCCESS, gin.H{
		"message":"Cart fetched successfully",
		"cart":cart,
	})
}

func (h *CartController) UpdateCartItem(c *gin.Context) {

	userID, exist := c.Get("user_id")
	if !exist {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error":"User id not found in the context"})
		return
	}
	
	itemID := c.Param("id")
	if itemID == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error":"cart item id required"})
		return
	}

	var req UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constant.BADREQUEST, gin.H{"error":"Invalid request body"})
		return
	}

	if req.Size == nil && req.Quantity == nil {
		c.JSON(constant.BADREQUEST, gin.H{"error":"Nothing to update"})
		return
	}

	if err := h.Service.UpdateCart(userID.(string), itemID, req.Size, req.Quantity); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, appErr.Message)
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error":"Failed to update cart item"})
		return
	}

	c.JSON(constant.SUCCESS, gin.H{"message":"Cart item updated successfully"})
}

func (h *CartController) RemoveCartItem(c *gin.Context) {

	_, exist := c.Get("user_id")
	if !exist {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error":"User id not found in the context"})
		return
	}
	
	itemID := c.Param("id")
	if itemID == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error":"cart item id required"})
		return
	}

	if err := h.Service.RemoveItemFromCart(itemID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, appErr.Message)
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error":"Failed to remove cart item"})
		return
	}

	c.JSON(constant.SUCCESS, gin.H{"message":"Cart item removed successfully"})
}