package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anglesson/simple-web-server/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func getConfig() aws.Config {
	cfg, err := awsCfg.LoadDefaultConfig(context.TODO(),
		awsCfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			config.AppConfig.S3AccessKey,
			config.AppConfig.S3SecretKey,
			"", // session token (deixe vazio se não estiver usando)
		)),
		awsCfg.WithRegion(config.AppConfig.S3Region), // ou sua região
	)
	if err != nil {
		log.Fatalf("erro ao carregar configuração: %v", err)
	}
	return cfg
}

type S3Storage interface {
	UploadFile(file *multipart.FileHeader, key, cacheControl string) (string, error)
	DeleteFile(key string) error
	GenerateDownloadLink(key string) string
	GenerateDownloadLinkWithExpiration(key string, expirationSeconds int) string
	GeneratePreviewLinkWithExpiration(key, contentType string, expirationSeconds int) string
}

type s3Storage struct {
	client *s3.Client
	bucket string
	region string
}

func NewS3Storage() S3Storage {
	cfg := getConfig()
	client := s3.NewFromConfig(cfg)

	return &s3Storage{
		client: client,
		bucket: config.AppConfig.S3BucketName,
		region: config.AppConfig.S3Region,
	}
}

func (s *s3Storage) UploadFile(file *multipart.FileHeader, key, cacheControl string) (string, error) {
	// Abrir arquivo
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer src.Close()

	// **IMPORTANTE:** Definir o Content-Type corretamente (assumindo um padrão)
	// Você deve inferir o Content-Type a partir da extensão ou do conteúdo do arquivo.
	// Exemplo simplificado:
	contentType := "application/octet-stream"
	if strings.HasSuffix(key, ".jpg") || strings.HasSuffix(key, ".jpeg") {
		contentType = "image/jpeg"
	} else if strings.HasSuffix(key, ".png") {
		contentType = "image/png"
	} else if strings.HasSuffix(key, ".pdf") {
		contentType = "application/pdf"
	}

	// Upload para S3
	_, err = s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:       aws.String(s.bucket),
		Key:          aws.String(key),
		Body:         src,
		ContentType:  aws.String(contentType),
		CacheControl: aws.String(cacheControl),
	})
	if err != nil {
		return "", fmt.Errorf("erro ao fazer upload: %w", err)
	}

	// Gerar URL pública
	fileURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, key)
	return fileURL, nil
}

func (s *s3Storage) DeleteFile(key string) error {
	_, err := s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (s *s3Storage) GenerateDownloadLink(key string) string {
	presigner := s3.NewPresignClient(s.client)

	params := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	presignedURL, err := presigner.PresignGetObject(context.TODO(), params, func(opts *s3.PresignOptions) {})
	if err != nil {
		log.Printf("Erro ao gerar URL pré-assinada: %v", err)
		return ""
	}

	return presignedURL.URL
}

// GeneratePreviewLinkWithExpiration gera uma URL pré-assinada com expiração customizada (em segundos)
func (s *s3Storage) GeneratePreviewLinkWithExpiration(key, contentType string, expirationSeconds int) string {
	presigner := s3.NewPresignClient(s.client)
	log.Printf("ContentType: %s", contentType)
	params := &s3.GetObjectInput{
		Bucket:                     aws.String(s.bucket),
		Key:                        aws.String(key),
		ResponseContentType:        aws.String(contentType),
		ResponseContentDisposition: aws.String("inline"),
	}
	presignedURL, err := presigner.PresignGetObject(context.TODO(), params, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(expirationSeconds) * time.Second
	})
	if err != nil {
		log.Printf("Erro ao gerar URL pré-assinada: %v", err)
		return ""
	}
	return presignedURL.URL
}

// GenerateDownloadLinkWithExpiration gera uma URL pré-assinada com expiração customizada (em segundos)
func (s *s3Storage) GenerateDownloadLinkWithExpiration(key string, expirationSeconds int) string {
	presigner := s3.NewPresignClient(s.client)
	params := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}
	presignedURL, err := presigner.PresignGetObject(context.TODO(), params, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(expirationSeconds) * time.Second
	})
	if err != nil {
		log.Printf("Erro ao gerar URL pré-assinada: %v", err)
		return ""
	}
	return presignedURL.URL
}

func GetFile(filename string) (string, error) {
	cfg := getConfig()

	s3Client := s3.NewFromConfig(cfg)

	params := &s3.GetObjectInput{
		Bucket: aws.String(config.AppConfig.S3BucketName),
		Key:    aws.String(filename),
	}
	output, err := s3Client.GetObject(context.TODO(), params)
	if err != nil {
		return "", fmt.Errorf("erro ao baixar arquivo do S3: %w", err)
	}
	defer output.Body.Close()

	// Criar diretório temporário se não existir
	tempDir := "./temp"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("erro ao criar diretório temporário: %w", err)
	}

	// Criar nome de arquivo seguro (remover caracteres problemáticos)
	safeFilename := strings.ReplaceAll(filename, "/", "_")
	safeFilename = strings.ReplaceAll(safeFilename, "\\", "_")
	safeFilename = strings.ReplaceAll(safeFilename, ":", "_")

	// Caminho local onde você quer salvar o arquivo
	localPath := filepath.Join(tempDir, safeFilename)

	// Criar arquivo local
	f, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("erro ao criar arquivo local: %w", err)
	}
	defer f.Close()

	// Copiar conteúdo do S3 para o arquivo local
	_, err = io.Copy(f, output.Body)
	if err != nil {
		return "", fmt.Errorf("erro ao salvar conteúdo: %w", err)
	}

	return localPath, nil
}

// Upload faz upload de um arquivo para o S3
func Upload(file multipart.File, filename string) error {
	cfg := getConfig()
	s3Client := s3.NewFromConfig(cfg)

	_, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(config.AppConfig.S3BucketName),
		Key:    aws.String(filename),
		Body:   file,
	})
	return err
}

// GenerateDownloadLink gera um link de download para um arquivo
func GenerateDownloadLink(filename string) string {
	cfg := getConfig()
	s3Client := s3.NewFromConfig(cfg)
	presigner := s3.NewPresignClient(s3Client)

	params := &s3.GetObjectInput{
		Bucket: aws.String(config.AppConfig.S3BucketName),
		Key:    aws.String(filename),
	}

	presignedURL, err := presigner.PresignGetObject(context.TODO(), params, func(opts *s3.PresignOptions) {})
	if err != nil {
		log.Printf("Erro ao gerar URL pré-assinada: %v", err)
		return ""
	}

	return presignedURL.URL
}
