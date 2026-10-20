package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Assinee-Inc/docffy/internal/auth/model"
	authrepo "github.com/Assinee-Inc/docffy/internal/auth/repository"
	authsvc "github.com/Assinee-Inc/docffy/internal/auth/service"
	"github.com/Assinee-Inc/docffy/internal/config"
	"github.com/Assinee-Inc/docffy/internal/shared/web"
	"github.com/Assinee-Inc/docffy/pkg/database"
)

type contextKey string

const (
	UserIDContextKey contextKey = "user_id"
	UserEmailKey     contextKey = "user_email"
	CSRFTokenKey     contextKey = "csrf_token"
	UserKey          contextKey = "user"
	UserContextKey   contextKey = "user"
)

var ErrUnauthorized = errors.New("não autorizado")

// AuthMiddleware valida o token Bearer JWT (ou faz fallback caso TokenService seja omitido)
// e injeta as informações do usuário autenticado no contexto da requisição.
func AuthMiddleware(deps ...any) func(http.Handler) http.Handler {
	var tokenService authsvc.TokenService
	var userRepo authrepo.UserRepository

	for _, dep := range deps {
		switch d := dep.(type) {
		case authsvc.TokenService:
			tokenService = d
		case authrepo.UserRepository:
			userRepo = d
		}
	}

	if tokenService == nil {
		expHours := config.AppConfig.JWTExpirationHours
		if expHours <= 0 {
			expHours = 24
		}
		tokenService = authsvc.NewTokenService(config.AppConfig.JWTSecret, time.Duration(expHours)*time.Hour)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr, err := extractBearerToken(r)
			if err != nil {
				web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Não autorizado", nil)
				return
			}

			claims, err := tokenService.ValidateToken(tokenStr)
			if err != nil {
				web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Token inválido ou expirado", nil)
				return
			}

			var user *model.User
			if userRepo != nil {
				user = userRepo.FindByID(claims.UID)
				if user == nil && claims.Email != "" {
					user = userRepo.FindByEmail(claims.Email)
				}
			} else if database.DB != nil {
				repo := authrepo.NewGormUserRepository(database.DB)
				user = repo.FindByID(claims.UID)
				if user == nil && claims.Email != "" {
					user = repo.FindByEmail(claims.Email)
				}
			}

			if user == nil {
				now := time.Now()
				user = &model.User{
					PublicID:        claims.Sub,
					Email:           claims.Email,
					EmailVerifiedAt: &now,
				}
				user.ID = claims.UID
			}

			ctx := WithAuthenticatedUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractBearerToken(r *http.Request) (string, error) {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if authHeader == "" {
		return "", ErrUnauthorized
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", ErrUnauthorized
	}

	tokenStr := strings.TrimSpace(parts[1])
	if tokenStr == "" {
		return "", ErrUnauthorized
	}

	return tokenStr, nil
}

// WithAuthenticatedUser injeta User, UserID e UserEmail no contexto.
func WithAuthenticatedUser(ctx context.Context, user *model.User) context.Context {
	if user == nil {
		return ctx
	}
	ctx = context.WithValue(ctx, UserIDContextKey, user.ID)
	ctx = context.WithValue(ctx, UserEmailKey, user.Email)
	ctx = context.WithValue(ctx, UserKey, user)
	return ctx
}

// GetAuthenticatedUserID retorna o ID primário do usuário autenticado no contexto.
func GetAuthenticatedUserID(ctx context.Context) (uint, bool) {
	if id, ok := ctx.Value(UserIDContextKey).(uint); ok && id > 0 {
		return id, true
	}
	if u, ok := ctx.Value(UserKey).(*model.User); ok && u != nil && u.ID > 0 {
		return u.ID, true
	}
	return 0, false
}

// GetAuthenticatedUserEmail retorna o e-mail do usuário autenticado no contexto.
func GetAuthenticatedUserEmail(ctx context.Context) string {
	if email, ok := ctx.Value(UserEmailKey).(string); ok && email != "" {
		return email
	}
	if u, ok := ctx.Value(UserKey).(*model.User); ok && u != nil {
		return u.Email
	}
	return ""
}

// GetAuthenticatedUser retorna a instância de User presente no contexto.
func GetAuthenticatedUser(ctx context.Context) *model.User {
	if u, ok := ctx.Value(UserKey).(*model.User); ok && u != nil {
		return u
	}
	return nil
}

// GetCSRFToken mantido por compatibilidade até a remoção completa de templates SSR.
func GetCSRFToken(r *http.Request) string {
	if token, ok := r.Context().Value(CSRFTokenKey).(string); ok {
		return token
	}
	return ""
}

// Auth recupera o User autenticado a partir do contexto da requisição ou do banco de dados.
func Auth(r *http.Request) *model.User {
	if u := GetAuthenticatedUser(r.Context()); u != nil {
		if database.DB != nil && u.Subscription == nil && u.Email != "" {
			userRepository := authrepo.NewGormUserRepository(database.DB)
			if dbUser := userRepository.FindByEmail(u.Email); dbUser != nil {
				return dbUser
			}
		}
		return u
	}

	userEmail, ok := r.Context().Value(UserEmailKey).(string)
	if !ok || userEmail == "" {
		return nil
	}

	if database.DB == nil {
		return nil
	}
	userRepository := authrepo.NewGormUserRepository(database.DB)
	return userRepository.FindByEmail(userEmail)
}
