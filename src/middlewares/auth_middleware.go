package middlewares

import (
	"net/http"
	"strings"

	"github.com/Barba2k2/aurora_backend/src/models"
	"github.com/Barba2k2/aurora_backend/src/services"
	"github.com/Barba2k2/aurora_backend/src/utils"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware é o middleware de autenticação
type AuthMiddleware struct {
	AuthService *services.AuthService
}

// NewAuthMiddleware cria uma nova instância do AuthMiddleware
func NewAuthMiddleware(authService *services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		AuthService: authService,
	}
}

// extractToken extrai o token JWT do cabeçalho Authorization
func (m *AuthMiddleware) extractToken(ctx *gin.Context) string {
	bearerToken := ctx.GetHeader("Authorization")
	if len(bearerToken) > 7 && strings.HasPrefix(bearerToken, "Bearer ") {
		return bearerToken[7:]
	}
	return ""
}

// RequireAuth exige que o usuário esteja autenticado
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Extraimos o token do cabeçalho
		tokenString := m.extractToken(ctx)
		if tokenString == "" {
			utils.SendErrorResponse(ctx, http.StatusUnauthorized, "UNAUTHORIZED", "Token de autenticação não fornecido", nil)
			ctx.Abort()
			return
		}

		// Validamos o token
		user, err := m.AuthService.GetUserFromToken(tokenString)
		if err != nil {
			utils.SendErrorResponse(ctx, http.StatusUnauthorized, "INVALID_TOKEN", "Token inválido ou expirado", nil)
			ctx.Abort()
			return
		}

		// Verificamos se o usuario esta ativo
		if user.Status != models.UserStatusActive {
			utils.SendErrorResponse(ctx, http.StatusForbidden, "USER_INACTIVE", "Usuario inativo", nil)
			ctx.Abort()
			return
		}

		// Armazenamos o usuário no contexto
		ctx.Set("user", user)
		ctx.Set("user_id", user.ID.String())
		ctx.Set("user_role", string(user.Role))

		ctx.Next()
	}
}

// RequireRole exige que o usuário tenha um role específico
func (m *AuthMiddleware) RequireRole(roles ...models.UserRole) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Verificamos se o usuario esta autenticado
		user, exists := ctx.Get("user")
		if !exists {
			utils.SendErrorResponse(ctx, http.StatusUnauthorized, "UNAUTHORIZED", "Usuário não autenticado", nil)
			ctx.Abort()
			return
		}

		// Verificamos se o usuário tem o papel necessário
		userObj := user.(*models.User)
		hasRole := false
		for _, role := range roles {
			if userObj.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			utils.SendErrorResponse(ctx, http.StatusForbidden, "FORBIDDEN", "Acesso negado", nil)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

// RequireClient exige que o usuário seja um cliente
func (m *AuthMiddleware) RequireClient() gin.HandlerFunc {
	return m.RequireRole(models.UserRoleClient)
}

// RequireProfessional exige que o usuário seja um profissional
func (m *AuthMiddleware) RequireProfessional() gin.HandlerFunc {
	return m.RequireRole(models.UserRoleProfessional, models.UserRoleStaff, models.UserRoleAdmin)
}

// RequireAdmin exige que o usuário seja um admin
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return m.RequireRole(models.UserRoleAdmin)
}

// RequireOwnerOrAdmin exige que o usuário seja um profissional proprietário ou adminastrador
func (m *AuthMiddleware) RequireOwnerOrAdmin() gin.HandlerFunc {
	return m.RequireRole(models.UserRoleProfessional, models.UserRoleAdmin)
}