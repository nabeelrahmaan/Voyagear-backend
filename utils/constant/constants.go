package constant

var (
	//HTTP status codes
	SUCCESS             = 200
	CREATED             = 201
	BADREQUEST          = 400
	UNAUTHORIZED        = 401
	FORBIDDEN           = 403
	NOTFOUND            = 404
	CONFLICT            = 409
	INTERNALSERVERERROR = 500

	// Generic status strings
	PENDING        = "Pending"
	INVALID_REQ    = "Invalid request body"
	UN_AUTH  = "Unauthorized"
	INTERNAL_ERROR = "INTERNAL_ERROR"
	PAID           = "PAID"
	FAILED         = "FAILED"

	// Payment or process
	PROCESSING = "PROCESSING"
	CANCELLED  = "CANCELLED"
	PLACED     = "PLACED"
	SHIPPED    = "SHIPPED"
	DELIVERED  = "DELIVERED"
)
