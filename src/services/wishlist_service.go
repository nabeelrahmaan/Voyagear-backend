package services

import (
	"voyagear/src/models"
	"voyagear/src/repository"
	"voyagear/utils/apperror"
	"voyagear/utils/constant"

	"github.com/google/uuid"
)

type WishlistService struct {
	Repo repository.PgSQLRepository
}

func SetupWishlist (repo repository.PgSQLRepository) *WishlistService {
	return &WishlistService{
		Repo: repo,
	}
}

func (w *WishlistService) GetWishlist (userID string) (*models.Wishlist, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperror.New(
			constant.BADREQUEST,
			"Invalid user id",
			err,
		)
	}

	var wishlist models.Wishlist

	args := []interface{}{}
	args = append(args, uid)

	if err := w.Repo.FindWhereWithPreload(wishlist, "user_id = ?", args, "WishlistItem"); err != nil {
		return nil, apperror.New(
			constant.NOTFOUND,
			"Wishlist not found",
			err,
		)
	}

	return &wishlist, nil
}

func (w *WishlistService) AddToWishlist (userID, productID string) error {

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

	// Check is wishlist exist for the user (if not, then create a new one)
	var wishlist models.Wishlist
	err = w.Repo.FindOneWhere(&wishlist, "user_id = ?", uID)
	if err != nil {
		wishlist = models.Wishlist{
			UserID: uID,
		}

		if err := w.Repo.Insert(&wishlist); err != nil {
			return apperror.New(
				constant.INTERNALSERVERERROR,
				"Failed to create new wishlist",
				err,
			)
		}
	}

	// Check product already wishlisted
	var existing models.WishlistItem
	err = w.Repo.FindOneWhere(&existing, "wishlist_id = ? AND product_id = ?", wishlist.ID, pID)
	if err == nil {
		return apperror.New(
			constant.BADREQUEST,
			"Item already exist in wishlist",
			nil,
		)
	}

	item := models.WishlistItem{
		WishlistID: wishlist.ID,
		ProductID: pID,
	}

	if err := w.Repo.Insert(&item); err != nil {
		return apperror.New(
			constant.INTERNALSERVERERROR,
			"Failed to add item to the wishlist",
			err,
		)
	}
	
	return nil
}

func (w *WishlistService) RemoveFromWishlist (userID, productID string) error {

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

	var wishlist models.Wishlist
	if err := w.Repo.FindOneWhere(&wishlist, "user_id = ?", uID); err != nil {
		return apperror.New(
			constant.BADREQUEST,
			"Wishlist not found",
			err,
		)
	}

	var wishItem models.WishlistItem
	if err := w.Repo.FindOneWhere(&wishItem, "user_id = ? AND product_id = ?", uID, pID); err != nil {
		return apperror.New(
			constant.NOTFOUND,
			"Item not found in the wishlist",
			err,
		)
	}
	
	err = w.Repo.DeleteOneWhere(&models.WishlistItem{}, "user_id = ? AND product_id = ?", uID, pID)
	if err != nil {
		return apperror.New(
			constant.INTERNALSERVERERROR,
			"Failed to remove item from wishlist",
			err,
		)
	}

	return nil
}