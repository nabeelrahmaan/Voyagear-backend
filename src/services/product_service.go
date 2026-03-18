package services

import (
	"voyagear/src/models"
	"voyagear/src/repository"
	"voyagear/utils/apperror"
	"voyagear/utils/constant"

	"github.com/google/uuid"
)

type ProductService struct {
	Repo repository.PgSQLRepository
}

func SetupProductService(repo repository.PgSQLRepository) *ProductService {
	return &ProductService{
		Repo: repo,
	}
}

// Input structs
type UpdateProductVarientsInput struct {
	ID       *string
	Size     string
	Quantity int
}

type UpdateProductInput struct {
	Name          *string
	Description   *string
	Price         *int
	OriginalPrice *int
	ImageURL      *string
	Category      *string
	IsActive      *bool
	Variants         *[]UpdateProductVarientsInput
}

func (s *ProductService) CreateProduct(product *models.Product) error {
	if product == nil {
		return apperror.New(
			constant.BADREQUEST,
			"Product data is nil",
			nil,
		)
	}

	if err := s.Repo.Insert(product); err != nil {
		return apperror.New(
			constant.INTERNALSERVERERROR,
			err.Error(),
			err,
		)
	}

	return nil
}

func (s *ProductService) GetAllProducts(filter repository.ProductFilter,
	page, pageSize int,
	sortBy, sortOrder string,
) ([]models.Product, int64, error) {

	if page <= 0 {
		page = 1
	}

	if pageSize <= 0 {
		pageSize = 10
	}

	return s.Repo.GetAllProducts(repository.ProductFilter(filter), page, pageSize, sortBy, sortOrder)
}

func (s *ProductService) GetProductById(productID string) (*models.Product, error) {

	var product models.Product
	if err := s.Repo.FindByIDWithPreload(&product, productID, "Variants"); err != nil {
		return nil, apperror.New(
			constant.NOTFOUND,
			"Product not found",
			err,
		)
	}

	return &product, nil
}

func (s *ProductService) DeleteProduct(productID string) error {

	var product models.Product
	if err := s.Repo.FindById(&product, productID); err != nil {
		return apperror.New(
			constant.NOTFOUND,
			"Product not found",
			err,
		)
	}

	if err := s.Repo.Delete(&product, productID); err != nil {
		return apperror.New(
			constant.INTERNALSERVERERROR,

			"Failed to delete product",
			err,
		)
	}

	return nil
}

func (s *ProductService) UpdateProduct(productID string, input *UpdateProductInput) (*models.Product, error) {

	var product models.Product
	if err := s.Repo.FindByIDWithPreload(&product, productID, "Variants"); err != nil {
		return nil, apperror.New(
			constant.NOTFOUND,
			"Product not found",
			err,
		)
	}

	// Storing updates into map
	updates := map[string]interface{}{}

	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Price != nil {
		updates["price"] = *input.Price
	}
	if input.ImageURL != nil {
		updates["image_url"] = *input.ImageURL
	}
	if input.Category != nil {
		updates["category"] = *input.Category
	}
	if input.OriginalPrice != nil {
		updates["original_price"] = *input.OriginalPrice
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}

	if len(updates) > 0 {
		if err := s.Repo.UpdateByFields(&models.Product{}, productID, updates); err != nil {
			return nil, apperror.New(
				constant.INTERNALSERVERERROR,
				err.Error(),
				err,
			)
		}
	}

	if input.Variants != nil {
		for _, pv := range *input.Variants {
			if pv.ID != nil {
				updates := map[string]interface{}{
					"size":pv.Size,
					"quantity":pv.Quantity,
				}

				if err := s.Repo.UpdateByFields(&models.Variants{}, productID, updates); err != nil {
					return nil, apperror.New(
						constant.INTERNALSERVERERROR,
						"Failed to update variants",
						err,
					)
				}

				continue
			}

			prodVariant := models.Variants{
				ProductID: uuid.MustParse(productID),
				Size: pv.Size,
				Quantity: pv.Quantity,
			}

			if err := s.Repo.Insert(&prodVariant); err != nil {
				return nil, apperror.New(
					constant.INTERNALSERVERERROR,
					"Failed to insert new variant",
					err,
				)
			}
		}
	}

	if err := s.Repo.FindByIDWithPreload(&product, productID, "Variants"); err != nil {
		return nil, apperror.New(
			constant.INTERNALSERVERERROR,
			"Failed to fetch updated product",
			err,
		)
	}

	return &product, nil
}

func (s *ProductService) SearchProduct(query string, category string) ([]models.Product, error) {

	var products []models.Product

	dbQuery := "1 = 1"
	args := []interface{}{}

	if query != "" {
		dbQuery += " AND name ILIKE ?"
		args = append(args, "%"+query+"%")
	}
	if category != "" {
		dbQuery += " AND category = ?"
		args = append(args, category)
	}

	if err := s.Repo.FindWhereWithPreload(&products, dbQuery, args, "Sizes"); err != nil {
		return nil, apperror.New(
			constant.INTERNALSERVERERROR,
			"Failed to fetch products",
			err,
		)
	}

	return products, nil
}
