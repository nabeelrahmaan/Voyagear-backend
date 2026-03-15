package controller

import (
	"voyagear/src/services"
	"voyagear/utils/apperror"
	"voyagear/utils/constant"
	"voyagear/utils/jwt"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService *services.AuthService
	jwtManager  *jwt.JWTmanger
}

func NewAuthController(service *services.AuthService, manager *jwt.JWTmanger) *AuthController {
	return &AuthController{
		authService: service,
		jwtManager:  manager,
	}
}

// Request structs
type signupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type verifyOTPRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	Email       string `json:"email"`
	OTP         string `json:"otp"`
	NewPassword string `json:"new_password"`
}

type updateProfileRequest struct {
	Name string `json:"name"`
}


func (h *AuthController) Signup(c *gin.Context) {
	var req signupRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constant.BADREQUEST, gin.H{"error": constant.INVALID_REQ})
		return
	}

	err := h.authService.Signup(req.Name, req.Email, req.Password)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error":appErr.Message})
			return
		}
		
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error": err.Error()})
		return
	}

	c.JSON(constant.CREATED, gin.H{"message":"User created successfully. OTP sent to email"})
}

func (h *AuthController) VerifyOTP(c *gin.Context) {
	var req verifyOTPRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constant.BADREQUEST, gin.H{"error":constant.INVALID_REQ})
		return
	}

	err := h.authService.VerifyOTP(req.Email, req.OTP)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error":appErr.Message})
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error":err.Error()})
		return
	}

	c.JSON(constant.SUCCESS, gin.H{"message":"Account verified successfully"})
}

func (h *AuthController) Login(c *gin.Context){

	var req loginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constant.BADREQUEST, gin.H{"error":constant.INVALID_REQ})
		return
	}

	user, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error":err.Error()})
		return
	}

	accessToken, err := h.jwtManager.GenerateAccessToken(user.ID.String(), user.Role)
	if err != nil {
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error":"Failed to generate access token"})
		return
	}

	refreshToken, err := h.jwtManager.GenerateRefreshToken(user.ID.String(), user.Role)
	if err != nil {
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error":"Failed to generate refresh token"})
		return
	}

	c.SetCookie(
		"access_token",
		accessToken,
		900,
		"/",
		"localhost",
		false,
		true,
	)

	c.SetCookie(
		"refresh_token",
		refreshToken,
		604800,
		"/",
		"localhost",
		false,
		true,
	)

	c.JSON(constant.SUCCESS, gin.H{
		"access_token":accessToken,
		"refresh_token":refreshToken,
	})
}

func (h *AuthController) RefreshToken(c *gin.Context){

	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error":"Refresh token is missing"})
		return
	}

	claims, err := h.jwtManager.ValidateRefresh(refreshToken)
	if err != nil {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error":"Invalid token"})
		return
	}

	userID, ok := claims["user_id"].(string)
	if userID == "" || !ok {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error":"Invalid token claims"})
		return
	}

	role, ok := claims["role"].(string)
	if !ok || role == "" {
		role = "user"
	}

	newAccess, err := h.jwtManager.GenerateAccessToken(userID, role)
	if err != nil {
		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error":"Failed to generate new access token"})
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

	c.JSON(constant.SUCCESS, gin.H{"access_token":newAccess})
}

func (h *AuthController) ForgotPassword(c *gin.Context) {
	var req forgotPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constant.BADREQUEST, gin.H{"error":constant.INVALID_REQ})
		return
	}

	err := h.authService.ForgotPassword(req.Email)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, gin.H{"error":appErr.Message})
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error":err.Error()})
		return
	}

	c.JSON(constant.SUCCESS, gin.H{"message":"OTP sent to the email"})
}

func (h *AuthController) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constant.BADREQUEST, gin.H{"error":constant.INVALID_REQ})
		return
	}

	if req.Email == "" || req.OTP == "" || req.NewPassword == "" {
		c.JSON(constant.BADREQUEST, gin.H{"error":"All fields are required"})
		return
	}

	err := h.authService.ResetPassword(req.Email, req.OTP, req.NewPassword)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, appErr.Message)
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error":err.Error()})
		return
	}

	c.JSON(constant.SUCCESS, gin.H{"message":"Password changed successfully"})
}

func (h *AuthController) GetProfile(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error":constant.UN_AUTH})
		return
	}

	user, err := h.authService.GetProfile(userID.(string))
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, appErr.Message)
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error":err.Error()})
		return
	}

	c.JSON(constant.SUCCESS, user)
}

func (h *AuthController) UpdateProfile(c *gin.Context) {

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(constant.UNAUTHORIZED, gin.H{"error":constant.UN_AUTH})
		return
	}

	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(constant.BADREQUEST, gin.H{"error":constant.INVALID_REQ})
		return
	}

	if len(req.Name) < 3 {
		c.JSON(constant.BADREQUEST, gin.H{"error":"Name is required"})
		return
	}

	user, err := h.authService.UpdateProfile(userID.(string), req.Name)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			c.JSON(appErr.Code, appErr.Message)
			return
		}

		c.JSON(constant.INTERNALSERVERERROR, gin.H{"error":err.Error()})
		return
	}

	c.JSON(constant.SUCCESS, user)
}


