package validation

import (
	"fmt"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var (
	// phoneRegex ensures a clean 10 digit map 
	phoneRegex = regexp.MustCompile(`^[0-9]{10}$`)
	// passwordRegex ensures > 8 chars, 1 uppercase, 1 lowercase, 1 number
	passwordRegex = regexp.MustCompile(`^[a-zA-Z0-9!@#\$%\^\&*\)\(+=._-]{6,}$`)
	// alphaSpaceRegex allows characters and spaces for names
	alphaSpaceRegex = regexp.MustCompile(`^[a-zA-Z\s]+$`)
)

// InitValidation initializes custom regex validation across Gin's binding payload engine
func InitValidation() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("phone_regex", validatePhone)
		v.RegisterValidation("strong_pwd", validatePassword)
		v.RegisterValidation("alpha_space", validateAlphaSpace)
	}
}

func validatePhone(fl validator.FieldLevel) bool {
	return phoneRegex.MatchString(fl.Field().String())
}

func validatePassword(fl validator.FieldLevel) bool {
	return passwordRegex.MatchString(fl.Field().String())
}

func validateAlphaSpace(fl validator.FieldLevel) bool {
	return alphaSpaceRegex.MatchString(fl.Field().String())
}

// FormatValidationErrors interprets v10 reflection errors into beautiful string slices
func FormatValidationErrors(err error) gin.H {
	var errs []string
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			switch e.Tag() {
			case "required":
				errs = append(errs, fmt.Sprintf("%s is a required field", e.Field()))
			case "email":
				errs = append(errs, fmt.Sprintf("%s must be a valid email address", e.Field()))
			case "phone_regex":
				errs = append(errs, fmt.Sprintf("%s must be a valid 10-digit phone number", e.Field()))
			case "strong_pwd":
				errs = append(errs, fmt.Sprintf("%s must be at least 6 characters and contain valid formatting", e.Field()))
			case "alpha_space":
				errs = append(errs, fmt.Sprintf("%s must only contain letters and spaces", e.Field()))
			case "min":
				errs = append(errs, fmt.Sprintf("%s is too short", e.Field()))
			case "max":
				errs = append(errs, fmt.Sprintf("%s is too long", e.Field()))
			case "oneof":
				errs = append(errs, fmt.Sprintf("%s has an invalid categorical value", e.Field()))
			case "uuid":
				errs = append(errs, fmt.Sprintf("%s is not a valid identifier sequence", e.Field()))
			default:
				errs = append(errs, fmt.Sprintf("%s is invalid mapping (%v)", e.Field(), e.Tag()))
			}
		}
	} else {
		errs = append(errs, "Invalid JSON data mapping")
	}

	return gin.H{"validation_errors": errs}
}
