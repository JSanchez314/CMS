package database

import (
	"JSanches/CMD/models"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DBConn *gorm.DB

func PostgresInit() {
	db, err := gorm.Open(postgres.Open(os.Getenv("DB_URL")), &gorm.Config{})
	if err != nil {
		log.Fatal("Error connecting to postgres database: ", err)
	}

	err = db.AutoMigrate(models.User{})
	if err != nil {
		log.Fatal("Error migrating database: ", err)
	}
}

func PostgresGetDB() *gorm.DB {
	return DBConn
}
