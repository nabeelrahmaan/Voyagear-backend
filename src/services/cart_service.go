package services

import (
	"voyagear/src/models"
	"voyagear/src/repository"
	"voyagear/utils/apperror"
	"voyagear/utils/constant"

	"github.com/google/uuid"
)

type CartService struct {
	Repo repository.PgSQLRepository
}

func SetupCart(repo repository.PgSQLRepository) *CartService {
	return &CartService{
		Repo: repo,
	}
}

func (c *CartService) GetCart(userID string) (*models.Cart, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperror.New(
			constant.BADREQUEST,
			"Invalid user ID",
			err,
		)
	}

	var cart models.Cart

	args := []interface{}{}
	args = append(args, uid)

	err = c.Repo.FindWhereWithPreload(cart, "user_id = ?", args, "CartItem")
	if err != nil {
		return nil, apperror.New(
			constant.NOTFOUND,
			"Cart not found",
			err,
		)
	}

	return &cart, nil
}

func (c *CartService) AddToCart(userID, productID, size string, quantity int) error {

	uID, err := uuid.Parse(userID)
	if err != nil {
		return apperror.New(
			constant.BADREQUEST,
			"Invalid user id",
			err,
		)
	}

	pID, err := uuid.Parse(productID)
	if err != nil {
		return apperror.New(
			constant.BADREQUEST,
			"Invalid product id",
			err,
		)
	}

	// Check cart already exist
	var cart models.Cart
	err = c.Repo.FindOneWhere(&cart, "user_id = ?", uID)
	if err != nil {

		cart = models.Cart{
			UserID: uID,
		}

		// Create new cart for user
		if err := c.Repo.Insert(&cart); err != nil {
			return apperror.New(
				constant.INTERNALSERVERERROR,
				"Failed to create cart",
				err,
			)
		}
	}

	// Check is size exist for product
	var variant models.Variant
	if err := c.Repo.FindOneWhere(&variant, "product_id = ? AND size = ?", pID, size); err != nil {
		return apperror.New(
			constant.BADREQUEST,
			"Size dont exist",
			err,
		)
	}

	// Check item already exist in cart
	var item models.CartItem
	err = c.Repo.FindOneWhere(&item, "cart_id = ? AND product_id = ? AND variant_id = ?", cart.ID, pID, variant.ID)
	if err == nil {

		newQuantity := item.Quantity + quantity
		updates := map[string]interface{}{
			"quantity": newQuantity,
		}

		// Updating existing item by quantity
		if eror := c.Repo.UpdateByFields(&models.CartItem{}, item.ID, updates); eror != nil {
			return apperror.New(
				constant.INTERNALSERVERERROR,
				"Failed to update quantity",
				eror,
			)
		}
		return nil
	}

	cartItem := models.CartItem{
		CartID:    cart.ID,
		ProductID: pID,
		VariantID: variant.ID,
		Quantity:  quantity,
	}

	// Add new item to cart
	if err := c.Repo.Insert(&cartItem); err != nil {
		return apperror.New(
			constant.INTERNALSERVERERROR,
			"Failed to add item to cart",
			err,
		)
	}

	return nil
}

func (c *CartService) UpdateCart(userID, itemID string, size *string, quantity *int) error {

	uID, err := uuid.Parse(userID)
	if err != nil {
		return apperror.New(
			constant.UNAUTHORIZED,
			"Invalid user id",
			err,
		)
	}

	iID, err := uuid.Parse(itemID)
	if err != nil {
		return apperror.New(
			constant.BADREQUEST,
			"Invalid product id",
			err,
		)
	}

	// Check cart exist for user
	var cart models.Cart
	if err := c.Repo.FindOneWhere(&cart, "user_id = ?", uID); err != nil {
		return apperror.New(
			constant.NOTFOUND,
			"Cart not found",
			err,
		)
	}

	// Check item exist in cart
	var cartItem models.CartItem
	if err := c.Repo.FindOneWhere(&cartItem, "id = ? AND cart_id = ?", iID, cart.ID); err != nil {
		return apperror.New(
			constant.NOTFOUND,
			"Item not found in cart",
			err,
		)
	}

	updates := map[string]interface{}{}

	if size != nil {
		updates["size"] = *size
	}

	if quantity != nil {
		if *quantity == 0 {
			return apperror.New(
				constant.BADREQUEST,
				"Quantity cant be zero",
				nil,
			)
		}
		updates["quantity"] = *quantity
	}
	
	if len(updates) == 0 {
		return apperror.New(
			constant.BADREQUEST,
			"No fields to be updated",
			nil,
		)
	}

	if err := c.Repo.UpdateByFields(&models.CartItem{}, cartItem.ID, updates); err != nil {
		apperror.New(
			constant.INTERNALSERVERERROR,
			"Failed to update cart item",
			err,
		)
	}

	return nil
}

func (c *CartService) RemoveItemFromCart(cartItemID string) error {
	itemID, err := uuid.Parse(cartItemID)
	if err != nil {
		return apperror.New(
			constant.BADREQUEST,
			"Invalid item id",
			err,
		)
	}

	var item models.CartItem
	err = c.Repo.FindById(&item, itemID)
	if err != nil {
		return apperror.New(
			constant.NOTFOUND,
			"Item not found",
			err,
		)
	}

	if err := c.Repo.Delete(&item, itemID); err != nil {
		return apperror.New(
			constant.INTERNALSERVERERROR,
			"Failed to remove item from cart",
			err,
		)
	} 

	return nil
}
