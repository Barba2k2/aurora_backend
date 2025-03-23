package controllers

import (
	"net/http"
	"strings"

	"github.com/Barba2k2/aurora_backend/src/models"
	"github.com/Barba2k2/aurora_backend/src/services"
	"github.com/Barba2k2/aurora_backend/src/utils"
	"github.com/gin-gonic/gin"
)

// ClientAuthController manipula as requisições de autenticação de clintes
type ClientAuthController struct {
	AuthService *services.AuthService
}

// NewClientAuthController cria uma nova instância de ClientAuthController
func NewClientAuthController(authService *services.AuthService) *ClientAuthController {
	return &ClientAuthController{
		AuthService: authService,
	}
}

// Register manipula o registro de novos clientes
// @Summary Registra um novo cliente
// @Description Cria um novo usuário com o perfil de cliente
// @Tags client-auth
// @Accept json
// @Produce json
// @Param request body services.RegisterRequest true "Dados de registro"
// @Success 201 {object} models.User "Usuário criado com sucesso"
// @Failure 400 {object} ErrorResponse "Dados inválidos"
// @Failure 409 {object} ErrorResponse "Usuário já existe"
// @Failure 422 {object} ErrorResponse "Validação falhou"
// @Failure 500 {object} ErrorResponse "Erro interno do servidor"
// @Router /api/v1/client/auth/register [POST]
func (c *ClientAuthController) Register(ctx *gin.Context) {
	var req services.RegisterRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(ctx, http.StatusBadRequest, "INVALID_REQUEST", "Formato da requisição inválido", nil)
		return
	}

	// Forçamos o role para client
	req.Role = models.UserRoleClient

	// Validamos os dados
	if req.Email == "" || req.Password == "" || req.Name == "" {
		utils.SendErrorResponse(ctx, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Campos obrigatórios não preenchidos", map[string]interface{}{
			"email":    "Email é obrigatório",
			"password": "Senha é obrigatória",
			"name":     "Nome é obrigatório",
		})
		return
	}

	// Verificamos se as senhas são iguais
	if req.Password != req.ConfirmPassword {
		utils.SendErrorResponse(ctx, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Senhas não conferem", map[string]interface{}{
			"confirm_password": "Senhas não conferem",
		})
		return
	}

	// Criamos o usuário
	user, err := c.AuthService.Register(req)
	if err != nil {
		switch err {
		case services.ErrPasswordTooWeak:
			utils.SendErrorResponse(ctx, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Senha muito fraca", map[string]interface{}{
				"password": "A senha deve conter letras maiúsculas, minúsculas, números e caracteres especiais",
			})
		case services.ErrPasswordConfirmation:
			utils.SendErrorResponse(ctx, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Senhas não conferem", map[string]interface{}{
				"confirm_password": "Senhas não conferem",
			})
		default:
			if strings.Contains(err.Error(), "already exists") {
				utils.SendErrorResponse(ctx, http.StatusConflict, "USER_EXISTS", "Usuário já cadastrado com este email ou telefone", nil)
			} else {
				utils.SendErrorResponse(ctx, http.StatusInternalServerError, "SERVER_ERROR", "Erro ao registrar usuário", nil)
			}
		}
		return
	}

	// Retornamos o usuário criado
	utils.SendSuccessResponse(ctx, http.StatusCreated, user, nil)
}

