package controller

import (
	"voyagear/src/services"
	"voyagear/utils/apperror"
	"voyagear/utils/constant"
	"voyagear/utils/logger"
	"voyagear/utils/validation"

	"github.com/gin-gonic/gin"
)

type OrderController struct {
	Service *services.OrderService
}

func SetupOrderController(service *services.OrderService) *OrderController {
	return &OrderController{
		Service: service,
	}
}

type UpdateOrderStatusUserReq struct {
	Status string `json:"status" validate:"required,oneof=PLACED PROCESSING SHIPPED DELIVERED CANCELLED"`
}

func (o *OrderController) GetUserOrders(c *gin.Context) {

	userID, exist := c.Get("user_id")
	if !exist {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error": "user not found"})
		return
	}

	orders, err := o.Service.GetUserOrders(userID.(string))
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("User fetch completely rejected structurally missing queries: %v", err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to fetch orders"})
		return
	}
	logger.Log.Infof("Returned strict list of Order arrays perfectly for User %s", userID.(string))
	c.JSON(constant.SUCCESS, gin.H{
		"message": "Orders retreived successfully",
		"orders":  orders,
	})
}

func (o *OrderController) PlaceOrder(c *gin.Context) {

	userID, exist := c.Get("user_id")
	if !exist {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error": "user not found"})
		return
	}

	var req services.PlaceOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constant.BADREQUEST, validation.FormatValidationErrors(err))
		return
	}

	if req.Type == "" {
		req.Type = "cart"
	}

	order, err := o.Service.PlaceOrder(userID.(string), req)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("Order validation logic physically denied transaction: %v", err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to place order"})
		return
	}
	logger.Log.Infof("User %s successfully instantiated physical order parameters completely natively", userID.(string))
	c.JSON(constant.SUCCESS, gin.H{
		"message": "Order placed successfully",
		"order":   order,
	})
}

func (o *OrderController) GetOrderDetails(c *gin.Context) {

	userID, exist := c.Get("user_id")
	if !exist {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error": "user not found"})
		return
	}

	orderID := c.Param("order_id")
	if orderID == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error": "Order id required"})
		return
	}

	order, err := o.Service.GetOrderDetails(userID.(string), orderID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("Internal parameter rejection dynamically tracking payload: %v", err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to fetch order details"})
		return
	}
	logger.Log.Infof("Audited structure payload fetched internally for order %s", orderID)
	c.JSON(constant.SUCCESS, gin.H{
		"message": "Order fetched successfully",
		"order":   order,
	})
}

func (o *OrderController) UpdateOrderStatusUser(c *gin.Context) {

	userID, exist := c.Get("user_id")
	if !exist {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error": "user not found"})
		return
	}

	orderID := c.Param("order_id")
	if orderID == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error": "Order id required"})
		return
	}

	var req UpdateOrderStatusUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constant.BADREQUEST, validation.FormatValidationErrors(err))
		return
	}

	if err := o.Service.UpdateOrderStatusUser(userID.(string), orderID, req.Status); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("Order physical bounds rejected override mechanically: %v", err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to update order status"})
		return
	}
	logger.Log.Infof("Dynamic mechanical status override completely resolved for order %s", orderID)
	c.JSON(constant.SUCCESS, gin.H{"message": "Order status updated successfully"})
}

func (o *OrderController) DeleteOrder(c *gin.Context) {

	userID, exist := c.Get("user_id")
	if !exist {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error": "user not found"})
		return
	}

	orderID := c.Param("order_id")
	if orderID == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error": "Order id required"})
		return
	}

	if err := o.Service.DeleteOrder(userID.(string), orderID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("Central logic failed structurally mapping deletion array: %v", err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to delete order"})
		return
	}
	logger.Log.Infof("Successfully disconnected order logically entirely %s", orderID)
	c.JSON(constant.SUCCESS, gin.H{"message": "Order deleted successfully"})
}

func (o *OrderController) GetAllOrders(c *gin.Context) {
	status := c.Param("status")
	userID := c.Param("user_id")

	orders, err := o.Service.GetAllOrders(status, userID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("[ADMIN] Fetch rejected structural mapping globally: %v", err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to fetch orders"})
		return
	}
	logger.Log.Info("[ADMIN] Fetched strictly unfiltered entire logic graph for mapping securely!")
	c.JSON(constant.SUCCESS, gin.H{
		"message": "Orders fetched successfully",
		"orders":  orders,
	})
}

func (o *OrderController) UpdateOrderStatusAdmin(c *gin.Context) {

	orderID := c.Param("order_id")
	if orderID == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error": "Order id required"})
		return
	}

	var req UpdateOrderStatusUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constant.BADREQUEST, validation.FormatValidationErrors(err))
		return
	}

	if err := o.Service.UpdateOrderStatusAdmin(orderID, req.Status); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("[ADMIN] Forced mechanical mapping status aborted completely: %v", err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to update order status"})
		return
	}
	logger.Log.Infof("[ADMIN] Safely hard-swapped logistics arrays successfully tracking Status %s globally on Order %s", req.Status, orderID)
	c.JSON(constant.SUCCESS, gin.H{"message": "Updated order status successfully"})
}