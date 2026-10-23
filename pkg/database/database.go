package database

import (
	"log"

	accountmodel "github.com/Assinee-Inc/docffy/internal/account/model"
	authmodel "github.com/Assinee-Inc/docffy/internal/auth/model"
	"github.com/Assinee-Inc/docffy/internal/config"
	deliverymodel "github.com/Assinee-Inc/docffy/internal/delivery/model"
	librarymodel "github.com/Assinee-Inc/docffy/internal/library/model"
	salesmodel "github.com/Assinee-Inc/docffy/internal/sales/model"
	subscriptionmodel "github.com/Assinee-Inc/docffy/internal/subscription/model"
	"github.com/Assinee-Inc/docffy/pkg/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB
var err error

func Connect() {
	config.LoadConfigs()
	connectWithPostgres()
}

func connectWithPostgres() {
	connectGormAndMigrate(postgres.Open(config.AppConfig.DatabaseURL))
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

	dropLegacyUserColumns()
	backfillPublicIDs()
}

func dropLegacyUserColumns() {
	legacyCols := []string{
		"password",
		"password_reset_token",
		"password_reset_at",
		"email_verification_token",
		"csrf_token",
		"session_token",
	}
	migrator := DB.Migrator()
	for _, col := range legacyCols {
		if migrator.HasColumn(&authmodel.User{}, col) {
			if err := migrator.DropColumn(&authmodel.User{}, col); err != nil {
				log.Printf("Aviso: não foi possível remover coluna legada users.%s: %v", col, err)
			}
		}
	}
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