// Login manipula o login de clientes
// @Summary Login de cliente
// @Description Autentica um cliente e retorna tokens de acesso
// @Tags client-auth
// @Accept json
// @Produce json
// @Param request body services.LoginRequest true "Dados de login"
// @Success 200 {object} services.TokenResponse "Tokens gerados com sucesso"
// @Failure 400 {object} ErrorResponse "Dados inválidos"
// @Failure 401 {object} ErrorResponse "Credenciais inválidas"
// @Failure 403 {object} ErrorResponse "Usuário bloqueado"
// @Failure 422 {object} ErrorResponse "Validação falhou"
// @Failure 500 {object} ErrorResponse "Erro interno do servidor"
// @Router /api/v1/client/auth/login [POST]
func (c *ClientAuthController) Login(ctx *gin.Context) {
	var req services.LoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(ctx, http.StatusBadRequest, "INVALID_REQUEST", "Formato da requisição inválido", nil)
		return
	}

	// Validamos os dados
	if req.Email == "" || req.Password == "" {
		utils.SendErrorResponse(ctx, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Campos obrigatórios não preenchidos", map[string]interface{}{
			"email":    "Email é obrigatório",
			"password": "Senha é obrigatória",
		})
		return
	}

	// Realizamos o login
	user, tokens, err := c.AuthService.Login(req)
	if err != nil {
		switch err {
		case services.ErrInvalidLogin:
			utils.SendErrorResponse(ctx, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Email ou senha inválidos", nil)
		case services.ErrUserBlocked:
			utils.SendErrorResponse(ctx, http.StatusForbidden, "USER_BLOCKED", "Usuário bloqueado por excesso de tentativas de login", nil)
		case services.ErrUserInactive:
			utils.SendErrorResponse(ctx, http.StatusForbidden, "USER_INACTIVE", "Usuário inativo", nil)
		default:
			utils.SendErrorResponse(ctx, http.StatusInternalServerError, "SERVER_ERROR", "Erro ao realizar login", nil)
		}
		return
	}

	// Verificamos se o usuário é cliente
	if user.Role != models.UserRoleClient {
		utils.SendErrorResponse(ctx, http.StatusForbidden, "INVALID_ROLE", "Este usuário não é um cliente", nil)
		return
	}

	// Retornamos os tokens
	utils.SendSuccessResponse(ctx, http.StatusOK, tokens, nil)
}

// RefreshToken manipula a renovação de token
// @Summary Renovação de token
// @Description Renova o token de acesso usando o refresh token
// @Tags client-auth
// @Accept json
// @Produce json
// @Param request body services.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} services.TokenResponse "Tokens renovados com sucesso"
// @Failure 400 {object} ErrorResponse "Dados inválidos"
// @Failure 401 {object} ErrorResponse "Token inválido"
// @Failure 500 {object} ErrorResponse "Erro interno do servidor"
// @Router /api/v1/client/auth/refresh [post]
func (c *ClientAuthController) RefreshToken(ctx *gin.Context) {
	var req services.RefreshTokenRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(ctx, http.StatusBadRequest, "INVALID_REQUEST", "Formato da requisição inválido", nil)
		return
	}

	// Validamos os dados
	if req.RefreshToken == "" {
		utils.SendErrorResponse(ctx, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Refresh token não fornecido", map[string]interface{}{
			"refresh_token": "Refresh token é obrigatório",
		})
		return
	}

	// Renovamos o token
	tokens, err := c.AuthService.RefreshToken(req)
	if err != nil {
		switch err {
		case services.ErrInvalidToken:
			utils.SendErrorResponse(ctx, http.StatusUnauthorized, "INVALID_TOKEN", "Token inválido ou expirado", nil)
		case services.ErrUserInactive:
			utils.SendErrorResponse(ctx, http.StatusForbidden, "USER_INACTIVE", "Usuário inativo", nil)
		default:
			utils.SendErrorResponse(ctx, http.StatusInternalServerError, "SERVER_ERROR", "Erro ao renovar token", nil)
		}
		return
	}

	// Retornamos os novos tokens
	utils.SendSuccessResponse(ctx, http.StatusOK, tokens, nil)
}

