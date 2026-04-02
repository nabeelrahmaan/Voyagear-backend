package services

import (
	"errors"
	"time"
	"voyagear/src/models"
	"voyagear/src/repository"
	"voyagear/utils/apperror"
	"voyagear/utils/constant"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderService struct {
	Repo repository.PgSQLRepository
}

func SetupOrderService(repo repository.PgSQLRepository) *OrderService {
	return &OrderService{
		Repo: repo,
	}
}

type PlaceOrderRequest struct {
	Type          string `json:"type" validate:"required,oneof=cart direct"`
	ProductID     string `json:"product_id"`
	Quantity      int    `json:"quantity"`
	PaymentMethod string `json:"payment_method"`
	Size          string `json:"size"`
	AddressID     string `json:"address_id" validate:"required,uuid"`
}

func (s *OrderService) GetUserOrders(userID string) ([]models.Order, error) {

	uId, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperror.New(
			constant.BADREQUEST,
			"Invalid user id",
			err,
		)
	}

	var orders []models.Order
	if err := s.Repo.FindWhereWithPreload(&orders, "user_id = ?", []interface{}{uId}, "Items.Product"); err != nil {
		return nil, apperror.New(
			constant.INTERNALSERVERERROR,
			"Failed to fetch orders",
			err,
		)
	}

	return orders, nil
}

func (s *OrderService) PlaceOrder(userID string, req PlaceOrderRequest) (order *models.Order, err error) {

	uId, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperror.New(
			constant.BADREQUEST,
			"Invalid user id",
			err,
		)
	}

	addressID, err := uuid.Parse(req.AddressID)
	if err != nil {
		return nil, apperror.New(
			constant.BADREQUEST,
			"Invalid address id",
			err,
		)
	}

	var address models.UserAddress
	if err := s.Repo.FindOneWhere(&address, "id = ? AND user_id = ?", addressID, uId); err != nil {
		return nil, apperror.New(
			constant.NOTFOUND,
			"Address not found",
			err,
		)
	}

	// Begin transaction
	tx := s.Repo.Begin()
	if tx.Error != nil {
		return nil, apperror.New(constant.INTERNALSERVERERROR, "Failed to start transaction", nil)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		} else if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	var orderItems []models.OrderItem
	var total int

	switch req.Type {
	case "direct":
		if req.Size == "" {
			return nil, apperror.New(constant.BADREQUEST, "Size is required for direct orders", nil)
		}
		if req.ProductID == "" {
			return nil, apperror.New(constant.BADREQUEST, "Product ID is required for direct orders", nil)
		}
		if req.Quantity <= 0 {
			return nil, apperror.New(constant.BADREQUEST, "Quantity must greater than 0", nil)
		}

		orderItems, total, err = s.ProcessDirectOrders(tx, req)
		if err != nil {
			return nil, err
		}

	case "cart":
		orderItems, total, err = s.ProcessCartOrder(tx, uId)
		if err != nil {
			return nil, err
		}

	default:
		return nil, apperror.New(constant.BADREQUEST, "Invalid order type", nil)
	}

	newOrder := models.Order{
		UserID:    uId,
		Total:     total,
		Status:    constant.OrderStatusConfirmed,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err = tx.Create(&newOrder).Error; err != nil {
		return nil, apperror.New(constant.INTERNALSERVERERROR, "Failed to create order", err)
	}

	for _, item := range orderItems {
		item.OrderID = newOrder.ID
		item.CreatedAt = time.Now()
		item.UpdatedAt = time.Now()

		if err = tx.Create(&item).Error; err != nil {
			return nil, apperror.New(constant.INTERNALSERVERERROR, "Failed to create order item", err)
		}
	}

	var fullOrder models.Order
	if err := s.Repo.FindByIDWithPreload(&fullOrder, newOrder.ID, "Items.Product"); err != nil {
		return nil, apperror.New(constant.INTERNALSERVERERROR, "Order created but failed to fetch details", err)
	}

	order = &fullOrder
	return
}

