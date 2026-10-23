package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/Assinee-Inc/docffy/internal/auth/model"
)

var (
	ErrInvalidToken = errors.New("token inválido")
	ErrExpiredToken = errors.New("token expirado")
)

// TokenClaims armazena os dados contidos no payload do JWT.
type TokenClaims struct {
	Sub   string `json:"sub"`
	UID   uint   `json:"uid"`
	Email string `json:"email"`
	Iat   int64  `json:"iat"`
	Exp   int64  `json:"exp"`
}

// TokenService define as operações stateless de geração e validação de JWT.
type TokenService interface {
	GenerateToken(user *model.User) (string, int64, error)
	ValidateToken(tokenStr string) (*TokenClaims, error)
}

type jwtTokenService struct {
	secret     []byte
	expiration time.Duration
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// NewTokenService cria uma instância de TokenService usando HMAC-SHA256 (HS256).
func NewTokenService(secret string, expiration time.Duration) TokenService {
	if secret == "" {
		secret = "docffy-default-dev-jwt-secret"
	}
	if expiration == 0 {
		expiration = 24 * time.Hour
	}
	return &jwtTokenService{
		secret:     []byte(secret),
		expiration: expiration,
	}
}

func (s *jwtTokenService) GenerateToken(user *model.User) (string, int64, error) {
	if user == nil {
		return "", 0, ErrInvalidToken
	}

	now := time.Now().UTC()
	exp := now.Add(s.expiration)
	expiresIn := int64(s.expiration.Seconds())

	headerBytes, err := json.Marshal(jwtHeader{
		Alg: "HS256",
		Typ: "JWT",
	})
	if err != nil {
		return "", 0, err
	}

	claims := TokenClaims{
		Sub:   user.PublicID,
		UID:   user.ID,
		Email: user.Email,
		Iat:   now.Unix(),
		Exp:   exp.Unix(),
	}
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", 0, err
	}

	encodedHeader := base64.RawURLEncoding.EncodeToString(headerBytes)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signingInput := encodedHeader + "." + encodedPayload

	sig := s.sign(signingInput)
	return signingInput + "." + sig, expiresIn, nil
}

func (s *jwtTokenService) ValidateToken(tokenStr string) (*TokenClaims, error) {
	tokenStr = strings.TrimSpace(tokenStr)
	if tokenStr == "" {
		return nil, ErrInvalidToken
	}

	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var hdr jwtHeader
	if err := json.Unmarshal(headerBytes, &hdr); err != nil || hdr.Alg != "HS256" {
		return nil, ErrInvalidToken
	}

	signingInput := parts[0] + "." + parts[1]
	expectedSig := s.sign(signingInput)
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return nil, ErrInvalidToken
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims TokenClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	if time.Now().UTC().Unix() >= claims.Exp {
		return nil, ErrExpiredToken
	}

	return &claims, nil
}

func (s *jwtTokenService) sign(signingInput string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(signingInput))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
