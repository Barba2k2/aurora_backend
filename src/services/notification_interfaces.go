package services

import (
	"errors"
)

// Common errors for notification services
var (
	ErrSendingEmail      = errors.New("error sending email")
	ErrSendingSMS        = errors.New("error sending SMS")
	ErrSendingWhatsApp   = errors.New("error sending WhatsApp message")
	ErrInvalidTemplate   = errors.New("invalid template")
	ErrProviderNotFound  = errors.New("notification provider not found")
)

// EmailServiceInterface defines the interface for the email service
type EmailServiceInterface interface {
	SendPasswordResetEmail(email, name, token string) error
	SendGenericEmail(email, subject, body string) error
}

// SMSServiceInterface defines the interface for the SMS service
type SMSServiceInterface interface {
	SendPasswordResetSMS(phone, code string) error
	SendGenericSMS(phone, message string) error
}

// WhatsAppServiceInterface defines the interface for the WhatsApp service
type WhatsAppServiceInterface interface {
	SendPasswordResetWhatsApp(phone, name, code string) error
	SendGenericWhatsApp(phone, message string) error
}

// EmailConfig contains the settings for the email service
type EmailConfig struct {
	Host         string
	Port         int
	Username     string
	Password     string
	FromEmail    string
	FromName     string
	TemplatesDir string
	IsSMTP       bool
	ServiceType  string // "smtp", "sendgrid", "aws_ses"
	APIKey       string // For SendGrid or other API-based services
}

// SMSConfig contains the settings for the SMS service
type SMSConfig struct {
	Provider   string // "twilio", "zenvia"
	AccountSID string // For Twilio
	AuthToken  string // For Twilio
	FromNumber string // Source number
	APIKey     string // For Zenvia
	APISecret  string // For Zenvia
}

// WhatsAppConfig contains the settings for the WhatsApp service
type WhatsAppConfig struct {
	Provider      string // "meta", "twilio" 
	PhoneNumberID string // For Meta API
	AccessToken   string // For Meta API
	AccountSID    string // For Twilio
	AuthToken     string // For Twilio
	FromNumber    string // Source number (with WhatsApp)
}