func (s *OrderService) ProcessDirectOrders(tx *gorm.DB, req PlaceOrderRequest) ([]models.OrderItem, int, error) {

	pID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return nil, 0, apperror.New(constant.BADREQUEST, "Invalid product id", err)
	}

	// Fetch product
	var product models.Product
	if err := tx.First(&product, "id = ?", pID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, apperror.New(constant.NOTFOUND, "Product not found", err)
		}

		return nil, 0, apperror.New(constant.INTERNALSERVERERROR, "Failed to fetch product", err)
	}

	// Check size for corresponding product
	var productVariant models.Variant
	if err := tx.Where("product_id = ? AND size = ?", pID, req.Size).First(&productVariant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, apperror.New(constant.NOTFOUND, "Size not available for this product", err)
		}

		return nil, 0, apperror.New(constant.INTERNALSERVERERROR, "Failed to check stock", err)
	}

	// Check stock available
	if productVariant.Quantity < req.Quantity {
		return nil, 0, apperror.New(constant.BADREQUEST, "Insufficient stock", nil)
	}

	// deduct quantity
	if err := tx.Model(&models.Variant{}).
		Where("id = ?", productVariant.ID).
		UpdateColumn("quantity", gorm.Expr("quantity - ?", req.Quantity)).
		UpdateColumn("updated_at", time.Now()).Error; err != nil {
		return nil, 0, apperror.New(constant.INTERNALSERVERERROR, "Failed to update quantity", err)
	}

	item := models.OrderItem{
		ProductID: pID,
		VariantID: productVariant.ID,
		Quantity:  req.Quantity,
		Price:     product.Price,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	total := product.Price * req.Quantity
	return []models.OrderItem{item}, total, nil
}

func (s *OrderService) ProcessCartOrder(tx *gorm.DB, userID uuid.UUID) ([]models.OrderItem, int, error) {

	// Fetching cart
	var cart models.Cart
	if err := tx.First(&cart, "user_id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, apperror.New(constant.NOTFOUND, "Cart not found", err)
		}

		return nil, 0, apperror.New(constant.INTERNALSERVERERROR, "Failed to fetch cart", err)
	}

	var cartItems []models.CartItem
	if err := tx.Preload("Product").Where("cart_id = ?", cart.ID).Find(&cartItems).Error; err != nil {
		return nil, 0, apperror.New(constant.INTERNALSERVERERROR, "Failed to fetch cart items", err)
	}

	// Check is cart empty
	if len(cartItems) == 0 {
		return nil, 0, apperror.New(constant.BADREQUEST, "Cart is empty", nil)
	}

	var orderItems []models.OrderItem
	total := 0

	for _, item := range cartItems {

		// Check and deduct stock for each cart item
		var productVariant models.Variant
		if err := tx.Where("id = ? AND product_id = ? ", item.VariantID, item.ProductID).First(&productVariant).Error; err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, 0, apperror.New(constant.NOTFOUND, "Size not available for this product", err)
			}

			return nil, 0, apperror.New(constant.INTERNALSERVERERROR, "Failed to check stock", err)
		}

		if productVariant.Quantity < item.Quantity {
			return nil, 0, apperror.New(constant.BADREQUEST, "Insufficient stock", nil)
		}

		// Deduct stock
		if err := tx.Model(&models.Variant{}).
			Where("id = ?", productVariant.ID).
			UpdateColumn("quantity", gorm.Expr("quantity - ?", item.Quantity)).
			UpdateColumn("updated_at", time.Now()).Error; err != nil {
			return nil, 0, apperror.New(constant.INTERNALSERVERERROR, "Failed to update quantity", err)
		}

		orderItems = append(orderItems, models.OrderItem{
			ProductID: item.ProductID,
			VariantID: productVariant.ID,
			Quantity:  item.Quantity,
			Price:     item.Product.Price,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})

		total += item.Product.Price * item.Quantity
	}

	// Clear cart
	if err := tx.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{}).Error; err != nil {
		return nil, 0, apperror.New(constant.INTERNALSERVERERROR, "Failed to clear cart", err)
	}

	return orderItems, total, nil
}

