package utils

import (
	"errors"
	"fmt"
)

// Definicao de constantes para senhas
const (
	// Custo padrao para o algoritmo bcrypt
	DefaultBcryptCost = 12
	// Minimo de caracteres para a senha
	MinPasswordLength = 8
	// Maximo de caracteres para a senha
	MaxPasswordLength = 72
	// Tamanho do codigo numerico
	NumericCodeLength = 6
)

var (
	// ErrPasswordTooShort indica que a senha é muito curta
	ErrPasswordTooShort = fmt.Errorf("password must be at least %d characters", MinPasswordLength)
	// ErrPasswordTooLong indica que a senha é muito longa
	ErrPasswordTooLong = fmt.Errorf("password must be at most %d characters", MaxPasswordLength)
	// ErrPasswordTooWeak indica que a senha é muito fraca
	ErrPasswordTooWeak = errors.New("password is too weak, it should include uppercase, lowercase, numbers and special characters")
)

// PasswordUtil fornece funcoes para trabalhar com senhas
type PasswordUtil struct {
	BcryptCost int
}

// NewPasswordUtil cria uma nova instancia de PasswordUtil
func NewPasswordUtil(bcryptCost int) *PasswordUtil {
	if bcryptCost <= 0 {
		bcryptCost = DefaultBcryptCost
	}
	return &PasswordUtil{
		BcryptCost: bcryptCost,
	}
}



