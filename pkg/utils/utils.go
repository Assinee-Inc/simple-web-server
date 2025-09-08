package utils

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
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
