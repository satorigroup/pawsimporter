package db

import (
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/jinzhu/gorm"
	"log"
)

func NewService(connectionString, mode string) (*gorm.DB, error) {
	db, err := gorm.Open("mssql", connectionString)
	if err != nil {
		return nil, err
	}
	db.DB()
	err = db.DB().Ping()

	if err != nil {
		return nil, err
	}

	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)

	log.Printf("DB is running in %s mode\n", mode)

	return db, nil
}
