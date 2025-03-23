package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type UserRole string

const (
	UserRoleClient       UserRole = "CLIENT"
	UserRoleProfessional UserRole = "PROFESSIONAL"
	UserRoleStaff        UserRole = "STAFF"
	UserRoleAdmin        UserRole = "ADMIN"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "ACTIVE"
	UserStatusInactive UserStatus = "INACTIVE"
	UserStatusPending  UserStatus = "PENDING"
	UserStatusBlocked  UserStatus = "BLOCKED"
)

type User struct {
	ID                uuid.UUID      `json:"id" gorm:"type:uuid;primary_key; default:gen_random_uuid()"`
	Email             string         `json:"email" gorm:"type:varchar(255);unique_index;not null"`
	Phone             string         `json:"phone" gorm:"type:varchar(20);index"`
	Name              string         `json:"name" gorm:"type:varchar(255);not null"`
	PasswordHash      string         `json:"-" gorm:"type:varchar(255);not null"`
	Role              UserRole       `json:"role" gorm:"type:varchar(20);not null"`
	Status            UserStatus     `json:"status" gorm:"type:varchar(20);not null;default:'ACTIVE'"`
	Timezone          string         `json:"timezone" gorm:"type:varchar(50);not null;default:'UTC'"`
	ProfileImageURL   string         `json:"profile_image_url,omitempty" gorm:"type:varchar(255)"`
	PushSubscriptions pq.StringArray `json:"-" gorm:"type:text[]"`
	FailedLoginCount  int            `json:"-" gorm:"type:int;dafult:0"`
	LastLoginAt       *time.Time     `json:"last_login_at,omitempty"`

	CreatedAt time.Time  `json:"created_at" gorm:"not null"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"not null"`
	DeletedAt *time.Time `json:"-" gorm:"index"`
	DeletedBy *uuid.UUID `json:"-" gorm:"type:uuid"`

	Estabilishment *Establishment `json:"establishment,omitempty" gorm:"foreignKey:UserID"`
}

type Establishment struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID         uuid.UUID  `json:"-" gorm:"type:uuid;not null"`
	BussinessName  string     `json:"bussiness_name" gorm:"type:varchar(255);not null"`
	Description    string     `json:"description,omitempty" gorm:"type:text"`
	Address        string     `json:"address,omitempty" gorm:"type:varchar(255)"`
	City           string     `json:"city,omitempty" gorm:"type:varchar(255)"`
	State          string     `json:"state,omitempty" gorm:"type:varchar(255)"`
	Country        string     `json:"country,omitempty" gorm:"type:varchar(255)"`
	ZipCode        string     `json:"zip_code,omitempty" gorm:"type:varchar(20)"`
	BussinessPhone string     `json:"bussiness_phone,omitempty" gorm:"type:varchar(20)"`
	BussinessEmail string     `json:"bussiness_email,omitempty" gorm:"type:varchar(255)"`
	LogoURL        string     `json:"logo_url,omitempty" gorm:"type:varchar(255)"`
	WebsiteURL     string     `json:"website_url,omitempty" gorm:"type:varchar(255)"`
	Timezone       string     `json:"timezone" gorm:"type:varchar(50);not null;default:'UTC'"`
	Status         UserStatus `json:"status" gorm:"type:varchar(20);not null;default:'ACTIVE'"`

	CreatedAt time.Time  `json:"created_at" gorm:"not null"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"not null"`
	DeletedAt *time.Time `json:"-" gorm:"index"`
	DeletedBy *uuid.UUID `json:"-" gorm:"type:uuid"`
}

func (User) TableName() string {
	return "users"
}

func (Establishment) TableName() string {
	return "estabilishments"
}
