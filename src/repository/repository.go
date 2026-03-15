package repository

import (
	"errors"
	"reflect"

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
	if err := r.DB.Debug().Model(obj).Where("id = ?", id).Updates(fields).Error; err != nil {
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