package repository

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"voyagear/src/models"

	"gorm.io/gorm"
)

type Repository struct {
	DB *gorm.DB
}

func SetupRepo(db *gorm.DB) *Repository {
	return &Repository{
		DB: db,
	}
}

type ProductFilter struct {
	Search   string
	Category string
	Size     string
	MinPrice int
	MaxPrice int
}

func (r *Repository) Insert(req interface{}) error {
	if err := r.DB.Debug().Create(req).Error; err != nil {
		return err
	}

	return nil
}

func (r *Repository) FindById(obj interface{}, id interface{}) error {
	if err := r.DB.Debug().Where("id = ?", id).Find(obj).Error; err != nil {
		return err
	}

	return nil
}

func (r *Repository) Update(obj interface{}, id interface{}, update interface{}) error {
	if err := r.DB.Debug().Where("id = ?", id).First(obj).Updates(update).Error; err != nil {
		return err
	}

	return nil
}

func (r *Repository) UpdateByFields(obj interface{}, id interface{}, fields map[string]interface{}) error {
	if err := r.DB.Debug().Session(&gorm.Session{NewDB: true}).Model(obj).Where("id = ?", id).Updates(fields).Error; err != nil {
		return err
	}
	return nil
}

func (r *Repository) Delete(obj interface{}, id interface{}) error {
	if err := r.DB.Debug().Where("id = ?", id).Delete(obj).Error; err != nil {
		return err
	}

	return nil
}

func (r *Repository) HardDelete(obj interface{}) error {
	if err := r.DB.Unscoped().Delete(obj).Error; err != nil {
		return err
	}

	return nil
}

func (r *Repository) FindAll(obj interface{}) error {
	if err := r.DB.Debug().Find(obj).Error; err != nil {
		return err
	}

	return nil
}

func (r *Repository) FindAllWhere(obj interface{}, query interface{}, args ...interface{}) error {
	if err := r.DB.Debug().Where(query, args...).Find(obj).Error; err != nil {
		return err
	}

	return nil
}

func (r *Repository) FindOneWhere(obj interface{}, query string, args ...interface{}) error {
	if err := r.DB.Debug().Where(query, args...).First(obj).Error; err != nil {
		return err
	}

	return nil
}

func (r *Repository) InsertAndReturnID(obj interface{}) (string, error) {
	if err := r.DB.Debug().Create(obj).Error; err != nil {
		return "", err
	}

	value := reflect.ValueOf(obj).Elem()
	idField := value.FieldByName("ID")
	if !idField.IsValid() {
		return "", errors.New("ID field not found")
	}

	id := string(idField.String())
	return id, nil
}

func (r *Repository) FindDistinct(obj interface{}, field string, query interface{}, args ...interface{}) error {
	if err := r.DB.Debug().Model(obj).Distinct(field).Where(query, args...).Find(obj).Error; err!= nil {
		return err
	}

	return nil
}

func (r  *Repository) Raw(sql string, values ...interface{}) *gorm.DB {
	return r.DB.Raw(sql, values...)
}

func (r *Repository) Save(req interface{}) error {
	if err := r.DB.Debug().Save(req).Error; err != nil {
		return err
	}

	return nil
}


// Preloads are used to fetch relations with the data (eg:- fetching sizes with product. otherwise it stays empty)
func (r *Repository) FindByIDWithPreload(obj interface{}, id interface{}, preloads ...string) error {

	for _, preload := range preloads {
		r.DB = r.DB.Preload(preload)
	}

	if err := r.DB.Where("id = ?", id).First(obj).Error; err != nil {
		return err
	}

	return nil
}

func (r *Repository) FindWhereWithPreload(obj interface{}, query string, args []interface{}, preloads ...string) error {

	for _, preload := range preloads {
		r.DB = r.DB.Preload(preload)
	}

	if err := r.DB.Where(query, args...).Find(obj).Error; err != nil {
		return err
	}

	return nil
}

func (r *Repository) FindAllWithPreload(obj interface{}, preloads ...string) error {
	for _, preload := range preloads {
		r.DB = r.DB.Preload(preload)
	}

	if err := r.DB.Find(obj).Error; err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetAllProducts(filter ProductFilter, page, pagesize int, sortBy, sortOrder string ) ([]models.Product, int64, error) {

	var (
		ids        []string
		Products   []models.Product
		totalCount int64
	)
	
	db := r.DB.Table("prducts p")

	if filter.Search != "" {
		search :=  "%" + filter.Search + "%"
		db = db.Where("(p.name ILIKE ? OR p.description ILIKE ?)", search, search)
	}

	if filter.Category != "" {
		db = db.Where("p.category = ?", filter.Category)
	}

	if filter.MinPrice > 0 {
		db = db.Where("p.price >= ?", filter.MinPrice)
	}

	if filter.MinPrice > 0 {
		db = db.Where("p.price <= ?", filter.MaxPrice)
	}

	if filter.Size != "" {
		db = db.Joins("JOIN variants v ON v.product_id = p.id").
		Where("v.size = ?", filter.Size)
	}

	db.Select("COUNT(DISTINCT p.id)").Count(&totalCount)

	offset := (page - 1) * pagesize

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

	err := db.Select("p.id").
		Group("p.id, " + sortCol).
		Order(sortCol + " " + sortOrder).
		Limit(pagesize).
		Offset(offset).
		Pluck("p.id", &ids).Error

	if err != nil {
		return nil, 0, err
	}

	if len(ids) == 0 {
		return []models.Product{}, totalCount, nil
	}

	var quotedIds []string
	for _, id := range ids {
		quotedIds = append(quotedIds, fmt.Sprintf("'%s'", id))
	}

	err = r.DB.Model(&models.Product{}).
		Where("id IN ?", ids).
		Preload("variants").
		Order(fmt.Sprintf("array_positions(ARRAY[%s]::uuid[], id)", strings.Join(quotedIds, ","))).
		Find(&Products).Error

	return Products, totalCount, err
}