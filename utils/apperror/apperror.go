package apperror

type AppError struct {
	Code    int
	Message string
	Err     error
}

func (a *AppError) Error() string {
	return a.Message
}

func New (code int, message string, err error) *AppError {
	return &AppError{
		Code: code,
		Message: message,
		Err: err,
	}
}
