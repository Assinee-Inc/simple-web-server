package middleware

import (
	"context"
	"net/http"

	accountmodel "github.com/Assinee-Inc/docffy/internal/account/model"
	accountsvc "github.com/Assinee-Inc/docffy/internal/account/service"
	authmw "github.com/Assinee-Inc/docffy/internal/auth/handler/middleware"
	"github.com/Assinee-Inc/docffy/internal/shared/web"
)

type creatorContextKey string

const (
	CreatorContextKey   creatorContextKey = "creator"
	CreatorIDContextKey creatorContextKey = "creator_id"
)

// CreatorOnboardingMiddleware verifica se o usuário autenticado já concluiu a Etapa 2 do onboarding (criação do Creator).
// Caso contrário, retorna 403 Forbidden em JSON com código CREATOR_ONBOARDING_REQUIRED.
func CreatorOnboardingMiddleware(creatorService accountsvc.CreatorService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var creator *accountmodel.Creator
			var err error

			if userID, ok := authmw.GetAuthenticatedUserID(r.Context()); ok && userID > 0 {
				creator, err = creatorService.FindCreatorByUserID(userID)
			} else if email := authmw.GetAuthenticatedUserEmail(r.Context()); email != "" {
				creator, err = creatorService.FindCreatorByEmail(email)
			} else if user := authmw.Auth(r); user != nil {
				if user.ID > 0 {
					creator, err = creatorService.FindCreatorByUserID(user.ID)
				} else if user.Email != "" {
					creator, err = creatorService.FindCreatorByEmail(user.Email)
				}
			} else {
				web.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Não autorizado", nil)
				return
			}

			if err != nil || creator == nil {
				web.RespondError(w, http.StatusForbidden, "CREATOR_ONBOARDING_REQUIRED", "Conclua o cadastro de criador para acessar este recurso", nil)
				return
			}

			ctx := WithCreator(r.Context(), creator)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// WithCreator armazena a instância de Creator e seu ID no contexto da requisição.
func WithCreator(ctx context.Context, creator *accountmodel.Creator) context.Context {
	if creator == nil {
		return ctx
	}
	ctx = context.WithValue(ctx, CreatorContextKey, creator)
	ctx = context.WithValue(ctx, CreatorIDContextKey, creator.ID)
	return ctx
}

// CreatorFromContext recupera o Creator previamente injetado no contexto da requisição.
func CreatorFromContext(ctx context.Context) *accountmodel.Creator {
	if creator, ok := ctx.Value(CreatorContextKey).(*accountmodel.Creator); ok && creator != nil {
		return creator
	}
	return nil
}
