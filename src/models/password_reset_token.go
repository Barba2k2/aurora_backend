package models

import (
	"github.com/google/uuid"
	"time"
)

type TokenChannel string

const (
	TokenChannelEmail    TokenChannel = "EMAIL"
	TokenChannelSMS      TokenChannel = "SMS"
	TokenChannelWhatsapp TokenChannel = "WHATSAPP"
)

type TokenStatus string

const (
	TokenStatusActive  TokenStatus = "ACTIVE"
	TokenStatusUsed    TokenStatus = "USED"
	TokenStatusExpired TokenStatus = "EXPIRED"
	TokenStatusRevoked TokenStatus = "REVOKED"
)

type PasswordResetToken struct {
	ID             uuid.UUID    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID         uuid.UUID    `json:"-" gorm:"type:uuid;index"`
	User           User         `json:"-" gorm:"foreignKey:UserID"`
	Token          string       `json:"-" gorm:"type:varchar(255);not null;unique_index"`
	Channel        TokenChannel `json:"channel" gorm:"type:varchar(20);not null"`
	Status         TokenStatus  `json:"status" gorm:"type:varchar(20);not null;default:'ACTIVE'"`
	ExpiresAt      time.Time    `json:"expires_at" gorm:"not null"`
	UsedAt         *time.Time   `json:"used_at,omitempty"`
	FailedAttempts int          `json:"-" gorm:"type:int;default:0"`
	IPAddress      string       `json:"-" gorm:"type:varchar(45)"`
	UserAgent      string       `json:"-" gorm:"type:text"`
	// Audit fields
	CreateAt time.Time `json:"created_at" gorm:"not null"`
	UpdateAt time.Time `json:"updated_at" gorm:"not null"`
}

func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}

func (t *PasswordResetToken) IsValid() bool {
	return t.Status == TokenStatusActive && time.Now().Before(t.ExpiresAt)
}

func (t *PasswordResetToken) MarkAsUsed() {
	now := time.Now()
	t.Status = TokenStatusUsed
	t.UsedAt = &now
	t.UpdateAt = now
}

func (t *PasswordResetToken) MarkAsExpired() {
	t.Status = TokenStatusExpired
	t.UpdateAt = time.Now()
}

func (t *PasswordResetToken) IncrementFailedAttempts() {
	t.FailedAttempts++
	t.UpdateAt = time.Now()

	if t.FailedAttempts >= 5 {
		t.Status = TokenStatusRevoked
	}
}