func (s *OrderService) GetOrderDetails(userID string, orderID string) (*models.Order, error) {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperror.New(constant.BADREQUEST, "Invalid user ID", err)
	}

	oID, err := uuid.Parse(orderID)
	if err != nil {
		return nil, apperror.New(constant.BADREQUEST, "Invalid order ID", err)
	}

	var order models.Order
	if err := s.Repo.FindOneWhere(&order, "id = ? AND user_id = ?", oID, uID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(constant.NOTFOUND, "Order not found", err)
		}
		return nil, apperror.New(constant.INTERNALSERVERERROR, "Failed to fetch order", err)
	}

	if err := s.Repo.FindByIDWithPreload(&order, order.ID, "Items.Product"); err != nil {
		return nil, apperror.New(constant.INTERNALSERVERERROR, "Failed to load order details", err)
	}

	return &order, nil
}

func (s *OrderService) CancelOrder(userID, orderID string) (err error) {

	uID, err := uuid.Parse(userID)
	if err != nil {
		return apperror.New(constant.BADREQUEST, "Invalid user ID", err)
	}

	oID, err := uuid.Parse(orderID)
	if err != nil {
		return apperror.New(constant.BADREQUEST, "Invalid order ID", err)
	}

	tx := s.Repo.Begin()
	if tx.Error != nil {
		return apperror.New(constant.INTERNALSERVERERROR, "Failed to start transaction", nil)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		} else if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	var order models.Order
	if err = tx.First(&order, "id = ? AND user_id = ?", oID, uID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.New(constant.NOTFOUND, "Order not found", err)
		}
		return apperror.New(constant.INTERNALSERVERERROR, "Failed to fetch order", err)
	}

	if order.Status == constant.OrderStatusCancelled || order.Status == constant.OrderStatusDelivered {
		return apperror.New(constant.BADREQUEST, "Cannot cancel delivered or already cancelled orders", nil)
	}

	var orderItems []models.OrderItem
	if err = tx.Where("order_id = ?", order.ID).Find(&orderItems).Error; err != nil {
		return apperror.New(constant.INTERNALSERVERERROR, "Failed to fetch order items", err)
	}

	for _, item := range orderItems {
		if err = tx.Model(&models.Variant{}).
			Where("id = ?", item.VariantID).
			UpdateColumn("quantity", gorm.Expr("quantity + ?", item.Quantity)).
			UpdateColumn("updated_at", time.Now()).Error; err != nil {
			return apperror.New(constant.INTERNALSERVERERROR, "Failed to update quantity", err)
		}
	}

	// Update order status
	if err = tx.Model(&order).Updates(map[string]interface{}{
		"status":     constant.OrderStatusCancelled,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return apperror.New(constant.INTERNALSERVERERROR, "Failed to cancel order", err)
	}

	return
}

func (s *OrderService) DeleteOrder(userID string, orderID string) error {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return apperror.New(constant.BADREQUEST, "Invalid user ID", err)
	}

	oID, err := uuid.Parse(orderID)
	if err != nil {
		return apperror.New(constant.BADREQUEST, "Invalid order ID", err)
	}

	// Check order exist
	var order models.Order
	if err := s.Repo.FindOneWhere(&order, "id = ? AND user_id = ?", oID, uID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.New(constant.NOTFOUND, "Order not found", err)
		}
		return apperror.New(constant.INTERNALSERVERERROR, "Failed to fetch order", err)
	}

	// Check order status(must be cancelled)
	if order.Status != constant.OrderStatusCancelled {
		return apperror.New(constant.BADREQUEST, "Can only delete cancelled orders", err)
	}

	if err := s.Repo.Raw("DELETE FROM order_items WHERE order_id = ?", oID); err != nil {
		return apperror.New(constant.INTERNALSERVERERROR, "Failed to delete order items", err)
	}

	if err := s.Repo.Delete(&models.Order{}, oID); err != nil {
		return apperror.New(constant.INTERNALSERVERERROR, "Failed to delete order", err)
	}

	return nil
}

