package utils

import (
	"errors"
	"time"

	"github.com/Barba2k2/aurora_backend/src/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

var (
	// ErrInvalidToken indicates that the JWT token is invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken indicates that the JWT token has expired
	ErrExpiredToken = errors.New("token has expired")
)

// Token duration definitions
const (
	// TokenExpirationAccess is the duration of the access token (15 minutes)
	TokenExpirationAccess = 15 * time.Minute
	// TokenExpirationRefresh is the duration of the refresh token (7 days)
	TokenExpirationRefresh = 7 * 24 * time.Hour
)

// JWTConfig contains the configuration for JWT
type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	Issuer        string
}

type JWTUtil struct {
	Config JWTConfig
}

// NewJWTUtil creates a new instance of JWTUtil
func NewJWTUtil(config JWTConfig) *JWTUtil {
	return &JWTUtil{
		Config: config,
	}
}

// Claims represents the data included in the JWT
type Claims struct {
	UserID uuid.UUID       `json:"user_id"`
	Role   models.UserRole `json:"role"`
	Type   string          `json:"type"`
	jwt.StandardClaims
}

// GenerateAccessToken generates a new JWT access token
func (j *JWTUtil) GenerateAccessToken(userID uuid.UUID, role models.UserRole) (string, error) {
	now := time.Now()
	expirationTime := now.Add(TokenExpirationAccess)

	claims := Claims{
		UserID: userID,
		Role:   role,
		Type:   "access",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  now.Unix(),
			Issuer:    j.Config.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.Config.AccessSecret))
}

// GenerateRefreshToken generates a new JWT refresh token
func (j *JWTUtil) GenerateRefreshToken(userID uuid.UUID, role models.UserRole) (string, error) {
	now := time.Now()
	expirationTime := now.Add(TokenExpirationRefresh)

	claims := Claims{
		UserID: userID,
		Role:   role,
		Type:   "refresh",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  now.Unix(),
			Issuer:    j.Config.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	return token.SignedString([]byte(j.Config.RefreshSecret))
}

// ValidateAccessToken validates an access token
func (j *JWTUtil) ValidateAccessToken(tokenString string) (*Claims, error) {
	return j.validateToken(tokenString, j.Config.AccessSecret, "access")
}

// ValidateRefreshToken validates a refresh token
func (j *JWTUtil) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return j.validateToken(tokenString, j.Config.RefreshSecret, "refresh")
}

// validateToken validates a JWT token
func (j *JWTUtil) validateToken(tokenString, secret, tokenType string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, ErrInvalidToken
		}
		ve, ok := err.(*jwt.ValidationError)
		if ok && ve.Errors&jwt.ValidationErrorExpired != 0 {
			return nil, ErrExpiredToken
		}
		return nil, err
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.Type != tokenType {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GenerateTokenPair generates a pair of tokens (access and refresh)
func (j *JWTUtil) GenerateTokenPair(userID uuid.UUID, role models.UserRole) (accessToken, refreshToken string, err error) {
	accessToken, err = j.GenerateAccessToken(userID, role)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = j.GenerateRefreshToken(userID, role)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
