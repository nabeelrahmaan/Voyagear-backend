package controller

import (
	"voyagear/src/models"
	"voyagear/src/services"
	"voyagear/utils/apperror"
	"voyagear/utils/constant"

	"github.com/gin-gonic/gin"
)

type AddressController struct {
	Service *services.AddressService
}

func SetupAddressController (service *services.AddressService) *AddressController {
	return &AddressController{
		Service: service,
	}
}

func (a *AddressController) CreateAddress(c *gin.Context) {
	var req models.UserAddress
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constant.BADREQUEST, gin.H{"error":"Invalid request body"})
		return
	}
	
	_, exist := c.Get("user_id")
	if !exist {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error":"User id not found"})
		return
	}

	if err := a.Service.CreateAddress(&req); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, appErr.Message)
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error":"Failed to create address"})
		return
	}

	c.JSON(constant.SUCCESS, gin.H{"message":"Address created"})
}

func (a *AddressController) GetAddresses(c *gin.Context) {

	userID, exist := c.Get("user_id")
	if !exist {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error":"User id not found"})
		return
	}

	addresses, err := a.Service.GetUserAddresses(userID.(string))
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, appErr.Message)
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error":"Failed to fetch address"})
		return
	}

	c.JSON(constant.SUCCESS, gin.H{
		"message":"Address fetched successfully",
		"addresses":addresses,
	})
}

func (a *AddressController) UpdateAddress(c *gin.Context) {

	addressID := c.Param("id")
	if addressID == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error":"Address id is required"})
		return
	}

	var fields map[string]interface{}
	if err := c.ShouldBindJSON(&fields); err != nil {
		c.JSON(constant.BADREQUEST, gin.H{"error":"Invalid request body"})
		return
	}

	if err := a.Service.UpdateAddress(addressID, fields); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, appErr.Message)
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error":"Failed to update address"})
		return
	}

	c.JSON(constant.SUCCESS, gin.H{"message":"Address updated successfully"})
}

func (a *AddressController) DeleteAddress(c *gin.Context) {

	addressID := c.Param("id")
	if addressID == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error":"Address id is required"})
		return
	}

	if err := a.Service.DeleteAddress(addressID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, appErr.Message)
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error":"Failed to delete address"})
		return
	}

	c.JSON(constant.SUCCESS, gin.H{"message":"Address deleted successfully"})
}