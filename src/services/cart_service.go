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

func SetupCart(repo repository.PgSQLRepository) *CartService{
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

	// Check item already exist in cart
	var item models.CartItem
	err = c.Repo.FindOneWhere(&item, "user_id = ? AND product_id = ? AND size = ?", uID, pID, size)
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
		CartID: cart.ID,
		ProductID: pID,
		Size: size,
		Quantity: quantity,
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