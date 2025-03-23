package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// WhatsAppService implements the WhatsAppServiceInterface
type WhatsAppService struct {
	Config WhatsAppConfig
}

// NewWhatsAppService creates a new instance of the WhatsApp service
func NewWhatsAppService(config WhatsAppConfig) WhatsAppServiceInterface {
	return &WhatsAppService{
		Config: config,
	}
}

// SendPasswordResetWhatsApp sends a WhatsApp message for password recovery
func (s *WhatsAppService) SendPasswordResetWhatsApp(phone, name, code string) error {
	message := fmt.Sprintf("Hello %s, your password recovery code is: %s. Valid for 5 minutes.", name, code)
	return s.SendGenericWhatsApp(phone, message)
}

// SendGenericWhatsApp sends a generic WhatsApp message
func (s *WhatsAppService) SendGenericWhatsApp(phone, message string) error {
	switch s.Config.Provider {
	case "meta":
		return s.sendMetaWhatsApp(phone, message)
	case "twilio":
		return s.sendTwilioWhatsApp(phone, message)
	default:
		return ErrProviderNotFound
	}
}

// sendMetaWhatsApp sends a WhatsApp message via Meta API (formerly Facebook)
func (s *WhatsAppService) sendMetaWhatsApp(phone, message string) error {
	// Meta API for WhatsApp
	apiURL := fmt.Sprintf("https://graph.facebook.com/v17.0/%s/messages", s.Config.PhoneNumberID)
	
	// Create the payload for the API
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type": "individual",
		"to": phone,
		"type": "text",
		"text": map[string]string{
			"preview_url": "false",
			"body": message,
		},
	}
	
	// Convert to JSON
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendingWhatsApp, err)
	}
	
	// Create the request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadJSON))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendingWhatsApp, err)
	}
	
	// Add headers
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+s.Config.AccessToken)
	
	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendingWhatsApp, err)
	}
	defer resp.Body.Close()
	
	// Check the response status
	if resp.StatusCode >= 400 {
		return fmt.Errorf("%w: status code %d", ErrSendingWhatsApp, resp.StatusCode)
	}
	
	return nil
}

// sendTwilioWhatsApp sends a WhatsApp message via Twilio
func (s *WhatsAppService) sendTwilioWhatsApp(phone, message string) error {
	// Twilio API
	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", s.Config.AccountSID)
	
	// Prepare form data
	formData := map[string]string{
		"To":   "whatsapp:" + phone,
		"From": "whatsapp:" + s.Config.FromNumber,
		"Body": message,
	}
	
	// Convert to form values
	formValues := &bytes.Buffer{}
	for key, value := range formData {
		if formValues.Len() > 0 {
			formValues.WriteString("&")
		}
		formValues.WriteString(fmt.Sprintf("%s=%s", key, value))
	}
	
	// Create the request
	req, err := http.NewRequest("POST", apiURL, formValues)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendingWhatsApp, err)
	}
	
	// Add headers
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(s.Config.AccountSID, s.Config.AuthToken)
	
	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendingWhatsApp, err)
	}
	defer resp.Body.Close()
	
	// Check the response status
	if resp.StatusCode >= 400 {
		return fmt.Errorf("%w: status code %d", ErrSendingWhatsApp, resp.StatusCode)
	}
	
	return nil
}