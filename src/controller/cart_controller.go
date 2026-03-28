package controller

import (
	"voyagear/src/services"
	"voyagear/utils/constant"

	"github.com/gin-gonic/gin"
)

type CartController struct {
	Service services.CartService
}

func (h *CartController) GetCart(c *gin.Context) {

	userID, exist := c.Get("user_id")
	if !exist {
		c.JSON(constant.BADREQUEST, gin.H{"error":"User id not found in the context"})
		return
	}

	
}