package repositories

import (
	"errors"
	"time"

	"github.com/Barba2k2/aurora_backend/src/models"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

// Common errors related to tokens
var (
	ErrTokenNotFound     = errors.New("token not found")
	ErrTokenAlreadyUsed  = errors.New("token already used")
	ErrTokenExpired      = errors.New("token expired")
	ErrTokenRevoked      = errors.New("token revoked")
	ErrTokenRateLimitHit = errors.New("token rate limit hit")
)

// TokenRepositoryInterface defines the interface for accessing token data
type TokenRepositoryInterface interface {
	Create(token *models.PasswordResetToken) error
	FindByToken(token string) (*models.PasswordResetToken, error)
	FindByUserAndChannel(userID uuid.UUID, channel models.TokenChannel) ([]*models.PasswordResetToken, error)
	InvalidateAllUserTokens(userID uuid.UUID) error
	InvalidateToken(tokenID uuid.UUID) error
	MarkTokenAsUsed(tokenID uuid.UUID) error
	IncrementFailedAttempts(tokenID uuid.UUID) error
	CountActiveTokensByUser(userID uuid.UUID, timeWindow time.Duration) (int, error)
}

// TokenRepository implements the TokenRepositoryInterface
type TokenRepository struct {
	DB *gorm.DB
}

// NewTokenRepository creates a new instance of TokenRepository
func NewTokenRepository(db *gorm.DB) TokenRepositoryInterface {
	return &TokenRepository{DB: db}
}

// Create creates a new token in the database
func (r *TokenRepository) Create(token *models.PasswordResetToken) error {
	// We define creation/update timestamps
	now := time.Now()
	token.CreatedAt = now
	token.UpdatedAt = now

	// We create the token
	return r.DB.Create(token).Error
}

// FindByToken finds a token by its token value
func (r *TokenRepository) FindByToken(token string) (*models.PasswordResetToken, error) {
	var passwordResetToken models.PasswordResetToken

	if err := r.DB.Where("token = ?", token).First(&passwordResetToken).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, ErrTokenNotFound
		}
		return nil, err
	}

	// We check the token status
	switch passwordResetToken.Status {
	case models.TokenStatusUsed:
		return &passwordResetToken, ErrTokenAlreadyUsed
	case models.TokenStatusExpired:
		return &passwordResetToken, ErrTokenExpired
	case models.TokenStatusRevoked:
		return &passwordResetToken, ErrTokenRevoked
	}

	// We check if the token has expired
	if time.Now().After(passwordResetToken.ExpiresAt) {
		// We update the status to expired
		passwordResetToken.MarkAsExpired()
		r.DB.Save(&passwordResetToken)
		return &passwordResetToken, ErrTokenExpired
	}

	return &passwordResetToken, nil
}

// FindByUserAndChannel finds tokens for a specific user and channel
func (r *TokenRepository) FindByUserAndChannel(userID uuid.UUID, channel models.TokenChannel) ([]*models.PasswordResetToken, error) {
	var tokens []*models.PasswordResetToken

	if err := r.DB.Where("user_id = ? AND channel = ?", userID, channel).Find(&tokens).Error; err != nil {
		return nil, err
	}

	return tokens, nil
}

// InvalidateAllUserTokens invalidates all active tokens for a user
func (r *TokenRepository) InvalidateAllUserTokens(userID uuid.UUID) error {
	now := time.Now()

	return r.DB.Model(&models.PasswordResetToken{}).
		Where("user_id = ? AND status = ?", userID, models.TokenStatusActive).
		Updates(map[string]interface{}{
			"status":     models.TokenStatusExpired,
			"updated_at": now,
		}).Error
}

// InvalidateToken invalidates a specific token
func (r *TokenRepository) InvalidateToken(tokenID uuid.UUID) error {
	var token models.PasswordResetToken

	if err := r.DB.Where("id = ?", tokenID).First(&token).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return ErrTokenNotFound
		}
		return err
	}

	token.Status = models.TokenStatusRevoked
	token.UpdatedAt = time.Now()

	return r.DB.Save(&token).Error
}

// MarkTokenAsUsed marks a token as used
func (r *TokenRepository) MarkTokenAsUsed(tokenID uuid.UUID) error {
	var token models.PasswordResetToken

	if err := r.DB.Where("id = ?", tokenID).First(&token).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return ErrTokenNotFound
		}
		return err
	}

	token.MarkAsUsed()

	return r.DB.Save(&token).Error
}

// IncrementFailedAttempts increments the failed attempts counter for a token
func (r *TokenRepository) IncrementFailedAttempts(tokenID uuid.UUID) error {
	var token models.PasswordResetToken

	if err := r.DB.Where("id = ?", tokenID).First(&token).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return ErrTokenNotFound
		}
		return err
	}

	token.IncrementFailedAttempts()

	return r.DB.Save(&token).Error
}

// CountActiveTokensByUser counts active tokens created by a user within a time period
func (r *TokenRepository) CountActiveTokensByUser(userID uuid.UUID, timeWindow time.Duration) (int, error) {
	var count int

	// We define the time period to check the number of tokens created
	fromTime := time.Now().Add(-timeWindow)

	if err := r.DB.Model(&models.PasswordResetToken{}).
		Where("user_id = ? AND created_at > ?", userID, fromTime).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}
