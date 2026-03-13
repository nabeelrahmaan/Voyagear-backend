package database

import (
	"fmt"
	"sync"
	"voyagear/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var pgOnce sync.Once

func SetupDatbase(cfg *config.Config) (*gorm.DB) {
	pgDB := &gorm.DB{}
	pgOnce.Do(func()  {
		
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s timeZone=%s",
			cfg.DB.Host,
			cfg.DB.User,
			cfg.DB.Password,
			cfg.DB.Name,
			cfg.DB.Port,
			cfg.DB.SSLMode,
			cfg.DB.TimeZone,
		)

		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return
		}

		vgDB, err := db.DB()
		if err != nil {
			return
		}

		vgDB.SetMaxIdleConns(2)
		vgDB.SetMaxOpenConns(10)

		pgDB = db
	})
	return pgDB
}