// ForgotPasswordEmail manipula a solicitação de recuperação de senha via email
// @Summary Recuperação de senha via email
// @Description Envia um email com token para recuperação de senha
// @Tags client-auth
// @Accept json
// @Produce json
// @Param request body services.ForgotPasswordRequest true "Email do usuário"
// @Success 200 {object} SuccessResponse "Email enviado com sucesso"
// @Failure 400 {object} ErrorResponse "Dados inválidos"
// @Failure 404 {object} ErrorResponse "Email não encontrado"
// @Failure 429 {object} ErrorResponse "Muitas requisições"
// @Failure 500 {object} ErrorResponse "Erro interno do servidor"
// @Router /api/v1/client/auth/forgot-password/email [post]
func (c *ClientAuthController) ForgotPasswordEmail(ctx *gin.Context) {
	var req services.ForgotPasswordRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(ctx, http.StatusBadRequest, "INVALID_REQUEST", "Formato da requisição inválido", nil)
		return
	}

	// Validamos os dados
	if req.Email == "" {
		utils.SendErrorResponse(ctx, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Email não fornecido", map[string]interface{}{
			"email": "Email é obrigatório",
		})
		return
	}

	// Adicionamos informações do cliente para auditoria
	req.ClientIP = ctx.ClientIP()
	req.UserAgent = ctx.GetHeader("User-Agent")

	// Enviamos o email de recuperação
	err := c.AuthService.ForgotPasswordEmail(req)
	if err != nil {
		switch err {
		case services.ErrEmailNotFound:
			utils.SendErrorResponse(ctx, http.StatusNotFound, "EMAIL_NOT_FOUND", "Não existe usuário com este email", nil)
		case services.ErrUserInactive:
			utils.SendErrorResponse(ctx, http.StatusForbidden, "USER_INACTIVE", "Usuário inativo", nil)
		case services.ErrTooManyRequests:
			utils.SendErrorResponse(ctx, http.StatusTooManyRequests, "TOO_MANY_REQUESTS", "Muitas solicitações em um curto período", nil)
		default:
			utils.SendErrorResponse(ctx, http.StatusInternalServerError, "SERVER_ERROR", "Erro ao enviar email de recuperação", nil)
		}
		return
	}

	// Retornamos sucesso
	utils.SendSuccessResponse(ctx, http.StatusOK, nil, map[string]interface{}{
		"message": "Email de recuperação enviado com sucesso",
	})
}

// ForgotPasswordSMS manipula a solicitação de recuperação de senha via SMS
// @Summary Recuperação de senha via SMS
// @Description Envia um SMS com código para recuperação de senha
// @Tags client-auth
// @Accept json
// @Produce json
// @Param request body services.ForgotPasswordRequest true "Telefone do usuário"
// @Success 200 {object} SuccessResponse "SMS enviado com sucesso"
// @Failure 400 {object} ErrorResponse "Dados inválidos"
// @Failure 404 {object} ErrorResponse "Telefone não encontrado"
// @Failure 429 {object} ErrorResponse "Muitas requisições"
// @Failure 500 {object} ErrorResponse "Erro interno do servidor"
// @Router /api/v1/client/auth/forgot-password/sms [post]
func (c *ClientAuthController) ForgotPasswordSMS(ctx *gin.Context) {
	var req services.ForgotPasswordRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(ctx, http.StatusBadRequest, "INVALID_REQUEST", "Formato da requisição inválido", nil)
		return
	}

	// Validamos os dados
	if req.Phone == "" {
		utils.SendErrorResponse(ctx, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Telefone não fornecido", map[string]interface{}{
			"phone": "Telefone é obrigatório",
		})
		return
	}

	// Adicionamos informações do cliente para auditoria
	req.ClientIP = ctx.ClientIP()
	req.UserAgent = ctx.GetHeader("User-Agent")

	// Enviamos o SMS de recuperação
	err := c.AuthService.ForgotPasswordSMS(req)
	if err != nil {
		switch err {
		case services.ErrPhoneNotFound:
			utils.SendErrorResponse(ctx, http.StatusNotFound, "PHONE_NOT_FOUND", "Não existe usuário com este telefone", nil)
		case services.ErrUserInactive:
			utils.SendErrorResponse(ctx, http.StatusForbidden, "USER_INACTIVE", "Usuário inativo", nil)
		case services.ErrTooManyRequests:
			utils.SendErrorResponse(ctx, http.StatusTooManyRequests, "TOO_MANY_REQUESTS", "Muitas solicitações em um curto período", nil)
		default:
			utils.SendErrorResponse(ctx, http.StatusInternalServerError, "SERVER_ERROR", "Erro ao enviar SMS de recuperação", nil)
		}
		return
	}

	// Retornamos sucesso
	utils.SendSuccessResponse(ctx, http.StatusOK, nil, map[string]interface{}{
		"message": "SMS de recuperação enviado com sucesso",
	})
}

