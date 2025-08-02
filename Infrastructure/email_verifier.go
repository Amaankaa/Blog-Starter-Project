package infrastructure

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type MailboxLayerVerifier struct {
	APIKey string
}

type mailboxResponse struct {
	MXFound   bool    `json:"mx_found"`
	SMTPCheck bool    `json:"smtp_check"`
	Score     float64 `json:"score"`
}

func NewMailboxLayerVerifier() (*MailboxLayerVerifier, error) {
	apiKey := os.Getenv("MAILBOXLAYER_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("MAILBOXLAYER_API_KEY is not set")
	}
	return &MailboxLayerVerifier{APIKey: apiKey}, nil
}

func (m *MailboxLayerVerifier) IsRealEmail(email string) (bool, error) {
	url := fmt.Sprintf("https://apilayer.net/api/check?access_key=%s&email=%s&smtp=1&format=1", m.APIKey, email)

	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var result mailboxResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	return result.MXFound && result.SMTPCheck, nil
}