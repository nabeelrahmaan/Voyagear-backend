package services

import (
	"fmt"
	"strings"
	"voyagear/src/models"
	"voyagear/src/repository"
	"voyagear/utils/apperror"
	"voyagear/utils/constant"
)

type ProductService struct {
	Repo *repository.Repository
}

func SetupProductService(repo *repository.Repository) *ProductService {
	return &ProductService{
		Repo: repo,
	}
}

// Input structs
type UpdateProductSizeInput struct {
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
	Sizes         *[]UpdateProductSizeInput
}

type ProductFilter struct {
	Search   string
	Category string
	Size     string
	MinPrice int
	MaxPrice int
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
			"Failed to create product",
			err,
		)
	}

	return nil
}

func (s *ProductService) GetAllProducts(filter ProductFilter,
	page, pageSize int,
	sortBy, sortOrder string,
) ([]models.Product, int64, error) {

	var (
		ids        []string
		Products   []models.Product
		totalCount int64
	)

	db := s.Repo.DB.Table("products p")

	// Checking filters for filter products
	if filter.Search != "" {
		search := "%" + filter.Search + "%"
		db = db.Where("(p.name ILIKE ? OR p.description ILIKE ?)", search, search)
	}

	if filter.Category != "" {
		db = db.Where("p.category = ?", filter.Category)
	}

	if filter.MinPrice > 0 {
		db = db.Where("p.price >= ?", filter.MinPrice)
	}

	if filter.MaxPrice > 0 {
		db = db.Where("p.price <= ?", filter.MaxPrice)
	}

	if filter.Category != "" {
		db = db.Where("p.category = ?", filter.Category)
	}

	if filter.Size != "" {
		db = db.Joins("JOIN product_sizes ps ON ps.product_id == p.id").
			Where("ps.size = ?", filter.Size)
	}

	db.Select("COUNT(DISTINCT p.id)").Count(&totalCount)

	offset := (page - 1) * pageSize

	// Sorting products
	sortCol := "p.category"
	if sortBy == "name" {
		sortCol = "p.name"
	}
	if sortBy == "price" {
		sortCol = "p.price"
	}

	if err := db.Select("p.id").
		Group("p.id, "+sortCol).
		Order(sortCol+" "+sortOrder).
		Limit(pageSize).Offset(offset).
		Pluck("p.id", &ids).Error; err != nil {
		return nil, 0, err
	}

	if len(ids) == 0 {
		return []models.Product{}, totalCount, nil
	}

	var quotedIds []string
	for _, id := range ids {
		quotedIds = append(quotedIds, fmt.Sprintf("'%s'", id))
	}

	err := s.Repo.DB.Model(&models.Product{}).
		Where("id IN ?", ids).
		Preload("sizes").
		Order(fmt.Sprintf("array_position(ARRAY[%s]::uuid[], id)", strings.Join(quotedIds, ","))).
		Find(&Products).Error

	return Products, totalCount, err
}

func (s *ProductService) GetProductById(productID string) (*models.Product, error) {

	var product models.Product
	if err := s.Repo.FindByIDWithPreload(&product, productID, "Sizes"); err != nil {
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
	if err := s.Repo.FindByIDWithPreload(&product, productID, "Sizes"); err != nil {
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
				"Failed to update product",
				err,
			)
		}
	}

	if input.Sizes != nil {
		for _, sReq := range *input.Sizes {
			if sReq.ID != nil {
				fields := map[string]interface{}{
					"size":     sReq.Size,
					"quantity": sReq.Quantity,
				}

				// Updating product size seperately
				if err := s.Repo.UpdateByFields(&models.Product{}, *sReq.ID, fields); err != nil {
					return nil, apperror.New(
						constant.INTERNALSERVERERROR,
						"Failed to update product size",
						err,
					)
				}
				continue
			}

			newSize := models.ProductSize{
				ProductID: product.ID,
				Size:      sReq.Size,
				Quantity:  sReq.Quantity,
			}
			if err := s.Repo.Insert(&newSize); err != nil {
				return nil, apperror.New(
					constant.INTERNALSERVERERROR,
					"Failed to add product size",
					err,
				)
			}
		}
	}

	if err := s.Repo.FindByIDWithPreload(&product, productID, "Sizes"); err != nil {
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
