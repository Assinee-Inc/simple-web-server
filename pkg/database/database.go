package database

import (
	"log"

	accountmodel "github.com/anglesson/simple-web-server/internal/account/model"
	authmodel "github.com/anglesson/simple-web-server/internal/auth/model"
	"github.com/anglesson/simple-web-server/internal/config"
	deliverymodel "github.com/anglesson/simple-web-server/internal/delivery/model"
	librarymodel "github.com/anglesson/simple-web-server/internal/library/model"
	salesmodel "github.com/anglesson/simple-web-server/internal/sales/model"
	subscriptionmodel "github.com/anglesson/simple-web-server/internal/subscription/model"
	"github.com/anglesson/simple-web-server/pkg/utils"
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
	err := DB.AutoMigrate(
		&authmodel.User{},
		&subscriptionmodel.Subscription{},
		&salesmodel.Client{},
		&accountmodel.Creator{},
		&librarymodel.Ebook{},
		&salesmodel.Purchase{},
		&deliverymodel.DownloadLog{},
		&salesmodel.Transaction{})

	if err != nil {
		log.Panic("failed to migrate database")
	}

	backfillPublicIDs()
}

func backfillPublicIDs() {
	backfill[authmodel.User]("usr_")
	backfill[subscriptionmodel.Subscription]("sub_")
	backfill[accountmodel.Creator]("crt_")
	backfill[librarymodel.Ebook]("ebk_")
	backfill[librarymodel.File]("fil_")
	backfill[salesmodel.Client]("cli_")
	backfill[salesmodel.Purchase]("pur_")
	backfill[salesmodel.Transaction]("txn_")
}

func backfill[T any](prefix string) {
	var records []T
	DB.Where("public_id IS NULL OR public_id = ''").Find(&records)
	for i := range records {
		publicID := utils.GeneratePublicID(prefix)
		DB.Model(&records[i]).Update("public_id", publicID)
	}
	if len(records) > 0 {
		log.Printf("Backfilled %d %T records with PublicID (prefix: %s)", len(records), *new(T), prefix)
	}
}

func Close() {
	sqlDB, err := DB.DB()
	if err != nil {
		log.Panic("failed to close database")
	}
	sqlDB.Close()
}
