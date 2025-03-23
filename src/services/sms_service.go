package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// SMSService implements the SMSServiceInterface
type SMSService struct {
	Config SMSConfig
}

// NewSMSService creates a new instance of the SMS service
func NewSMSService(config SMSConfig) SMSServiceInterface {
	return &SMSService{
		Config: config,
	}
}

// SendPasswordResetSMS sends a password recovery SMS
func (s *SMSService) SendPasswordResetSMS(phone, code string) error {
	message := fmt.Sprintf("Your password recovery code is: %s. Valid for 5 minutes.", code)
	return s.SendGenericSMS(phone, message)
}

// SendGenericSMS sends a generic SMS
func (s *SMSService) SendGenericSMS(phone, message string) error {
	switch s.Config.Provider {
	case "twilio":
		return s.sendTwilioSMS(phone, message)
	case "zenvia":
		return s.sendZenviaSMS(phone, message)
	default:
		return ErrProviderNotFound
	}
}

// sendTwilioSMS sends an SMS via Twilio
func (s *SMSService) sendTwilioSMS(phone, message string) error {
	// Twilio API
	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", s.Config.AccountSID)
	
	// Prepare form data
	formData := map[string]string{
		"To":   phone,
		"From": s.Config.FromNumber,
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
		return fmt.Errorf("%w: %v", ErrSendingSMS, err)
	}
	
	// Add headers
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(s.Config.AccountSID, s.Config.AuthToken)
	
	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendingSMS, err)
	}
	defer resp.Body.Close()
	
	// Check the response status
	if resp.StatusCode >= 400 {
		return fmt.Errorf("%w: status code %d", ErrSendingSMS, resp.StatusCode)
	}
	
	return nil
}

// sendZenviaSMS sends an SMS via Zenvia
func (s *SMSService) sendZenviaSMS(phone, message string) error {
	// Zenvia API
	apiURL := "https://api.zenvia.com/v2/channels/sms/messages"
	
	// Create the payload for the API
	payload := map[string]interface{}{
		"from": s.Config.FromNumber,
		"to": phone,
		"contents": []map[string]string{
			{
				"type": "text",
				"text": message,
			},
		},
	}
	
	// Convert to JSON
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendingSMS, err)
	}
	
	// Create the request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadJSON))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendingSMS, err)
	}
	
	// Add headers
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-API-TOKEN", s.Config.APIKey)
	
	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSendingSMS, err)
	}
	defer resp.Body.Close()
	
	// Check the response status
	if resp.StatusCode >= 400 {
		return fmt.Errorf("%w: status code %d", ErrSendingSMS, resp.StatusCode)
	}
	
	return nil
}