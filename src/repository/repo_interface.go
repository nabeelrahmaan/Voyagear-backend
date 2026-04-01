package repository

import (
	"voyagear/src/models"
	"gorm.io/gorm"
)

// THis interface will be accessed by all services(instead of repository struct directly). So service doesnt affect the db (we can easily change that)
// It is better for good abstraction
type PgSQLRepository interface {
	Insert(req interface{}) error
	FindById(obj interface{}, id interface{}) error
	Update(obj interface{}, id interface{}, update interface{}) error
	UpdateByFields(obj interface{}, id interface{}, fields map[string]interface{}) error
	Delete(obj interface{}, id interface{}) error
	DeleteOneWhere(obj interface{}, query string, args ...interface{}) error
	HardDelete(obj interface{}) error
	FindAll(obj interface{}) error
	FindAllWhere(obj interface{}, query interface{}, args ...interface{}) error
	FindOneWhere(obj interface{}, query string, args ...interface{}) error
	InsertAndReturnID(obj interface{}) (string, error)
	FindDistinct(obj interface{}, field string, query interface{}, args ...interface{}) error
	FindByIDWithPreload(obj interface{}, id interface{}, prealoads ...string) error
	FindWhereWithPreload(obj interface{}, query string, args []interface{}, preloads ...string) error
	FindAllWithPreload(obj interface{}, preloads ...string) error
	GetAllProducts(filter ProductFilter, page, pageSize int, sortBy, sortOrder string) ([]models.Product, int64, error)
	Raw(sql string, values ...interface{}) error
	Save(req interface{}) error
	Begin() *gorm.DB
	Commit(tx *gorm.DB) error
	Rollback(tx *gorm.DB) error
}
