package services

import (
	"errors"
	"time"
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
