package repository

import "gorm.io/gorm"

type PgSQLRepository interface {
	Insert(req interface{}) error
	FindById(obj interface{}, id interface{}) error
	Update(obj interface{}, id interface{}, update interface{}) error
	UpdateByFields(obj interface{}, id interface{}, fields map[string]interface{}) error
	Delete(obj interface{}, id interface{}) error
	HardDelete(obj interface{}) error
	FindAll(obj interface{}) error
	FindAllWhere(obj interface{}, query interface{}, args ...interface{}) error
	FindOneWhere(obj interface{}, query string, args ...interface{}) error
	InsertAndReturnID(obj interface{}) (string, error)
	FindDistinct(obj interface{}, field string, query interface{}, args ...interface{}) error
	Raw(sql string, values interface{}) *gorm.DB
	Save(req interface{}) error
}
