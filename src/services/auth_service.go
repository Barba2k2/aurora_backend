package services

import (
	"errors"
	"net/http"
	"time"

	"github.com/Barba2k2/aurora_backend/src/models"
	"github.com/Barba2k2/aurora_backend/src/repositories"
	"github.com/Barba2k2/aurora_backend/src/utils"
)

// Erros do serviço de autenticação
var (
	ErrInvalidLogin         = errors.New("invalid email or password")
	ErrUserBlocked          = errors.New("user account is blocked due to too many failed login attempts")
	ErrUserInactive         = errors.New("user account is inactive")
	ErrInvalidToken         = errors.New("invalid or expired token")
	ErrTooManyRequests      = errors.New("too many requests, please try again later")
	ErrEmailNotFound        = errors.New("no user found with this email")
	ErrPhoneNotFound        = errors.New("no user found with this phone number")
	ErrPasswordTooWeak      = errors.New("password is too weak")
	ErrPasswordConfirmation = errors.New("password and confirmation do not match")
)

// AuthConfig contém as configurações para o serviço de autenticação
type AuthConfig struct {
	// Limite de tentativas de login
	MaxLoginAttempts int
	// Tempo para bloqueio após exceder tentativas de login
	LoginLockDuration time.Duration
	// Limite de tokens de recuperação de senha por período
	ResetTokenRateLimit int
	// Período para verificação de rate limit
	ResetTokenRateWindow time.Duration
	// Tempo de expiração de token de recuperação via email
	ResetTokenEmailExpiration time.Duration
	// Tempo de expiração de token de recuperação via SMS/WhatsApp
	ResetTokenSMSExpiration time.Duration
}

// DefaultAuthConfig retorna uma configuração padrão para o serviço de autenticação
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		MaxLoginAttempts:          5,
		LoginLockDuration:         1 * time.Hour,
		ResetTokenRateLimit:       3,
		ResetTokenRateWindow:      1 * time.Hour,
		ResetTokenEmailExpiration: 15 * time.Minute,
		ResetTokenSMSExpiration:   5 * time.Minute,
	}
}

// AuthService implementa os serviços de autenticação
type AuthService struct {
	UserRepo        repositories.UserRepository
	TokenRepo       repositories.TokenRepositoryInterface
	PasswordUtil    *utils.PasswordUtil
	JWTUtil         *utils.JWTUtil
	EmailService    EmailServiceInterface
	SMSService      SMSServiceInterface
	WhatsAppService WhatsAppServiceInterface
	Config          AuthConfig
}

// NewAuthService cria uma nova instância do serviço de autenticação
func NewAuthService(
	userRepo repositories.UserRepository,
	tokenRepo repositories.TokenRepositoryInterface,
	passwordUtil *utils.PasswordUtil,
	jwtUtil *utils.JWTUtil,
	emailService EmailServiceInterface,
	smsService SMSServiceInterface,
	whatsAppService WhatsAppServiceInterface,
	config AuthConfig,
) *AuthService {
	return &AuthService{
		UserRepo:        userRepo,
		TokenRepo:       tokenRepo,
		PasswordUtil:    passwordUtil,
		JWTUtil:         jwtUtil,
		EmailService:    emailService,
		SMSService:      smsService,
		WhatsAppService: whatsAppService,
		Config:          config,
	}
}

// RegisterRequest representa os dados de requisição para registro
type RegisterRequest struct {
	Name            string          `json:"name" validate:"required"`
	Email           string          `json:"email" validate:"required,email"`
	Phone           string          `json:"phone" validate:"required"`
	Password        string          `json:"password" validate:"required"`
	ConfirmPassword string          `json:"confirm_password" validate:"required,eqfield=Password"`
	Role            models.UserRole `json:"role" validate:"required,oneof=CLIENT PROFESSIONAL"`
	Timezone        string          `json:"timezone"`
}

