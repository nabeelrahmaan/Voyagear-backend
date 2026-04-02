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
	INVALID_REQ    = "Invalid request body"
	UN_AUTH        = "Unauthorized"
	INTERNAL_ERROR = "Internal error"

	// Order status
	OrderStatusPending   = "pending"
	OrderStatusConfirmed = "confirmed"
	OrderStatusShipped   = "shipped"
	OrderStatusDelivered = "delivered"
	OrderStatusCancelled = "cancelled"

	// payment methods
	PaymentMethodsCOD     = "cod"
	PaymentMethodRazorpay = "razorpay"

	//payment status
	PaymentStatusPending  = "pending"
	PaymentStatusPaid     = "paid"
	PaymentStatusFailed   = "failed"
	PaymentStatusRefunded = "refunded"
)

// const (
// 	//order status
// 	OrderStatusPending   OrderStatus = "pending"
// 	OrderStatusConfirmed OrderStatus = "confirmed"
// 	OrderStatusShipped   OrderStatus = "shipped"
// 	OrderStatusDelivered OrderStatus = "delivered"
// 	OrderStatusCancelled OrderStatus = "cancelled"

// 	// payment methods
// 	PaymentMethodsCOD     PaymentMethod = "cod"
// 	PaymentMethodRazorpay PaymentMethod = "razorpay"

// 	//payment status
// 	PaymentStatusPending  PaymentStatus = "pending"
// 	PaymentStatusPaid     PaymentStatus = "paid"
// 	PaymentStatusFailed   PaymentStatus = "failed"
// 	PaymentStatusRefunded PaymentStatus = "refunded"
// )
