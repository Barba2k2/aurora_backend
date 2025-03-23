package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"text/template"
)

// EmailService implements the EmailServiceInterface
type EmailService struct {
	Config EmailConfig
}

// NewEmailService creates a new instance of the email service
func NewEmailService(config EmailConfig) EmailServiceInterface {
	return &EmailService{
		Config: config,
	}
}

// sendSMTPEmail sends an email via SMTP
func (s *EmailService) sendSMTPEmail(email, subject, body string) error {
	// SMTP server configuration
	smtpHost := s.Config.Host
	smtpPort := s.Config.Port
	smtpUsername := s.Config.Username
	smtpPassword := s.Config.Password

	// Build the email header
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", s.Config.FromName, s.Config.FromEmail)
	headers["To"] = email
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	// Build the message
	message := ""
	for key, value := range headers {
		message += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	message += "\r\n" + body

	// Authentication
	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)

	// Send the email
	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", smtpHost, smtpPort),
		auth,
		s.Config.FromEmail,
		[]string{email},
		[]byte(message),
	)

	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendingEmail, err)
	}

	return nil
}

// sendSendGridEmail sends an email via SendGrid
func (s *EmailService) sendSendGridEmail(email, subject, body string) error {
	// SendGrid API
	apiURL := "https://api.sendgrid.com/v3/mail/send"

	// Create the payload for the API
	payload := map[string]interface{}{
		"personalizations": []map[string]interface{}{
			{
				"to": []map[string]string{
					{"email": email},
				},
				"subject": subject,
			},
		},
		"from": map[string]string{
			"email": s.Config.FromEmail,
			"name":  s.Config.FromName,
		},
		"content": []map[string]string{
			{
				"type":  "text/html",
				"value": body,
			},
		},
	}

	// Convert to JSON
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendingEmail, err)
	}

	// Create the request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadJSON))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendingEmail, err)
	}

	// Add headers
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+s.Config.APIKey)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendingEmail, err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode >= 400 {
		return fmt.Errorf("%w: status code %d", ErrSendingEmail, resp.StatusCode)
	}

	return nil
}

// sendAWSSESEmail sends an email via AWS SES
func (s *EmailService) sendAWSSESEmail(email, subject, body string) error {
	// Here we would implement the integration with AWS SES
	// Using the AWS SDK for Go

	// For simplicity, let's just return a generic error
	return fmt.Errorf("%w: AWS SES implementation not available", ErrSendingEmail)
}

// SendGenericEmail sends a generic email
func (s *EmailService) SendGenericEmail(email, subject, body string) error {
	switch s.Config.ServiceType {
	case "smtp":
		return s.sendSMTPEmail(email, subject, body)
	case "sendgrid":
		return s.sendSendGridEmail(email, subject, body)
	case "aws_ses":
		return s.sendAWSSESEmail(email, subject, body)
	default:
		return ErrProviderNotFound
	}
}

// SendPasswordResetEmail sends a password reset email
func (s *EmailService) SendPasswordResetEmail(email, name, token string) error {
	subject := "Password Recovery - Scheduling System"

	// Data for the template
	data := map[string]interface{}{
		"Name":     name,
		"Token":    token,
		"ResetURL": fmt.Sprintf("https://seuapp.com/reset-password?token=%s", token),
	}

	// Load the template
	tmpl, err := template.ParseFiles(s.Config.TemplatesDir + "/password_reset.html")
	if err != nil {
		return ErrInvalidTemplate
	}

	// Fill the template
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return ErrInvalidTemplate
	}

	// Send the email
	return s.SendGenericEmail(email, subject, body.String())
}
