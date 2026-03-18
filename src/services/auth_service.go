package services

import (
	"time"
	"voyagear/internal/cache"
	"voyagear/src/models"
	"voyagear/src/repository"
	"voyagear/utils/apperror"
	"voyagear/utils/constant"
	"voyagear/utils/email"
	"voyagear/utils/jwt"
	"voyagear/utils/otp"
	passwords "voyagear/utils/password" // check this if something went wrong

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type AuthService struct {
	Repo       repository.PgSQLRepository
	Redis      *cache.Redis
	JwtManager *jwt.JWTmanger
}

func CreateAuthService(repo repository.PgSQLRepository, redis *cache.Redis, jwt *jwt.JWTmanger) *AuthService {
	return &AuthService{
		Repo:       repo,
		Redis:      redis,
		JwtManager: jwt,
	}
}

func (s *AuthService) Signup(name, useremail, password string) error {
	var exist *models.User

	// Check user already exist
	if err := s.Repo.FindOneWhere(&exist, "email = ?", useremail); err == nil {
		return apperror.New(
			constant.BADREQUEST,
			"Email already exist",
			err,
		)
	}

	hashpass, err := passwords.HashPassword(password)
	if err != nil {
		return err
	}

	user := models.User{
		Name:       name,
		Email:      useremail,
		Password:   hashpass,
		IsVerified: false,
	}

	// Inserting new user into database
	if err := s.Repo.Insert(&user); err != nil {
		return apperror.New(
			constant.INTERNALSERVERERROR,
			"Failed to create user",
			err,
		)
	}

	// Generating otp for verification
	newotp := otp.GenerateOTP()

	otphash, err := passwords.HashPassword(newotp)
	if err != nil {
		return err
	}

	key := "otp:verify:" + useremail

	// store hashed otp on redis
	s.Redis.Client.Set(
		cache.Ctx,
		key,
		otphash,
		5*time.Minute,
	)

	// Sending otp into user's email
	if err := email.SendOTP(useremail, newotp); err != nil {
		return apperror.New(
			constant.INTERNALSERVERERROR,
			"Failed to send OTP",
			err,
		)
	}
	return nil
}

// VerifyOTP verify user's OTP
func (s *AuthService) VerifyOTP(useremail, otp string) error {

	var user models.User

	// Check is user exist
	if err := s.Repo.FindOneWhere(&user, "email = ? ", useremail); err != nil {
		return apperror.New(
			constant.NOTFOUND,
			"User not found",
			err,
		)
	}

	// Check is user already verified
	if user.IsVerified {
		return apperror.New(
			constant.BADREQUEST,
			"User already verified",
			nil,
		)
	}

	key := "otp:verify:" + useremail

	storedOTP, err := s.Redis.Client.Get(cache.Ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return apperror.New(
				constant.BADREQUEST,
				"OTP expired",
				err,
			)
		} else {
			return err
		}
	}

	// Checking otp is correct by comparing hash values
	if !passwords.CheckPassword(otp, storedOTP) {
		return apperror.New(
			constant.UNAUTHORIZED,
			"Invalid OTP",
			err,
		)
	}

	updates := map[string]interface{}{
		"is_verified": true,
	}

	// Updating user(IsVerified)
	if err := s.Repo.UpdateByFields(&models.User{}, user.ID, updates); err != nil {
		return err
	}

	s.Redis.Client.Del(cache.Ctx, key)
	return nil
}

