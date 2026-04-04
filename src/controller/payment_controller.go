package controller

import (
	"voyagear/src/services"
	"voyagear/utils/apperror"
	"voyagear/utils/constant"
	"voyagear/utils/logger"
	"voyagear/utils/validation"

	"github.com/gin-gonic/gin"
)

type PaymentController struct {
	Service *services.PaymentService
}

func SetupPaymentController(service *services.PaymentService) *PaymentController {
	return &PaymentController{
		Service: service,
	}
}

type VerifyPaymentRequest struct {
	RazorpayOrderID   string `json:"razorpay_order_id" validate:"required"`
	RazorpayPaymentID string `json:"razorpay_payment_id" validate:"required"`
	RazorpaySignature string `json:"razorpay_signature" validate:"required"`
}

type CreatePaymentRequest struct {
	OrderID string `json:"order_id" validate:"required"`
}

func (pc *PaymentController) VerifyPayment(c *gin.Context) {
	_, exist := c.Get("user_id")
	if !exist {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error": "User id not found in the context"})
		return
	}
	var req VerifyPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constant.BADREQUEST, gin.H{"error": "Invalid request body"})
		return
	}
	if req.RazorpayOrderID == "" || req.RazorpayPaymentID == "" || req.RazorpaySignature == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error": "Missing razorpay verification keys"})
		return
	}
	if err := pc.Service.VerifyPayment(req.RazorpayOrderID, req.RazorpayPaymentID, req.RazorpaySignature); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("Payment signature verification completely failed for Razorpay ID %s: %v", req.RazorpayOrderID, err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to verify transaction securely"})
		return
	}
	
	logger.Log.Infof("Razorpay payment successfully verified mathematically for Order %s", req.RazorpayOrderID)
	c.JSON(constant.SUCCESS, gin.H{"message": "Payment verified successfully, order confirmed!"})
}

func (pc *PaymentController) CreatePayment(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error": "User id not found in the context"})
		return
	}
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constant.BADREQUEST, validation.FormatValidationErrors(err))
		return
	}
	rzpID, err := pc.Service.CreatePayment(userID.(string), req.OrderID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("Failed to generate Razorpay Payment token for Order %s: %v", req.OrderID, err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to create payment"})
		return
	}
	logger.Log.Infof("Payment token created cleanly for Order %s with Razorpay ID %s", req.OrderID, rzpID)
	c.JSON(constant.CREATED, gin.H{
		"message":           "Payment initiated successfully",
		"razorpay_order_id": rzpID,
	})
}

func (pc *PaymentController) GetUserPayments(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error": "User id not found in the context"})
		return
	}
	payments, err := pc.Service.GetUserPayments(userID.(string))
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("Failed to fetch payments for User %s: %v", userID.(string), err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to fetch payments"})
		return
	}
	logger.Log.Infof("Successfully fetched all %d payments locally for User %s", len(payments), userID.(string))
	c.JSON(constant.SUCCESS, gin.H{
		"message":  "Payments successfully retrieved",
		"payments": payments,
	})
}

func (pc *PaymentController) GetUserPaymentByID(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error": "User id not found in the context"})
		return
	}
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error": "Payment (order) id is required"})
		return
	}
	payment, err := pc.Service.GetUserPaymentByID(userID.(string), orderID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("Failed explicitly to fetch isolated Payment details for Order %s: %v", orderID, err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to fetch payment details"})
		return
	}
	logger.Log.Infof("Successfully verified isolated Payment explicitly for User %s mapping Order %s", userID.(string), orderID)
	c.JSON(constant.SUCCESS, gin.H{
		"message": "Payment details successfully retrieved",
		"payment": payment,
	})
}

func (pc *PaymentController) CancelPayment(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error": "User id not found in the context"})
		return
	}
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error": "Payment (order) id is required"})
		return
	}
	if err := pc.Service.CancelPayment(userID.(string), orderID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("User explicitly failed attempting to cancel checkout for Order %s: %v", orderID, err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to cancel payment"})
		return
	}
	logger.Log.Infof("Checkout explicitly aborted and permanently cancelled for Order %s by User %s", orderID, userID.(string))
	c.JSON(constant.SUCCESS, gin.H{"message": "Payment cancelled successfully"})
}

type UpdatePaymentStatusReq struct {
	Status string `json:"status" validate:"required"`
}

func (pc *PaymentController) UpdatePaymentStatusAdmin(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error": "Payment (order) id is required"})
		return
	}
	var req UpdatePaymentStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constant.BADREQUEST, gin.H{"error": "Invalid request body"})
		return
	}
	if err := pc.Service.UpdatePaymentStatusAdmin(orderID, req.Status); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("[ADMIN] Failed to hard-override payment status for Order %s: %v", orderID, err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to update payment status"})
		return
	}
	logger.Log.Infof("[ADMIN] Distinct system override: Order %s Payment Status flipped natively to %s", orderID, req.Status)
	c.JSON(constant.SUCCESS, gin.H{"message": "Payment status updated successfully"})
}

func (pc *PaymentController) GetPaymentByIDAdmin(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error": "Payment (order) id is required"})
		return
	}
	payment, err := pc.Service.GetPaymentByIDAdmin(orderID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("[ADMIN] Failed extracting internal payload keys for Order %s: %v", orderID, err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to fetch payment details"})
		return
	}
	logger.Log.Infof("[ADMIN] Successfully audited structural keys mapped structurally for Order %s", orderID)
	c.JSON(constant.SUCCESS, gin.H{
		"message": "Payment details successfully retrieved",
		"payment": payment,
	})
}
