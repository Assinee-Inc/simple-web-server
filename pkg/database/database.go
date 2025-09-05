package database

import (
	"log"

	"github.com/anglesson/simple-web-server/internal/config"
	"github.com/anglesson/simple-web-server/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB
var err error

func Connect() {
	if config.AppConfig.IsProduction() {
		connectWithPostgres()
	} else {
		connectWithSQLite()
	}
}

func connectWithPostgres() {
	connectGormAndMigrate(postgres.Open(config.AppConfig.DatabaseURL))
}

func connectWithSQLite() {
	connectGormAndMigrate(sqlite.Open("./mydb.db"))
}

func connectGormAndMigrate(dialector gorm.Dialector) {
	DB, err = gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Panic("failed to connect database")
	}
	migrate()
}

func migrate() {
	DB.AutoMigrate(&models.User{})
	DB.AutoMigrate(&models.Subscription{})
	DB.AutoMigrate(&models.ClientCreator{})
	DB.AutoMigrate(&models.Client{})
	DB.AutoMigrate(&models.Contact{})
	DB.AutoMigrate(&models.Creator{})
	DB.AutoMigrate(&models.Ebook{})
	DB.AutoMigrate(&models.Purchase{})
	DB.AutoMigrate(&models.DownloadLog{})
	DB.AutoMigrate(&models.Transaction{})
}

func Close() {
	sqlDB, err := DB.DB()
	if err != nil {
		log.Panic("failed to close database")
	}
	sqlDB.Close()
}
