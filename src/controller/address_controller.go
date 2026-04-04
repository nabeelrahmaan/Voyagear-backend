package controller

import (
	"voyagear/src/models"
	"voyagear/src/services"
	"voyagear/utils/apperror"
	"voyagear/utils/constant"
	"voyagear/utils/logger"
	"voyagear/utils/validation"

	"github.com/gin-gonic/gin"
)

type AddressController struct {
	Service *services.AddressService
}

func SetupAddressController(service *services.AddressService) *AddressController {
	return &AddressController{
		Service: service,
	}
}

func (a *AddressController) CreateAddress(c *gin.Context) {
	var req models.UserAddress
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constant.BADREQUEST, validation.FormatValidationErrors(err))
		return
	}

	_, exist := c.Get("user_id")
	if !exist {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error": "User id not found"})
		return
	}

	if err := a.Service.CreateAddress(&req); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("Central logic rejected physical mapping payload creation: %v", err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to create address"})
		return
	}
	logger.Log.Info("Shipping address block dynamically locked strictly into memory natively.")
	c.JSON(constant.SUCCESS, gin.H{"message": "Address created"})
}

func (a *AddressController) GetAddresses(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error": "User id not found"})
		return
	}

	addresses, err := a.Service.GetUserAddresses(userID.(string))
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("Central logic failed executing database parameters: %v", err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to fetch address"})
		return
	}
	logger.Log.Infof("Returned strict list of shipping identifiers perfectly for User %s", userID.(string))
	c.JSON(constant.SUCCESS, gin.H{
		"message":   "Address fetched successfully",
		"addresses": addresses,
	})
}

func (a *AddressController) UpdateAddress(c *gin.Context) {
	addressID := c.Param("id")
	if addressID == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error": "Address id is required"})
		return
	}

	var fields map[string]interface{}
	if err := c.ShouldBindJSON(&fields); err != nil {
		c.JSON(constant.BADREQUEST, validation.FormatValidationErrors(err))
		return
	}

	if err := a.Service.UpdateAddress(addressID, fields); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("Address memory update totally failed structurally: %v", err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to update address"})
		return
	}
	logger.Log.Infof("Structurally overrode payload fields perfectly for Address %s", addressID)
	c.JSON(constant.SUCCESS, gin.H{"message": "Address updated successfully"})
}

func (a *AddressController) DeleteAddress(c *gin.Context) {
	addressID := c.Param("id")
	if addressID == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error": "Address id is required"})
		return
	}

	if err := a.Service.DeleteAddress(addressID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		logger.Log.Errorf("Internal parameter rejection structurally erasing Address %s: %v", addressID, err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to delete address"})
		return
	}
	logger.Log.Infof("Dropped memory completely entirely mapping Address %s perfectly", addressID)
	c.JSON(constant.SUCCESS, gin.H{"message": "Address deleted successfully"})
}