// LoginRequest representa os dados de requisição para login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RefreshTokenRequest representa os dados de requisição para refresh token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// ResetPasswordRequest representa os dados de requisição para recuperação de senha
type ResetPasswordRequest struct {
	Token           string `json:"token" validate:"required"`
	Password        string `json:"password" validate:"required"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

// ResetPasswordToken representa os dados de requisição para recuperação de senha
type ResetPasswordToken struct {
	Token           string `json:"token" validate:"required"`
	Password        string `json:"password" validate:"required"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

// ForgotPasswordRequest representa os dados de requisição para solicitação de recuperação de senha
type ForgotPasswordRequest struct {
	Email     string `json:"email" validate:"omitempty,email"`
	Phone     string `json:"phone,omitempty" validate:"omitempty,phone"`
	ClientIP  string `json:"client_ip" validate:"required"`
	UserAgent string
}

// TokenResponse representa a resposta com tokens JWT
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// Register registra um novo usuário
func (s *AuthService) Register(req RegisterRequest) (*models.User, error) {
	// Validamos a senha
	if err := s.PasswordUtil.ValidatePasswordStrength(req.Password); err != nil {
		return nil, ErrPasswordTooWeak
	}

	// Verificamos se as senhas sao iguais
	if req.Password != req.ConfirmPassword {
		return nil, ErrPasswordConfirmation
	}

	// Geramos o hash da senha
	hashedPassword, err := s.PasswordUtil.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Criamos o usuário
	user := &models.User{
		Email:        req.Email,
		Phone:        req.Phone,
		Name:         req.Name,
		PasswordHash: hashedPassword,
		Role:         req.Role,
		Status:       models.UserStatusActive,
		Timezone:     req.Timezone,
	}

	// Salvamos no banco de dados
	if err := s.UserRepo.Create(user); err != nil {
		return nil, err
	}

	// Se for um profissional, criamos tambem o estabelecimento
	if req.Role == models.UserRoleProfessional {
		establishment := &models.Establishment{
			UserID:        user.ID,
			BussinessName: req.Name,
			Timezone:      req.Timezone,
			Status:        models.UserStatusActive,
		}

		if err := s.UserRepo.CreateEstablishment(establishment); err != nil {
			return nil, err
		}
	}

	return user, nil
}

// Login realiza o login de um usuário
func (s *AuthService) Login(req LoginRequest) (*models.User, *TokenResponse, error) {
	// Buscamos o usuario pelo email
	user, err := s.UserRepo.FindByEmail(req.Email)
	if err != nil {
		if err == repositories.ErrUserNotFound {
			// Retornamos erro generico para evitar enumeracao de usuarios
			return nil, nil, ErrInvalidLogin
		}
		return nil, nil, err
	}

	// Variuficamos se o usuario esta ativo
	if user.Status != models.UserStatusActive {
		// Para usuarios bloqueados, informamos explicitamente
		if user.Status == models.UserStatusBlocked {
			return nil, nil, ErrUserBlocked
		}
		return nil, nil, ErrUserInactive
	}

	// Verificamos se o usuario esta bloqueado por tentativas de login
	if user.FailedLoginCount >= s.Config.MaxLoginAttempts {
		return nil, nil, ErrUserBlocked
	}

	// Verificamos a senha
	if err := s.PasswordUtil.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		// Incrementamos o contador de falhas
		s.UserRepo.IncrementFailedLoginCount(user.ID)
		return nil, nil, ErrInvalidLogin
	}

	// Resetamos o contador de falhas e atualizamos o ultimo login
	s.UserRepo.ResetFailedLoginCount(user.ID)
	s.UserRepo.UpdateLastLogin(user.ID)

	// Geramos o par de token
	accessToken, refreshToken, err := s.JWTUtil.GenerateTokenPair(user.ID, user.Role)
	if err != nil {
		return nil, nil, err
	}

	// Criamos a respota
	tokenResponse := &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(utils.TokenExpirationAccess.Seconds()),
	}

	return user, tokenResponse, nil
}

// RefreshToken renova o token de acesso usando um refresh token
func (s *AuthService) RefreshToken(req RefreshTokenRequest) (*TokenResponse, error) {
	// Validamos o refresh token
	claims, err := s.JWTUtil.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	// Buscamos o usuario
	user, err := s.UserRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	// Verificamos se o usuario esta ativo
	if user.Status != models.UserStatusActive {
		return nil, ErrUserInactive
	}

	// Geramos o novo par de tokens
	accessToken, refreshToken, err := s.JWTUtil.GenerateTokenPair(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	// Criamos a resposta
	tokenResponse := &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(utils.TokenExpirationAccess.Seconds()),
	}

	return tokenResponse, nil
}

