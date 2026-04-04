package controller

import (
	"voyagear/src/services"
	"voyagear/utils/apperror"
	"voyagear/utils/constant"
	"voyagear/utils/logger"
	"voyagear/utils/validation"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService *services.AuthService
}

func NewAuthController(service *services.AuthService) *AuthController {
	return &AuthController{
		authService: service,
	}
}

// Request structs
type signupRequest struct {
	Name     string `json:"name" validate:"required,min=3,max=50,alpha_space"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,strong_pwd"`
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type verifyOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp" validate:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type forgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type resetPasswordRequest struct {
	Email       string `json:"email" validate:"required,email"`
	OTP         string `json:"otp" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6,strong_pwd"`
}

type updateProfileRequest struct {
	Name string `json:"name" validate:"required,min=3,max=50,alpha_space"`
}

func (h *AuthController) Test(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Server working successfully"})
}

func (h *AuthController) Signup(c *gin.Context) {
	var req signupRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constant.BADREQUEST, validation.FormatValidationErrors(err))
		return
	}

	err := h.authService.Signup(req.Name, req.Email, req.Password)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}

		logger.Log.Errorf("User signup failed internally: %v", err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": err.Error()})
		return
	}

	logger.Log.Infof("New user registered successfully with email: %s", req.Email)
	c.JSON(constant.CREATED, gin.H{"message": "User created successfully. OTP sent to email"})
}

func (h *AuthController) VerifyOTP(c *gin.Context) {
	var req verifyOTPRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constant.BADREQUEST, validation.FormatValidationErrors(err))
		return
	}

	err := h.authService.VerifyOTP(req.Email, req.OTP)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constant.SUCCESS, gin.H{"message": "Account verified successfully"})
}

func (h *AuthController) Login(c *gin.Context) {

	var req loginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constant.BADREQUEST, validation.FormatValidationErrors(err))
		return
	}

	access, refresh, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}

		logger.Log.Errorf("User login failed internally for %s: %v", req.Email, err)
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": err.Error()})
		return
	}

	logger.Log.Infof("User logged in successfully: %s", req.Email)

	// accessToken, err := h.jwtManager.GenerateAccessToken(user.ID.String(), user.Role)
	// if err != nil {
	// 	c.JSON(constant.INTERNALSERVERERROR, gin.H{"error":"Failed to generate access token"})
	// 	return
	// }

	// refreshToken, err := h.jwtManager.GenerateRefreshToken(user.ID.String(), user.Role)
	// if err != nil {
	// 	c.JSON(constant.INTERNALSERVERERROR, gin.H{"error":"Failed to generate refresh token"})
	// 	return
	// }

	c.SetCookie(
		"access_token",
		access,
		900,
		"/",
		"localhost",
		false,
		true,
	)

	c.SetCookie(
		"refresh_token",
		refresh,
		604800,
		"/",
		"localhost",
		false,
		true,
	)

	c.JSON(constant.SUCCESS, gin.H{
		"access_token":  access,
		"refresh_token": refresh,
	})
}

func (h *AuthController) RefreshToken(c *gin.Context) {

	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error": "Refresh token is missing"})
		return
	}

	// claims, err := h.jwtManager.ValidateRefresh(refreshToken)
	// if err != nil {
	// 	c.JSON(constant.UNAUTHORIZED, gin.H{"error":"Invalid token"})
	// 	return
	// }

	// userID, ok := claims["user_id"].(string)
	// if userID == "" || !ok {
	// 	c.JSON(constant.UNAUTHORIZED, gin.H{"error":"Invalid token claims"})
	// 	return
	// }

	// role, ok := claims["role"].(string)
	// if !ok || role == "" {
	// 	role = "user"
	// }

	// newAccess, err := h.jwtManager.GenerateAccessToken(userID, role)
	// if err != nil {
	// 	c.JSON(constant.INTERNALSERVERERROR, gin.H{"error":"Failed to generate new access token"})
	// 	return
	// }

	newAccess, newRefresh, err := h.authService.Refresh(refreshToken)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, appErr.Message)
			return
		}

		c.JSON(constant.UNAUTHORIZED, gin.H{"error": "Invalid token"})
		return
	}

	c.SetCookie(
		"access_token",
		newAccess,
		900,
		"/",
		"localhost",
		false,
		true,
	)

	c.SetCookie(
		"refresh_token",
		newRefresh,
		604800,
		"/",
		"localhost",
		false,
		true,
	)

	c.JSON(constant.SUCCESS, gin.H{
		"access_token":  newAccess,
		"refresh_token": newRefresh,
	})
}

func (h *AuthController) ForgotPassword(c *gin.Context) {
	var req forgotPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constant.BADREQUEST, validation.FormatValidationErrors(err))
		return
	}

	err := h.authService.ForgotPassword(req.Email)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constant.SUCCESS, gin.H{"message": "OTP sent to the email"})
}

func (h *AuthController) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constant.BADREQUEST, validation.FormatValidationErrors(err))
		return
	}

	if req.Email == "" || req.OTP == "" || req.NewPassword == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error": "All fields are required"})
		return
	}

	err := h.authService.ResetPassword(req.Email, req.OTP, req.NewPassword)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, appErr.Message)
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constant.SUCCESS, gin.H{"message": "Password changed successfully"})
}

func (h *AuthController) GetProfile(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error": constant.UN_AUTH})
		return
	}

	user, err := h.authService.GetProfile(userID.(string))
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, appErr.Message)
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constant.SUCCESS, user)
}

func (h *AuthController) UpdateProfile(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error": constant.UN_AUTH})
		return
	}

	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constant.BADREQUEST, validation.FormatValidationErrors(err))
		return
	}

	if len(req.Name) < 3 {
		c.JSON(constant.BADREQUEST, gin.H{"error": "Name is required"})
		return
	}

	user, err := h.authService.UpdateProfile(userID.(string), req.Name)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, appErr.Message)
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constant.SUCCESS, user)
}

func (h *AuthController) GetAllUsers(c *gin.Context) {

	users, err := h.authService.GetAllUsers()
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, appErr.Message)
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constant.SUCCESS, users)
}

func (h *AuthController) ToggleISBlock(c *gin.Context) {

	userID := c.Param("id")
	if userID == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error": "User not found"})
		return
	}

	user, err := h.authService.ToggleIsBlock(userID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, appErr.Message)
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(constant.SUCCESS, user)
}

func (h *AuthController) DeleteUserById(c *gin.Context) {

	userID := c.Param("id")
	if userID == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error": "Invalid ID"})
		return
	}

	err := h.authService.DeleteUserByID(userID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, appErr.Message)
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(constant.SUCCESS, gin.H{"message": "User deleted successfully"})
}

func (h *AuthController) Logout(c *gin.Context) {

	token, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error":"Invalid token"})
		return
	}
	
	if err := h.authService.Logout(token); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, appErr.Message)
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error":"Failed to logout user"})
		return
	}

	c.SetCookie(
		"access_token",
		"",
		-1,
		"/",
		"localhost",
		false,
		true,
	)

	c.SetCookie(
		"refresh_token",
		"",
		-1,
		"/",
		"localhost",
		false,
		true,
	)

	c.JSON(constant.SUCCESS, gin.H{"message":"Logged out successfully"})
}
