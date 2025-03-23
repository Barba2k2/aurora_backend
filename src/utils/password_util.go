package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Definition of constants for passwords
const (
	// DefaultBcryptCost is the default cost for the bcrypt algorithm
	DefaultBcryptCost = 12
	// MinPasswordLength is the minimum password length
	MinPasswordLength = 8
	// MaxPasswordLength is the maximum password length
	MaxPasswordLength = 72
	// NumericCodeLength is the default length for numeric codes (SMS, WhatsApp)
	NumericCodeLength = 6
)

var (
	// ErrPasswordTooShort indicates that the password is too short
	ErrPasswordTooShort = fmt.Errorf("password must be at least %d characters", MinPasswordLength)
	// ErrPasswordTooLong indicates that the password is too long
	ErrPasswordTooLong = fmt.Errorf("password must be at most %d characters", MaxPasswordLength)
	// ErrPasswordTooWeak indicates that the password is too weak
	ErrPasswordTooWeak = errors.New("password is too weak, it should include uppercase, lowercase, numbers and special characters")
)

// PasswordUtil provides functions for working with passwords
type PasswordUtil struct {
	BcryptCost int
}

// NewPasswordUtil creates a new instance of PasswordUtil
func NewPasswordUtil(bcryptCost int) *PasswordUtil {
	if bcryptCost <= 0 {
		bcryptCost = DefaultBcryptCost
	}
	return &PasswordUtil{
		BcryptCost: bcryptCost,
	}
}

// ValidatePasswordLength validates the password length
func (p *PasswordUtil) ValidatePasswordLength(password string) error {
	if len(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}
	if len(password) > MaxPasswordLength {
		return ErrPasswordTooLong
	}
	return nil
}

// HashPassword generates a bcrypt hash for the password
func (p *PasswordUtil) HashPassword(password string) (string, error) {
	// We validate the password length
	if err := p.ValidatePasswordLength(password); err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), p.BcryptCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// VerifyPassword checks if the password matches the hash
func (p *PasswordUtil) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// ValidatePasswordStrength checks if the password is strong enough
func (p *PasswordUtil) ValidatePasswordStrength(password string) error {
	if err := p.ValidatePasswordLength(password); err != nil {
		return err
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*()-_[]{}|;:,.<>?/", char):
			hasSpecial = true
		}
	}

	// We require at least 3 of the 4 character types
	score := 0
	if hasUpper {
		score++
	}
	if hasLower {
		score++
	}
	if hasNumber {
		score++
	}
	if hasSpecial {
		score++
	}

	if score < 3 {
		return ErrPasswordTooWeak
	}

	return nil
}

// GenerateRandomToken generates a random token
func (p *PasswordUtil) GenerateRandomToken(length int) (string, error) {
	if length <= 0 {
		length = 32
	}

	// Generate random bytes
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// Encode in base64 and remove problematic characters
	token := base64.URLEncoding.EncodeToString(b)
	token = strings.ReplaceAll(token, "+", "")
	token = strings.ReplaceAll(token, "/", "")
	token = strings.ReplaceAll(token, "=", "")

	// Truncate to the desired length
	if len(token) > length {
		token = token[:length]
	}

	return token, nil
}

// GenerateNumericCode generates a random numeric code (for SMS, WhatsApp)
func (p *PasswordUtil) GenerateNumericCode(length int) (string, error) {
	if length <= 0 {
		length = NumericCodeLength
	}

	// Generate random digits
	var sb strings.Builder
	for i := 0; i < length; i++ {
		// Generate a number between 0 and 9
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		sb.WriteString((num.String()))
	}

	return sb.String(), nil
}
