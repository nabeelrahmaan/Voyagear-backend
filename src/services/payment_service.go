package services

import (
	"errors"
	"time"
	"voyagear/src/models"
	"voyagear/src/repository"
	"voyagear/utils/apperror"
	"voyagear/utils/constant"
	"voyagear/utils/razorpay"

	"gorm.io/gorm"
)

type PaymentService struct {
	Repo           repository.PgSQLRepository
	RazorpayClient *razorpay.RazorpayClient
}

func SetupPaymentService(repo repository.PgSQLRepository, rzpClient *razorpay.RazorpayClient) *PaymentService {
	return &PaymentService{
		Repo:           repo,
		RazorpayClient: rzpClient,
	}
}

// VerifyPayment checks the callback signature from Razorpay and updates the order status
func (s *PaymentService) VerifyPayment(razorpayOrderID, razorpayPaymentID, razorpaySignature string) error {
	isValid := s.RazorpayClient.VerifyPayment(razorpayOrderID, razorpayPaymentID, razorpaySignature)
	if !isValid {
		return apperror.New(constant.BADREQUEST, "Invalid payment signature. Payment failed validation.", nil)
	}
	var order models.Order
	if err := s.Repo.FindOneWhere(&order, "razorpay_order_id = ?", razorpayOrderID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.New(constant.NOTFOUND, "Order associated with this payment not found", err)
		}
		return apperror.New(constant.INTERNALSERVERERROR, "Failed to fetch order", err)
	}
	if err := s.Repo.UpdateByFields(&order, order.ID, map[string]interface{}{
		"payment_status":      constant.PaymentStatusPaid,
		"status":              constant.OrderStatusConfirmed,
		"razorpay_payment_id": razorpayPaymentID,
		"updated_at":          time.Now(),
	}); err != nil {
		return apperror.New(constant.INTERNALSERVERERROR, "Verified payment securely, but failed to update order status in database", err)
	}
	return nil
}

// CreatePayment restarts the payment process for an unpaid order
func (s *PaymentService) CreatePayment(userID, orderID string) (string, error) {

	var order models.Order
	if err := s.Repo.FindOneWhere(&order, "id = ? AND user_id = ?", orderID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", apperror.New(constant.NOTFOUND, "Order not found", err)
		}
		return "", apperror.New(constant.INTERNALSERVERERROR, "Failed to fetch order", err)
	}

	if order.PaymentStatus == constant.PaymentStatusPaid {
		return "", apperror.New(constant.BADREQUEST, "Order is already paid", nil)
	}

	rzpID, err := s.RazorpayClient.CreateOrder(float64(order.Total), order.ID.String())
	if err != nil {
		return "", apperror.New(constant.INTERNALSERVERERROR, "Handshake with Razorpay failed", err)
	}

	if err := s.Repo.UpdateByFields(&order, order.ID, map[string]interface{}{
		"razorpay_order_id": rzpID,
		"payment_status":    constant.PaymentStatusPending,
		"updated_at":        time.Now(),
	}); err != nil {
		return "", apperror.New(constant.INTERNALSERVERERROR, "Failed to update order with payment token", err)
	}

	return rzpID, nil
}

func (s *PaymentService) GetUserPayments(userID string) ([]models.Order, error) {

	var orders []models.Order
	if err := s.Repo.FindWhereWithPreload(&orders, "user_id = ?", []interface{}{userID}, "Items.Product"); err != nil {
		return nil, apperror.New(constant.INTERNALSERVERERROR, "Failed to fetch payments", err)
	}

	return orders, nil
}

func (s *PaymentService) GetUserPaymentByID(userID, orderID string) (*models.Order, error) {

	var order models.Order
	if err := s.Repo.FindOneWhere(&order, "id = ? AND user_id = ?", orderID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(constant.NOTFOUND, "Payment details not found", err)
		}
		return nil, apperror.New(constant.INTERNALSERVERERROR, "Failed to fetch payment details", err)
	}

	return &order, nil
}

func (s *PaymentService) CancelPayment(userID, orderID string) error {

	var order models.Order
	if err := s.Repo.FindOneWhere(&order, "id = ? AND user_id = ?", orderID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.New(constant.NOTFOUND, "Order not found", err)
		}
		return apperror.New(constant.INTERNALSERVERERROR, "Failed to fetch order", err)
	}

	if order.PaymentStatus == constant.PaymentStatusPaid {
		return apperror.New(constant.BADREQUEST, "Cannot cancel a paid transaction", nil)
	}

	if err := s.Repo.UpdateByFields(&order, order.ID, map[string]interface{}{
		"payment_status": constant.PaymentStatusFailed,
		"status":         constant.OrderStatusCancelled,
		"updated_at":     time.Now(),
	}); err != nil {
		return apperror.New(constant.INTERNALSERVERERROR, "Failed to cancel payment", err)
	}

	return nil
}

func (s *PaymentService) UpdatePaymentStatusAdmin(orderID, status string) error {

	var order models.Order
	if err := s.Repo.FindById(&order, orderID); err != nil {
		return apperror.New(constant.NOTFOUND, "Order not found", err)
	}

	if err := s.Repo.UpdateByFields(&order, orderID, map[string]interface{}{
		"payment_status": status,
		"updated_at":     time.Now(),
	}); err != nil {
		return apperror.New(constant.INTERNALSERVERERROR, "Failed to update payment status", err)
	}

	return nil
}

func (s *PaymentService) GetPaymentByIDAdmin(orderID string) (*models.Order, error) {

	var order models.Order
	if err := s.Repo.FindById(&order, orderID); err != nil {
		return nil, apperror.New(constant.NOTFOUND, "Order/Payment not found", err)
	}
	
	_ = s.Repo.FindByIDWithPreload(&order, order.ID, "Items.Product")
	return &order, nil
}