// ForgotPasswordWhatsApp manipula a solicitação de recuperação de senha via WhatsApp
// @Summary Recuperação de senha via WhatsApp
// @Description Envia uma mensagem de WhatsApp com código para recuperação de senha
// @Tags client-auth
// @Accept json
// @Produce json
// @Param request body services.ForgotPasswordRequest true "Telefone do usuário"
// @Success 200 {object} SuccessResponse "Mensagem enviada com sucesso"
// @Failure 400 {object} ErrorResponse "Dados inválidos"
// @Failure 404 {object} ErrorResponse "Telefone não encontrado"
// @Failure 429 {object} ErrorResponse "Muitas requisições"
// @Failure 500 {object} ErrorResponse "Erro interno do servidor"
// @Router /api/v1/client/auth/forgot-password/whatsapp [post]
func (c *ClientAuthController) ForgotPasswordWhatsApp(ctx *gin.Context) {
	var req services.ForgotPasswordRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(ctx, http.StatusBadRequest, "INVALID_REQUEST", "Formato da requisição inválido", nil)
		return
	}

	// Validamos os dados
	if req.Phone == "" {
		utils.SendErrorResponse(ctx, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Telefone não fornecido", map[string]interface{}{
			"phone": "Telefone é obrigatório",
		})
		return
	}

	// Adicionamos informações do cliente para auditoria
	req.ClientIP = ctx.ClientIP()
	req.UserAgent = ctx.GetHeader("User-Agent")

	// Enviamos a mensagem de WhatsApp
	err := c.AuthService.ForgotPasswordWhatsApp(req)
	if err != nil {
		switch err {
		case services.ErrPhoneNotFound:
			utils.SendErrorResponse(ctx, http.StatusNotFound, "PHONE_NOT_FOUND", "Não existe usuário com este telefone", nil)
		case services.ErrUserInactive:
			utils.SendErrorResponse(ctx, http.StatusForbidden, "USER_INACTIVE", "Usuário inativo", nil)
		case services.ErrTooManyRequests:
			utils.SendErrorResponse(ctx, http.StatusTooManyRequests, "TOO_MANY_REQUESTS", "Muitas solicitações em um curto período", nil)
		default:
			utils.SendErrorResponse(ctx, http.StatusInternalServerError, "SERVER_ERROR", "Erro ao enviar mensagem de WhatsApp", nil)
		}
		return
	}

	// Retornamos sucesso
	utils.SendSuccessResponse(ctx, http.StatusOK, nil, map[string]interface{}{
		"message": "Mensagem de WhatsApp enviada com sucesso",
	})
}

// ValidateResetToken valida um token de recuperação de senha
// @Summary Validação de token de recuperação
// @Description Verifica se um token de recuperação de senha é válido
// @Tags client-auth
// @Produce json
// @Param token path string true "Token de recuperação"
// @Success 200 {object} SuccessResponse "Token válido"
// @Failure 400 {object} ErrorResponse "Token inválido"
// @Failure 500 {object} ErrorResponse "Erro interno do servidor"
// @Router /api/v1/client/auth/reset-password/validate/{token} [get]
func (c *ClientAuthController) ValidateResetToken(ctx *gin.Context) {
	token := ctx.Param("token")

	if token == "" {
		utils.SendErrorResponse(ctx, http.StatusBadRequest, "INVALID_TOKEN", "Token não fornecido", nil)
		return
	}

	// Validamos o token
	err := c.AuthService.ValidateResetToken(token)
	if err != nil {
		utils.SendErrorResponse(ctx, http.StatusBadRequest, "INVALID_TOKEN", "Token inválido ou expirado", nil)
		return
	}

	// Retornamos sucesso
	utils.SendSuccessResponse(ctx, http.StatusOK, nil, map[string]interface{}{
		"valid": true,
	})
}