// ForgotPasswordEmail inicia o processo de recuperação de senha via email
func (s *AuthService) ForgotPasswordEmail(req ForgotPasswordRequest) error {
	// Buscamos o usuario pelo email
	user, err := s.UserRepo.FindByEmail(req.Email)
	if err != nil {
		if err == repositories.ErrUserNotFound {
			return ErrEmailNotFound
		}
		return err
	}

	// Verificamos se o usuario esta ativo
	if user.Status != models.UserStatusActive {
		return ErrUserInactive
	}

	// Verificamos o rate limit
	count, err := s.TokenRepo.CountActiveTokensByUser(user.ID, s.Config.ResetTokenRateWindow)
	if err != nil {
		return nil
	}
	if count >= s.Config.ResetTokenRateLimit {
		return ErrTooManyRequests
	}

	// Invalidamos todos os tokens ativos do usuario
	if err := s.TokenRepo.InvalidateAllUserTokens(user.ID); err != nil {
		return err
	}

	// Geramos um token unico para recuperacao
	resetToken, err := s.PasswordUtil.GenerateRandomToken(32)
	if err != nil {
		return err
	}

	// Criamos o registro do token
	token := &models.PasswordResetToken{
		UserID:    user.ID,
		Token:     resetToken,
		Channel:   models.TokenChannelEmail,
		Status:    models.TokenStatusActive,
		ExpiresAt: time.Now().Add(s.Config.ResetTokenEmailExpiration),
		IPAddress: req.ClientIP,
		UserAgent: req.UserAgent,
	}

	if err := s.TokenRepo.Create(token); err != nil {
		return err
	}

	// Enviamos o email com o token
	return s.EmailService.SendPasswordResetEmail(user.Email, user.Name, resetToken)
}

// ForgotPasswordSMS inicia o processo de recuperação de senha via SMS
func (s *AuthService) ForgotPasswordSMS(req ForgotPasswordRequest) error {
	// Buscamos o usario pelo telefone
	user, err := s.UserRepo.FindByPhone(req.Phone)
	if err != nil {
		if err == repositories.ErrUserNotFound {
			return ErrPhoneNotFound
		}
		return err
	}

	// Verificamos se o usario esta ativo
	if user.Status != models.UserStatusActive {
		return ErrUserInactive
	}

	// Verificamos o rate limit
	count, err := s.TokenRepo.CountActiveTokensByUser(user.ID, s.Config.ResetTokenEmailExpiration)
	if err != nil {
		return err
	}
	if count >= s.Config.ResetTokenRateLimit {
		return ErrTooManyRequests
	}

	// Invalidamos todos os tokens ativos do usuario
	if err := s.TokenRepo.InvalidateAllUserTokens(user.ID); err != nil {
		return nil
	}

	// Geramos um codigo numerico para recuperacao
	code, err := s.PasswordUtil.GenerateNumericCode(6)
	if err != nil {
		return err
	}

	token := &models.PasswordResetToken{
		UserID:    user.ID,
		Token:     code,
		Channel:   models.TokenChannelSMS,
		Status:    models.TokenStatusActive,
		ExpiresAt: time.Now().Add(s.Config.ResetTokenSMSExpiration),
		IPAddress: req.ClientIP,
		UserAgent: req.UserAgent,
	}

	if err := s.TokenRepo.Create(token); err != nil {
		return err
	}

	// Enviamos o SMS com o codigo
	return s.SMSService.SendPasswordResetSMS(user.Phone, code)
}

