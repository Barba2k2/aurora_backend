package repositories

import (
	"errors"
	"time"

	"github.com/Barba2k2/aurora_backend/src/models"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

var (
	ErrTokenNotFound     = errors.New("token not found")
	ErrTokenAlreadyUser  = errors.New("token already used")
	ErrTokenExpired      = errors.New("token expired")
	ErrTokenRevoked      = errors.New("token revoked")
	ErrTokenRateLimitHit = errors.New("token rate limit hit")
)

type TokenRepoitoryInterface interface {
	Create(token *models.PasswordResetToken) error
	FindByToken(token string) (*models.PasswordResetToken, error)
	FindUserAndChannel(userID uuid.UUID, channel models.TokenChannel) ([]*models.PasswordResetToken, error)
	InvalidateAllUserToken(userID uuid.UUID) error
	InvalidateToken(tokenID uuid.UUID) error
	MarkTokenAsUsed(tokenID uuid.UUID) error
	IncrementFailedAttempts(tokenID uuid.UUID) error
	CountActiveTokensByUser(userID uuid.UUID, timeWindow time.Duration) (int, error)
}

type TokenRepository struct {
	DB *gorm.DB
}

func NewTokenRepository(db *gorm.DB) TokenRepoitoryInterface {
	return &TokenRepository{DB: db}
}
