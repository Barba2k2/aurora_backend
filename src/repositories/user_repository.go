package repositories

import (
	"errors"
	"time"

	"github.com/Barba2k2/aurora_backend/src/models"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

// Common errors related to users
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// UserRepository defines the interface for accessing user data
type UserRepository interface {
	// Basic CRUD operations
	Create(user *models.User) error
	FindByID(id uuid.UUID) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	FindByPhone(phone string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uuid.UUID, deletedBy uuid.UUID) error
	
	// Authentication operations
	UpdateLastLogin(id uuid.UUID) error
	IncrementFailedLoginCount(id uuid.UUID) error
	ResetFailedLoginCount(id uuid.UUID) error
	
	// For clients
	FindAllClients(page, limit int, filters map[string]interface{}) ([]*models.User, int64, error)
	
	// For professionals
	FindAllProfessionals(page, limit int, filters map[string]interface{}) ([]*models.User, int64, error)
	
	// For establishments
	CreateEstablishment(establishment *models.Establishment) error
	FindEstablishmentByUserID(userID uuid.UUID) (*models.Establishment, error)
	UpdateEstablishment(establishment *models.Establishment) error
}

// UserRepositoryImpl implements the UserRepository interface
type UserRepositoryImpl struct {
	DB *gorm.DB
}

// NewUserRepository creates a new instance of UserRepository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &UserRepositoryImpl{DB: db}
}

// Create creates a new user in the database
func (r *UserRepositoryImpl) Create(user *models.User) error {
	// We check if a user with this email already exists
	var count int
	if err := r.DB.Model(&models.User{}).Where("email = ?", user.Email).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrUserAlreadyExists
	}
	
	// We check if a user with this phone number already exists (if provided)
	if user.Phone != "" {
		count = 0
		if err := r.DB.Model(&models.User{}).Where("phone = ?", user.Phone).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrUserAlreadyExists
		}
	}
	
	// We define creation/update timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	
	// We create the user
	return r.DB.Create(user).Error
}

// FindByID finds a user by ID
func (r *UserRepositoryImpl) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	
	// We only search for active users by default
	if err := r.DB.Where("id = ? AND status = ?", id, models.UserStatusActive).First(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	
	return &user, nil
}

// FindByEmail finds a user by email
func (r *UserRepositoryImpl) FindByEmail(email string) (*models.User, error) {
	var user models.User
	
	// We only search for active users by default
	if err := r.DB.Where("email = ? AND status = ?", email, models.UserStatusActive).First(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	
	return &user, nil
}

// FindByPhone finds a user by phone number
func (r *UserRepositoryImpl) FindByPhone(phone string) (*models.User, error) {
	var user models.User
	
	// We only search for active users by default
	if err := r.DB.Where("phone = ? AND status = ?", phone, models.UserStatusActive).First(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	
	return &user, nil
}

// Update updates a user's data
func (r *UserRepositoryImpl) Update(user *models.User) error {
	// We update the timestamp
	user.UpdatedAt = time.Now()
	
	// We check if the user exists
	if err := r.DB.First(&models.User{}, "id = ?", user.ID).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return ErrUserNotFound
		}
		return err
	}
	
	// We update the user
	return r.DB.Save(user).Error
}

// Delete performs a soft delete of the user
func (r *UserRepositoryImpl) Delete(id uuid.UUID, deletedBy uuid.UUID) error {
	// We check if the user exists
	var user models.User
	if err := r.DB.First(&user, "id = ?", id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return ErrUserNotFound
		}
		return err
	}
	
	// We update the status and soft delete fields
	now := time.Now()
	return r.DB.Model(&user).Updates(map[string]interface{}{
		"status":     models.UserStatusInactive,
		"deleted_at": now,
		"deleted_by": deletedBy,
		"updated_at": now,
	}).Error
}

// UpdateLastLogin updates the last login timestamp
func (r *UserRepositoryImpl) UpdateLastLogin(id uuid.UUID) error {
	now := time.Now()
	return r.DB.Model(&models.User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_login_at": now,
		"updated_at":    now,
	}).Error
}

// IncrementFailedLoginCount increments the failed login counter
func (r *UserRepositoryImpl) IncrementFailedLoginCount(id uuid.UUID) error {
	return r.DB.Model(&models.User{}).Where("id = ?", id).
		UpdateColumn("failed_login_count", gorm.Expr("failed_login_count + 1")).
		UpdateColumn("updated_at", time.Now()).
		Error
}

// ResetFailedLoginCount resets the failed login counter
func (r *UserRepositoryImpl) ResetFailedLoginCount(id uuid.UUID) error {
	return r.DB.Model(&models.User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"failed_login_count": 0,
		"updated_at":         time.Now(),
	}).Error
}

// FindAllClients returns all clients with pagination and filters
func (r *UserRepositoryImpl) FindAllClients(page, limit int, filters map[string]interface{}) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64
	
	// Add role filter for clients
	filters["role"] = models.UserRoleClient
	
	// Configure the base query
	query := r.DB.Model(&models.User{})
	
	// Apply filters
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}
	
	// Count the total number of records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Apply pagination
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	
	return users, total, nil
}

// FindAllProfessionals returns all professionals with pagination and filters
func (r *UserRepositoryImpl) FindAllProfessionals(page, limit int, filters map[string]interface{}) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64
	
	// Add role filter for professionals
	filters["role"] = models.UserRoleProfessional
	
	// Configure the base query
	query := r.DB.Model(&models.User{})
	
	// Apply filters
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}
	
	// Preload the establishment relationship
	query = query.Preload("Establishment")
	
	// Count the total number of records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Apply pagination
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	
	return users, total, nil
}

// CreateEstablishment creates a new establishment
func (r *UserRepositoryImpl) CreateEstablishment(establishment *models.Establishment) error {
	// Define creation/update timestamps
	now := time.Now()
	establishment.CreatedAt = now
	establishment.UpdatedAt = now
	
	// Create the establishment
	return r.DB.Create(establishment).Error
}

// FindEstablishmentByUserID finds an establishment by user ID
func (r *UserRepositoryImpl) FindEstablishmentByUserID(userID uuid.UUID) (*models.Establishment, error) {
	var establishment models.Establishment
	
	if err := r.DB.Where("user_id = ? AND status = ?", userID, models.UserStatusActive).First(&establishment).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	
	return &establishment, nil
}

// UpdateEstablishment updates an establishment's data
func (r *UserRepositoryImpl) UpdateEstablishment(establishment *models.Establishment) error {
	// Update the timestamp
	establishment.UpdatedAt = time.Now()
	
	// Check if the establishment exists
	if err := r.DB.First(&models.Establishment{}, "id = ?", establishment.ID).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return ErrUserNotFound
		}
		return err
	}
	
	// Update the establishment
	return r.DB.Save(establishment).Error
}