func (s *OrderService) GetAllOrders(statusFilter, userIdFilter string) ([]models.Order, error) {
	var orders []models.Order
	var args []interface{}
	var query string

	if statusFilter != "" && userIdFilter != "" {
		uID, err := uuid.Parse(userIdFilter)
		if err != nil {
			return nil, apperror.New(constant.BADREQUEST, "Invalid user ID filter", err)
		}

		query = "status = ? AND user_id = ?"
		args = []interface{}{statusFilter, uID}

	} else if statusFilter != "" {
		query = "status = ?"
		args = []interface{}{statusFilter}

	} else if userIdFilter != "" {

		uID, err := uuid.Parse(userIdFilter)
		if err != nil {
			return nil, apperror.New(constant.BADREQUEST, "Invalid user ID filter", err)
		}

		query = "user_id = ?"
		args = []interface{}{uID}

	} else {
		if err := s.Repo.FindAllWithPreload(&orders, "Items.Product"); err != nil {
			return nil, apperror.New(constant.INTERNALSERVERERROR, "Failed to fetch all orders", err)
		}

		return orders, nil
	}

	if err := s.Repo.FindWhereWithPreload(&orders, query, args, "Items.Product"); err != nil {
		return nil, apperror.New(constant.INTERNALSERVERERROR, "Failed to fetch orders", err)
	}

	return orders, nil
}

func (s *OrderService) UpdateOrderStatusAdmin(orderID, newStatus string) error {

	oID, err := uuid.Parse(orderID)
	if err != nil {
		return apperror.New(constant.BADREQUEST, "Invalid order ID", err)
	}

	var order models.Order
	if err := s.Repo.FindById(&order, oID); err != nil {
		return apperror.New(constant.NOTFOUND, "Order not found", err)
	}

	// If cancelling, restore stock
	if newStatus == constant.OrderStatusCancelled && order.Status != constant.OrderStatusCancelled {
		tx := s.Repo.Begin()
		if tx.Error != nil {
			return apperror.New(constant.INTERNALSERVERERROR, "Failed to start transaction", nil)
		}

		var orderItems []models.OrderItem
		if err := tx.Where("order_id = ?", order.ID).Find(&orderItems).Error; err != nil {
			s.Repo.Rollback(tx)
			return apperror.New(constant.INTERNALSERVERERROR, "Failed to fetch order items", err)
		}

		for _, item := range orderItems {
			if err := tx.Model(&models.Variant{}).
				Where("product_id = ? AND variant_id", item.ProductID, item.VariantID).
				UpdateColumn("quantity", gorm.Expr("quantity + ?", item.Quantity)).
				UpdateColumn("updated_at", time.Now()).Error; err != nil {
				s.Repo.Rollback(tx)
				return apperror.New(constant.INTERNALSERVERERROR, "Failed to restore inventory", err)
			}
		}

		if err := tx.Model(&order).Updates(map[string]interface{}{
			"status":     newStatus,
			"updated_at": time.Now(),
		}).Error; err != nil {
			s.Repo.Rollback(tx)
			return apperror.New(constant.INTERNALSERVERERROR, "Failed to update order status", err)
		}

		if err := s.Repo.Commit(tx); err != nil {
			return apperror.New(constant.INTERNALSERVERERROR, "Failed to finalize update", err)
		}
		return nil
	}

	// Normal status update
	if err := s.Repo.UpdateByFields(&order, oID, map[string]interface{}{
		"status":     newStatus,
		"updated_at": time.Now(),
	}); err != nil {
		return apperror.New(constant.INTERNALSERVERERROR, "Failed to update order status", err)
	}

	return nil
}

func (s *OrderService) UpdateOrderStatusUser(userID string, orderID string, status string) error {
	if status != constant.OrderStatusCancelled {
		return apperror.New(constant.BADREQUEST, "Users can only cancel orders", nil)
	}
	return s.CancelOrder(userID, orderID)
}