// ForgotPasswordWhatsApp inicia o processo de recuperação de senha via WhatsApp
func (s *AuthService) ForgotPasswordWhatsApp(req ForgotPasswordRequest) error {
	// Buscamos o usuário pelo telefone
	user, err := s.UserRepo.FindByPhone(req.Phone)
	if err != nil {
		if err == repositories.ErrUserNotFound {
			return ErrPhoneNotFound
		}
		return err
	}

	// Verificamos se o usuário está ativo
	if user.Status != models.UserStatusActive {
		return ErrUserInactive
	}

	// Verificamos o rate limit
	count, err := s.TokenRepo.CountActiveTokensByUser(user.ID, s.Config.ResetTokenRateWindow)
	if err != nil {
		return err
	}
	if count >= s.Config.ResetTokenRateLimit {
		return ErrTooManyRequests
	}

	// Invalidamos todos os tokens ativos do usuário
	if err := s.TokenRepo.InvalidateAllUserTokens(user.ID); err != nil {
		return err
	}

	// Geramos um código numérico para recuperação
	code, err := s.PasswordUtil.GenerateNumericCode(6)
	if err != nil {
		return err
	}

	// Criamos o registro do token
	token := &models.PasswordResetToken{
		UserID:    user.ID,
		Token:     code,
		Channel:   models.TokenChannelWhatsApp,
		Status:    models.TokenStatusActive,
		ExpiresAt: time.Now().Add(s.Config.ResetTokenSMSExpiration),
		IPAddress: req.ClientIP,
		UserAgent: req.UserAgent,
	}

	if err := s.TokenRepo.Create(token); err != nil {
		return err
	}

	// Enviamos a mensagem WhatsApp com o código
	return s.WhatsAppService.SendPasswordResetWhatsApp(user.Phone, user.Name, code)
}

// ValidateResetToken valida um token de recuperação de senha
func (s *AuthService) ValidateResetToken(token string) error {
	// Buscamos o token no banco
	tokenObj, err := s.TokenRepo.FindByToken(token)
	if err != nil {
		return ErrInvalidToken
	}

	// Verificamos se o token eh valido
	if !tokenObj.IsValid() {
		return ErrInvalidToken
	}

	return nil
}

// ResetPassword redefine a senha de um usuário usando o token de recuperação
func (s *AuthService) ResetPassword(req ResetPasswordRequest) error {
	// Validamos a nova senha
	if err := s.PasswordUtil.ValidatePasswordStrength(req.Password); err != nil {
		return ErrPasswordTooWeak
	}

	// Verificamos se a senhas sao iguais
	if req.Password != req.ConfirmPassword {
		return ErrPasswordConfirmation
	}

	// Buscamos o token no banco
	token, err := s.TokenRepo.FindByToken(req.Token)
	if err != nil {
		return ErrInvalidToken
	}

	// Verificamos se o token eh valido
	if !token.IsValid() {
		return ErrInvalidToken
	}

	// Buscamos o usuario
	user, err := s.UserRepo.FindByID(token.UserID)
	if err != nil {
		return err
	}

	// Verificamos se o usuario esta ativo
	if user.Status != models.UserStatusActive {
		return ErrUserInactive
	}

	// Geramos o hash da nova senha
	hashedPassword, err := s.PasswordUtil.HashPassword(req.Password)
	if err != nil {
		return err
	}

	// Atualizamos a senha do usuario
	user.PasswordHash = hashedPassword
	user.FailedLoginCount = 0 // Resetamos o contador de falhas
	if err := s.UserRepo.Update(user); err != nil {
		return err
	}

	// Marcamos o token como usado
	if err := s.TokenRepo.MarkTokenAsUsed(token.ID); err != nil {
		return err
	}

	// Invalidamos todos os tokens ativos do usuário
	return s.TokenRepo.InvalidateAllUserTokens(user.ID)
}

// GetUserFromToken obtem os dados do usuario a partir de um token JWT
func (s *AuthService) GetUserFromToken(tokenString string) (*models.User, error) {
	// Validamos o token
	claims, err := s.JWTUtil.ValidateAccessToken(tokenString)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Buscamos o usuario
	user, err := s.UserRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// ExtractTokenFromRequest extrai o token JWT do cabecalho de Authorization
func (s *AuthService) ExtractTokenFromRequest(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	if len(bearerToken) > 7 && bearerToken[:7] == "Bearer " {
		return bearerToken[7:]
	}
	return ""
}
