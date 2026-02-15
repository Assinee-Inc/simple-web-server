package config

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/joho/godotenv"
)

type AppConfiguration struct {
	AppName                string
	AppMode                string
	AppKey                 string
	Host                   string
	Port                   string
	DatabaseURL            string
	MailHost               string
	MailPort               string
	MailUsername           string
	MailPassword           string
	MailAuth               string
	MailFromAddress        string
	MailFromName           string
	MailContactAddress     string
	S3AccessKey            string
	S3SecretKey            string
	S3Region               string
	S3BucketName           string
	HubDesenvolvedorApi    string
	HubDesenvolvedorToken  string
	StripeSecretKey        string
	StripePriceID          string
	StripeWebhookSecret    string
	PlatformFeePercentage  float64
	HubDesenvolvedorActive bool
}

func (ac *AppConfiguration) IsProduction() bool {
	return ac.AppMode == "production"
}

var AppConfig AppConfiguration

func LoadConfigs() {
	AppConfig.AppMode = GetEnv("APPLICATION_MODE", "development")

	if AppConfig.AppMode == "development" {
		loadEnvFile(".env")
	}

	if AppConfig.AppMode == "testing" {
		log.Println("Carregando configurações para ambiente de teste...")
		loadEnvFile(".env.test.local")
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

	hubDevActiveStr := GetEnv("HUB_DEVSENVOLVEDOR_ACTIVE", "true")
	if active, err := strconv.ParseBool(hubDevActiveStr); err == nil {
		AppConfig.HubDesenvolvedorActive = active
	} else {
		AppConfig.HubDesenvolvedorActive = true
	}

	// Carrega a taxa da plataforma do ambiente (se definida). Caso contrário, mantém o padrão da Business (2,91%).
	if envVal, ok := os.LookupEnv("PLATFORM_FEE_PERCENTAGE"); ok && envVal != "" {
		if fee, err := strconv.ParseFloat(envVal, 64); err == nil {
			AppConfig.PlatformFeePercentage = fee
			Business.PlatformFeePercentage = fee
		} else {
			log.Printf("Aviso: PLATFORM_FEE_PERCENTAGE inválido, mantendo padrão %.4f", Business.PlatformFeePercentage)
			AppConfig.PlatformFeePercentage = Business.PlatformFeePercentage
		}
	} else {
		AppConfig.PlatformFeePercentage = Business.PlatformFeePercentage
	}
}

func GetEnv(key, fallback string) string {
	env, exists := os.LookupEnv(key)
	if exists {
		return env
	}

	return fallback
}

func loadEnvFile(filename ...string) {
	// Se não especificar arquivo, tenta carregar .env da raiz
	if len(filename) == 0 {
		err := godotenv.Load(".env")
		if err != nil {
			log.Printf("Aviso: Arquivo .env não encontrado: %v", err)
		}
		return
	}

	// Para arquivos específicos, tenta múltiplos caminhos possíveis
	// (raiz do projeto e diretório atual)
	for _, file := range filename {
		// Tenta na raiz do projeto (a partir do diretório do pacote config)
		_, callerFile, _, _ := runtime.Caller(0)
		configDir := filepath.Dir(callerFile)
		projectRoot := filepath.Dir(configDir)  // sobe um nível: internal/ -> raiz
		projectRoot = filepath.Dir(projectRoot) // sobe outro nível para raiz

		paths := []string{
			file,                             // diretório atual
			filepath.Join(projectRoot, file), // raiz do projeto
			filepath.Join("../..", file),     // relativo a internal/config
		}

		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				if loadErr := godotenv.Load(path); loadErr == nil {
					return // carregou com sucesso
				}
			}
		}

		log.Printf("Aviso: Arquivo %s não encontrado em nenhum caminho", file)
	}
}