func (s *AuthService) Login(email, password string) (string, string, error) {
	var user models.User

	// Find user by email
	if err := s.Repo.FindOneWhere(&user, "email = ? ", email); err != nil {
		return "", "", apperror.New(
			constant.UNAUTHORIZED,
			"Invalid credentials",
			err,
		)
	}

	// Check if user is verified
	if !user.IsVerified {
		return "", "", apperror.New(
			constant.UNAUTHORIZED,
			"User not verified",
			nil,
		)
	}

	// check if user is blocked
	if user.IsBlocked {
		return "", "", apperror.New(
			constant.FORBIDDEN,
			"Your account has been blocked",
			nil,
		)
	}

	// Verifying password
	if passwords.CheckPassword(password, user.Password) {
		return "", "", apperror.New(
			constant.UNAUTHORIZED,
			"Invalid credentials",
			nil,
		)
	}

	sessionID := uuid.New()

	// Generate new access, refresh tokens
	accessToken, err := s.JwtManager.GenerateAccessToken(user.ID.String(), user.Role)
	if err != nil {
		return "", "", apperror.New(
			constant.INTERNALSERVERERROR,
			"Failed to create access token",
			err,
		)
	}

	refreshToken, err := s.JwtManager.GenerateRefreshToken(user.ID.String(), user.Role, sessionID.String())
	if err != nil {
		return "", "", apperror.New(
			constant.INTERNALSERVERERROR,
			"Failed to create refresh token",
			err,
		)
	}

	hashToken := passwords.HashToken(refreshToken)

	refresh := models.RefreshToken{
		ID:     sessionID,
		UserID: user.ID,
		Token:  hashToken,
	}

	// Storing refresh token into database
	if err := s.Repo.Insert(&refresh); err != nil {
		return "", "", apperror.New(
			constant.INTERNALSERVERERROR,
			err.Error(),
			err,
		)
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) ForgotPassword(useremail string) error {

	var user models.User

	// Find user by email
	if err := s.Repo.FindOneWhere(&user, "email = ? ", useremail); err != nil {
		return nil // Return nil for security(prevent user enumeration attacks)
	}

	newotp := otp.GenerateOTP()

	hashotp, err := passwords.HashPassword(newotp)
	if err != nil {
		return err
	}

	key := "otp:newpass:" + useremail

	// Storing hashed otp on redis
	s.Redis.Client.Set(
		cache.Ctx,
		key,
		hashotp,
		5*time.Minute,
	)

	// Sending OTP email to user
	if err := email.SendOTP(useremail, newotp); err != nil {
		apperror.New(
			constant.INTERNALSERVERERROR,
			"Failed to send OTP",
			err,
		)
	}

	return nil
}

func (s *AuthService) ResetPassword(email, otp, password string) error {

	var user models.User

	// Find user by email
	if err := s.Repo.FindOneWhere(&user, "email = ? ", email); err != nil {
		return apperror.New(
			constant.UNAUTHORIZED,
			"Invalid credentials",
			err,
		)
	}

	key := "otp:newpass:" + email

	// Getting OTP from redis
	hashotp, err := s.Redis.Client.Get(cache.Ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return apperror.New(
				constant.BADREQUEST,
				"OTP expired",
				err,
			)
		} else {
			return err
		}
	}

	// Verifying OTP
	if !passwords.CheckPassword(otp, hashotp) {
		return apperror.New(
			constant.UNAUTHORIZED,
			"Invalid OTP",
			nil,
		)
	}

	newPassword, err := passwords.HashPassword(password)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"password": newPassword,
	}

	// Updating user with new password
	if err := s.Repo.UpdateByFields(&models.User{}, user.ID, updates); err != nil {
		return apperror.New(
			constant.INTERNALSERVERERROR,
			"Failed to update password",
			err,
		)
	}

	s.Redis.Client.Del(cache.Ctx, key)
	return nil
}

func (s *AuthService) GetProfile(userID string) (*models.User, error) {
	var user models.User

	if err := s.Repo.FindById(&user, userID); err != nil {
		return nil, apperror.New(
			constant.NOTFOUND,
			"User not found",
			err,
		)
	}

	return &user, nil
}

func (s *AuthService) UpdateProfile(userID, newname string) (*models.User, error) {
	var user models.User

	if err := s.Repo.FindById(&user, userID); err != nil {
		return nil, apperror.New(
			constant.NOTFOUND,
			"User not found",
			err,
		)
	}

	// Storing updates in to map
	updates := map[string]interface{}{
		"name": newname,
	}

	if err := s.Repo.UpdateByFields(&models.User{}, userID, updates); err != nil {
		return nil, apperror.New(
			constant.INTERNALSERVERERROR,
			"Failed to update user",
			err,
		)
	}

	if err := s.Repo.FindById(&user, userID); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *AuthService) ToggleIsBlock(userID string) (*models.User, error) {

	var user models.User

	if err := s.Repo.FindById(&user, userID); err != nil {
		return nil, apperror.New(
			constant.NOTFOUND,
			"User not found",
			err,
		)
	}

	status := !user.IsBlocked

	updates := map[string]interface{}{
		"is_blocked": status,
	}

	if err := s.Repo.UpdateByFields(&models.User{}, userID, updates); err != nil {
		return nil, apperror.New(
			constant.INTERNALSERVERERROR,
			"Failed to update user",
			err,
		)
	}

	if err := s.Repo.FindById(&user, userID); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *AuthService) GetById(userID string) (*models.User, error) {

	var user models.User

	if err := s.Repo.FindById(&user, userID); err != nil {
		return nil, apperror.New(
			constant.NOTFOUND,
			"User not found",
			err,
		)
	}

	return &user, nil
}