// ResetPassword redefine a senha de um usuário
// @Summary Redefinição de senha
// @Description Redefine a senha de um usuário usando um token de recuperação
// @Tags client-auth
// @Accept json
// @Produce json
// @Param request body services.ResetPasswordRequest true "Nova senha e token"
// @Success 200 {object} SuccessResponse "Senha redefinida com sucesso"
// @Failure 400 {object} ErrorResponse "Dados inválidos"
// @Failure 401 {object} ErrorResponse "Token inválido"
// @Failure 422 {object} ErrorResponse "Validação falhou"
// @Failure 500 {object} ErrorResponse "Erro interno do servidor"
// @Router /api/v1/client/auth/reset-password [post]
func (c *ClientAuthController) ResetPassword(ctx *gin.Context) {
	var req services.ResetPasswordRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(ctx, http.StatusBadRequest, "INVALID_REQUEST", "Formato da requisição inválido", nil)
		return
	}

	// Validamos os dados
	if req.Token == "" || req.Password == "" || req.ConfirmPassword == "" {
		utils.SendErrorResponse(ctx, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Campos obrigatórios não preenchidos", map[string]interface{}{
			"token":            "Token é obrigatório",
			"password":         "Senha é obrigatória",
			"confirm_password": "Confirmação de senha é obrigatória",
		})
		return
	}

	// Verificamos se as senhas são iguais
	if req.Password != req.ConfirmPassword {
		utils.SendErrorResponse(ctx, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Senhas não conferem", map[string]interface{}{
			"confirm_password": "Senhas não conferem",
		})
		return
	}

	// Redefinimos a senha
	err := c.AuthService.ResetPassword(req)
	if err != nil {
		switch err {
		case services.ErrInvalidToken:
			utils.SendErrorResponse(ctx, http.StatusUnauthorized, "INVALID_TOKEN", "Token inválido ou expirado", nil)
		case services.ErrPasswordTooWeak:
			utils.SendErrorResponse(ctx, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Senha muito fraca", map[string]interface{}{
				"password": "A senha deve conter letras maiúsculas, minúsculas, números e caracteres especiais",
			})
		case services.ErrPasswordConfirmation:
			utils.SendErrorResponse(ctx, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Senhas não conferem", map[string]interface{}{
				"confirm_password": "Senhas não conferem",
			})
		case services.ErrUserInactive:
			utils.SendErrorResponse(ctx, http.StatusForbidden, "USER_INACTIVE", "Usuário inativo", nil)
		default:
			utils.SendErrorResponse(ctx, http.StatusInternalServerError, "SERVER_ERROR", "Erro ao redefinir senha", nil)
		}
		return
	}

	// Retornamos sucesso
	utils.SendSuccessResponse(ctx, http.StatusOK, nil, map[string]interface{}{
		"message": "Senha redefinida com sucesso",
	})
}

// RegisterRoutes registra as rotas do controlador
func (c *ClientAuthController) RegisterRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", c.Register)
		auth.POST("/login", c.Login)
		auth.POST("/refresh", c.RefreshToken)
		auth.POST("/forgot-password/email", c.ForgotPasswordEmail)
		auth.POST("/forgot-password/sms", c.ForgotPasswordSMS)
		auth.POST("/forgot-password/whatsapp", c.ForgotPasswordWhatsApp)
		auth.GET("/reset-password/validate/:token", c.ValidateResetToken)
		auth.POST("/reset-password", c.ResetPassword)
	}
}
