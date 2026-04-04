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
	ID       string
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
	Variant       *[]UpdateProductVarientsInput
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

func (s *ProductService) GetAllProducts(
	filter repository.ProductFilter,
	page, pageSize int,
	sortBy, sortOrder string,
	userID string,
) ([]models.ProductResponse, int64, error) {

	var productResponses []models.ProductResponse
	var totalCount int64

	query := s.Repo.GetDB().Table("products p")

	if userID != "" {
		uID, err := uuid.Parse(userID)
		if err != nil {
			return nil, 0, apperror.New(constant.BADREQUEST, "Invalid user id", err)
		}
		query = query.Select(`
			p.id, p.name, p.description, p.price, p.category,
			p.image_url, p.original_price, p.is_active, p.created_at, p.updated_at,
			EXISTS (
				SELECT 1 FROM wishlist_items wi
				INNER JOIN wishlists w ON w.id = wi.wishlist_id
				WHERE wi.product_id = p.id
				AND w.user_id = ?
			) as is_wishlisted
		`, uID)
	} else {
		query = query.Select(`
			p.id, p.name, p.description, p.price, p.category,
			p.image_url, p.original_price, p.is_active, p.created_at, p.updated_at,
			false as is_wishlisted
		`)
	}

	// Apply filters
	if filter.Search != "" {
		search := "%" + filter.Search + "%"
		query = query.Where("(p.name ILIKE ? OR p.description ILIKE ?)", search, search)
	}
	if filter.Category != "" {
		query = query.Where("p.category = ?", filter.Category)
	}
	if filter.MinPrice > 0 {
		query = query.Where("p.price >= ?", filter.MinPrice)
	}
	if filter.MaxPrice > 0 {
		query = query.Where("p.price <= ?", filter.MaxPrice)
	}
	if filter.Size != "" {
		query = query.
			Joins("JOIN variants v ON v.product_id = p.id").
			Where("v.size = ?", filter.Size)
	}

	// Count query
	countQuery := s.Repo.GetDB().Table("products p").
		Select("COUNT(DISTINCT p.id)")

	if filter.Search != "" {
		search := "%" + filter.Search + "%"
		countQuery = countQuery.Where("(p.name ILIKE ? OR p.description ILIKE ?)", search, search)
	}
	if filter.Category != "" {
		countQuery = countQuery.Where("p.category = ?", filter.Category)
	}
	if filter.MinPrice > 0 {
		countQuery = countQuery.Where("p.price >= ?", filter.MinPrice)
	}
	if filter.MaxPrice > 0 {
		countQuery = countQuery.Where("p.price <= ?", filter.MaxPrice)
	}
	if filter.Size != "" {
		countQuery = countQuery.
			Joins("JOIN variants v ON v.product_id = p.id").
			Where("v.size = ?", filter.Size)
	}

	if err := countQuery.Scan(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Sorting
	sortCol := "p.created_at"
	switch sortBy {
	case "name":
		sortCol = "p.name"
	case "price":
		sortCol = "p.price"
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc"
	}

	offset := (page - 1) * pageSize

	err := query.
		Group("p.id").
		Order(sortCol + " " + sortOrder).
		Limit(pageSize).
		Offset(offset).
		Find(&productResponses).Error

	if err != nil {
		return nil, 0, err
	}

	// Load variants
	if len(productResponses) > 0 {
		productIDs := make([]uuid.UUID, len(productResponses))
		for i, p := range productResponses {
			productIDs[i] = p.ID
		}

		var variants []models.Variant
		err = s.Repo.GetDB().Where("product_id IN ?", productIDs).Find(&variants).Error
		if err == nil {
			variantMap := make(map[uuid.UUID][]models.VariantResponse)
			for _, v := range variants {
				variantMap[v.ProductID] = append(variantMap[v.ProductID], models.VariantResponse{
					ID:        v.ID,
					Size:      v.Size,
					Quantity:  v.Quantity,
					CreatedAt: v.CreatedAt,
					UpdatedAt: v.UpdatedAt,
				})
			}
			for i := range productResponses {
				productResponses[i].Variants = variantMap[productResponses[i].ID]
			}
		}
	}

	return productResponses, totalCount, nil
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

	if input.Variant != nil {
		for _, pv := range *input.Variant {
			if pv.ID != "" {
				updates := map[string]interface{}{
					"size":     pv.Size,
					"quantity": pv.Quantity,
				}

				pvID, err := uuid.Parse(pv.ID)
				if err != nil {
					return nil, apperror.New(
						constant.BADREQUEST,
						"Invalid variant ID",
						err,
					)
				}

				if err := s.Repo.UpdateByFields(&models.Variant{}, pvID, updates); err != nil {
					return nil, apperror.New(
						constant.INTERNALSERVERERROR,
						"Failed to update variant",
						err,
					)
				}

				continue
			}

			prodVariant := models.Variant{
				ProductID: uuid.MustParse(productID),
				Size:      pv.Size,
				Quantity:  pv.Quantity,
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

	if err := s.Repo.FindByIDWithPreload(&product, productID, "Variant"); err != nil {
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

	if err := s.Repo.FindWhereWithPreload(&products, dbQuery, args, "Variant"); err != nil {
		return nil, apperror.New(
			constant.INTERNALSERVERERROR,
			"Failed to fetch products",
			err,
		)
	}

	return products, nil
}
