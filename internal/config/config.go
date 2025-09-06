package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type AppConfiguration struct {
	AppName               string
	AppMode               string
	AppKey                string
	Host                  string
	Port                  string
	DatabaseURL           string
	MailHost              string
	MailPort              string
	MailUsername          string
	MailPassword          string
	MailAuth              string
	MailFromAddress       string
	MailFromName          string
	MailContactAddress    string
	S3AccessKey           string
	S3SecretKey           string
	S3Region              string
	S3BucketName          string
	HubDesenvolvedorApi   string
	HubDesenvolvedorToken string
	StripeSecretKey       string
	StripePriceID         string
	StripeWebhookSecret   string
	PlatformFeePercentage float64
}

func (ac *AppConfiguration) IsProduction() bool {
	return ac.AppMode == "production"
}

var AppConfig AppConfiguration

func LoadConfigs() {
	AppConfig.AppMode = GetEnv("APPLICATION_MODE", "development")

	if AppConfig.AppMode == "development" {
		err := godotenv.Load()
		if err != nil {
			log.Printf("Aviso: Arquivo .env não encontrado: %v", err)
		}
	}

	AppConfig.AppName = GetEnv("APPLICATION_NAME", "Docffy")
	AppConfig.AppKey = GetEnv("APP_KEY", "")
	AppConfig.Host = GetEnv("HOST", "http://localhost")
	AppConfig.Port = GetEnv("PORT", "8080")
	AppConfig.DatabaseURL = GetEnv("DATABASE_URL", "./mydb.db")
	AppConfig.MailHost = GetEnv("MAIL_HOST", "")
	AppConfig.MailPort = GetEnv("MAIL_PORT", "587")
	AppConfig.MailUsername = GetEnv("MAIL_USERNAME", "")
	AppConfig.MailPassword = GetEnv("MAIL_PASSWORD", "")
	AppConfig.MailAuth = GetEnv("MAIL_AUTH", "PLAIN")
	AppConfig.MailFromAddress = GetEnv("MAIL_FROM_ADDRESS", "")
	AppConfig.MailContactAddress = GetEnv("MAIL_CONTACT_ADDRESS", "")
	AppConfig.S3AccessKey = GetEnv("S3_ACCESS_KEY", "")
	AppConfig.S3SecretKey = GetEnv("S3_SECRET_KEY", "")
	AppConfig.S3Region = GetEnv("S3_REGION", "sa-east-1")
	AppConfig.S3BucketName = GetEnv("S3_BUCKET_NAME", "")
	AppConfig.HubDesenvolvedorApi = GetEnv("HUB_DEVSENVOLVEDOR_API", "")
	AppConfig.HubDesenvolvedorToken = GetEnv("HUB_DEVSENVOLVEDOR_TOKEN", "")
	AppConfig.StripeSecretKey = GetEnv("STRIPE_SECRET_KEY", "")
	AppConfig.StripePriceID = GetEnv("STRIPE_PRICE_ID", "")
	AppConfig.StripeWebhookSecret = GetEnv("STRIPE_WEBHOOK_SECRET", "")

	// Carrega a taxa da plataforma do ambiente ou usa o padrão (5%)
	if platformFeeStr := GetEnv("PLATFORM_FEE_PERCENTAGE", "0.05"); platformFeeStr != "" {
		if fee, err := strconv.ParseFloat(platformFeeStr, 64); err == nil {
			AppConfig.PlatformFeePercentage = fee
			Business.PlatformFeePercentage = fee
		} else {
			log.Printf("Aviso: PLATFORM_FEE_PERCENTAGE inválido, usando padrão de 5%%")
			AppConfig.PlatformFeePercentage = 0.05
		}
	} else {
		AppConfig.PlatformFeePercentage = 0.05
	}
}

func GetEnv(key, fallback string) string {
	env, exists := os.LookupEnv(key)
	if exists {
		return env
	}

	return fallback
}