func (s *AuthService) GetAllUsers() ([]models.User, error) {

	var users []models.User

	if err := s.Repo.FindAll(&users); err != nil {
		return nil, apperror.New(
			constant.INTERNALSERVERERROR,
			"Failed to fetch users",
			err,
		)
	}

	return users, nil
}

func (s *AuthService) DeleteUserByID(userID string) error {

	var user models.User

	if err := s.Repo.FindById(&user, userID); err != nil {
		return apperror.New(
			constant.NOTFOUND,
			"User not found",
			err,
		)
	}

	// Avoiding deletion of admin users
	if user.Role == "admin" {
		return apperror.New(
			constant.FORBIDDEN,
			"Admin users can't be deleted",
			nil,
		)
	}

	if err := s.Repo.Delete(&user, userID); err != nil {
		return apperror.New(
			constant.INTERNALSERVERERROR,
			"Failed to delete user",
			err,
		)
	}

	return nil
}

func (s *AuthService) Refresh(token string) (string, string, error) {

	// Validating refresh token
	claims, err := s.JwtManager.ValidateRefresh(token)
	if err != nil {
		return "", "", apperror.New(
			constant.UNAUTHORIZED,
			"Invalid token",
			err,
		)
	}

	// Slicing datas from claims
	sessionID, ok := claims["session_id"].(string)
	if !ok {
		return "", "", apperror.New(
			constant.UNAUTHORIZED,
			"Session id is missing",
			nil,
		)
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", "", apperror.New(
			constant.UNAUTHORIZED,
			"User not found",
			nil,
		)
	}

	role, ok := claims["role"].(string)
	if !ok {
		role = "user"
	}

	// Finding refresh token from database
	var userToken models.RefreshToken
	if err := s.Repo.FindOneWhere(&userToken, "id = ?", sessionID); err != nil {
		return "", "", apperror.New(
			constant.NOTFOUND,
			"User not found",
			err,
		)
	}

	// Checking token strings from databse and request
	if !passwords.CompareTokens(token, userToken.Token) {
		return "", "", apperror.New(
			constant.UNAUTHORIZED,
			"Invalid token",
			err,
		)
	}

	// Validating max session time
	if time.Since(userToken.CreatedAt) > s.JwtManager.MaxSession {
		return "", "", apperror.New(
			constant.UNAUTHORIZED,
			"Session expired",
			nil,
		)
	}

	// Generating new access and refresh tokens
	accessToken, err := s.JwtManager.GenerateAccessToken(userID, role)
	if err != nil {
		return "", "", apperror.New(
			constant.INTERNALSERVERERROR,
			"Failed to create access token",
			err,
		)
	}

	refreshToken, err := s.JwtManager.GenerateRefreshToken(userID, role, sessionID)
	if err != nil {
		return "", "", apperror.New(
			constant.INTERNALSERVERERROR,
			"Failed to create refresh token",
			err,
		)
	}

	// Updating current refresh token string with new one in database
	hashToken := passwords.HashToken(refreshToken)

	updates := map[string]interface{}{
		"token": hashToken,
	}

	if err := s.Repo.UpdateByFields(&models.RefreshToken{}, userID, updates); err != nil {
		return "", "", apperror.New(
			constant.INTERNALSERVERERROR,
			"Failed to store refresh token",
			err,
		)
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) Logout(token string) error {

	claims, err := s.JwtManager.ValidateRefresh(token)
	if err != nil {
		return apperror.New(
			constant.UNAUTHORIZED,
			"Invalid token",
			err,
		)
	}

	sessionID, ok := claims["session_id"].(string)
	if !ok {
		return apperror.New(
			constant.UNAUTHORIZED,
			"Session id not found",
			nil,
		)
	}

	if err := s.Repo.Delete(&models.RefreshToken{}, sessionID); err != nil {
		return apperror.New(
			constant.INTERNALSERVERERROR,
			"Failed to delete token",
			err,
		)
	}

	return nil
}
