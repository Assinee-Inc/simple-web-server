package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type Encrypter interface {
	HashPassword(password string) string
	CheckPasswordHash(hashedPassword, password string) bool
	GenerateToken(length int) string
}

type encrypterImpl struct {
}

func NewEncrypter() Encrypter {
	return &encrypterImpl{}
}

func (e *encrypterImpl) HashPassword(password string) string {
	// In a real application, use a secure hashing algorithm
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		return "" // Return empty string on error
	}

	// Convert the hashed password to a string
	hashedPassword := string(bytes)

	// Return the hashed password
	return hashedPassword
}

func (e *encrypterImpl) CheckPasswordHash(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (e *encrypterImpl) GenerateToken(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		log.Printf("Failed to generate random token: %v", err)
		return "" // Return empty string on error
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

// Error response structure
type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// ClientError handles 4xx client errors
func ClientError(w http.ResponseWriter, r *http.Request, status int, message string) {
	log.Printf("CLIENT ERROR: %s", message)
	w.WriteHeader(status)

	switch status {
	case http.StatusNotFound:
		http.ServeFile(w, r, "web/pages/error/404.html")
		return
	case http.StatusUnauthorized:
		http.ServeFile(w, r, "web/pages/error/401.html")
		return
	case http.StatusForbidden:
		http.ServeFile(w, r, "web/pages/error/403.html")
		return
	default:
		http.ServeFile(w, r, "web/pages/error/500.html")
	}
}

// ServerError handles 500 internal server errors
func ServerError(w http.ResponseWriter, r *http.Request, err error) {
	ClientError(w, r, http.StatusInternalServerError, "The server encountered an internal error and was unable to complete your request")
}

// NotFound returns a 404 not found error
func NotFound(w http.ResponseWriter, r *http.Request) {
	ClientError(w, r, http.StatusNotFound, "The requested resource could not be found")
}

// Unauthorized returns a 401 unauthorized error
func Unauthorized(w http.ResponseWriter, r *http.Request) {
	ClientError(w, r, http.StatusUnauthorized, "You are not authorized to access this resource")
}

// Forbidden returns a 403 forbidden error
func Forbidden(w http.ResponseWriter, r *http.Request) {
	ClientError(w, r, http.StatusForbidden, "You do not have permission to access this resource")
}

// BadRequest returns a 400 bad request error
func BadRequest(w http.ResponseWriter, r *http.Request, message string) {
	ClientError(w, r, http.StatusBadRequest, message)
}

// MethodNotAllowed returns a 405 method not allowed error
func MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	ClientError(w, r, http.StatusMethodNotAllowed, "The method is not allowed for the requested URL")
}

// NormalizeText remove acentos (diacríticos), caracteres especiais e converte para minúsculas.
// Ex: "Pão de Açúcar, 123!" -> "pao de acucar 123"
func NormalizeText(s string) string {
	// 1. Decomposição Canônica (NFD): Separa o caractere base do acento (Ex: 'ã' -> 'a' + '~').
	// O objetivo é isolar o acento.
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

	// unicode.Mn (Mark Nonspacing): Filtra e remove as marcas (os acentos)

	normalized, _, err := transform.String(t, s)
	if err != nil {
		// Em caso de erro (improvável para strings comuns),
		// retornamos a string original em minúsculas
		return strings.ToLower(s)
	}

	// 2. Converte para minúsculas
	// É crucial converter para minúsculas para resolver o Case-Insensitivity.
	result := strings.ToLower(normalized)

	// Opcional: Remover ou substituir caracteres não-letras/números/espaços.
	// Depende de quão limpa a busca precisa ser. Para um nome de produto,
	// talvez seja melhor remover pontuações.

	return result
}

func UuidV7() string {
	v7, err := uuid.NewV7()
	if err != nil {
		panic(fmt.Sprintf("erro ao gerar uuid: %s", err.Error()))
	}
	return v7.String()
}
