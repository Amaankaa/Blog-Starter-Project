package infrastructure

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

type EmailListVerifyVerifier struct {
	APIKey string
}

type emailListVerifyResponse struct {
	Email       string `json:"email"`
	Status      string `json:"status"`
	Result      string `json:"result"`
	Reason      string `json:"reason"`
	Disposable  bool   `json:"disposable"`
	AcceptAll   bool   `json:"accept_all"`
	MXRecord    bool   `json:"mx_record"`
	SMTPCheck   bool   `json:"smtp_check"`
	Deliverable bool   `json:"deliverable"`
}

func NewEmailListVerifyVerifier() (*EmailListVerifyVerifier, error) {
	apiKey := os.Getenv("EMAILLISTVERIFY_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("EMAILLISTVERIFY_API_KEY is not set")
	}
	return &EmailListVerifyVerifier{APIKey: apiKey}, nil
}

func (e *EmailListVerifyVerifier) IsRealEmail(email string) (bool, error) {
	// EmailListVerify API endpoint - try the correct endpoint
	baseURL := "https://apps.emaillistverify.com/api/verifyEmail"

	// Create URL with parameters - EmailListVerify uses 'secret' and 'email' parameters
	params := url.Values{}
	params.Add("secret", e.APIKey)
	params.Add("email", email)

	requestURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// Debug: Print the request URL (remove API key for security)
	debugURL := fmt.Sprintf("%s?secret=***&email=%s", baseURL, email)
	fmt.Printf("Making request to: %s\n", debugURL)

	resp, err := http.Get(requestURL)
	if err != nil {
		return false, fmt.Errorf("failed to make request to EmailListVerify: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("EmailListVerify API returned status %d", resp.StatusCode)
	}

	// Read the response body first to debug
	body := make([]byte, 0)
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			body = append(body, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	// Log the raw response for debugging
	fmt.Printf("EmailListVerify raw response: %s\n", string(body))

	// Check if response looks like JSON
	if len(body) == 0 {
		return false, fmt.Errorf("EmailListVerify returned empty response")
	}

	// Check if response starts with '{' (JSON) or is plain text
	if body[0] != '{' {
		responseText := string(body)
		// Handle common plain text responses
		if responseText == "ok" {
			return true, nil
		}
		if responseText == "invalid" || responseText == "error" {
			return false, nil
		}
		return false, fmt.Errorf("EmailListVerify returned non-JSON response: %s", responseText)
	}

	var result emailListVerifyResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Errorf("failed to decode EmailListVerify JSON response: %w. Raw response: %s", err, string(body))
	}

	// Consider email valid if:
	// - Status is "ok"
	// - Result is "deliverable" or "risky" (risky emails might still be valid)
	// - Has MX record
	// - Not disposable (optional - you might want to allow disposable emails for testing)
	isValid := result.Status == "ok" &&
		(result.Result == "deliverable" || result.Result == "risky") &&
		result.MXRecord &&
		!result.Disposable

	return isValid, nil
}
