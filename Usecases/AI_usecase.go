package usecases

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	// Import the AI domain package where IAIUseCase interface and entities are defined
	// Replace "github.com/Amaankaa/Blog-Starter-Project/Domain/AI" with your actual project path
	"github.com/Amaankaa/Blog-Starter-Project/Domain/AI"
	aidomain "github.com/Amaankaa/Blog-Starter-Project/Domain/AI"
)

// aiUseCase implements the aidomain.IAIUseCase interface.
// It holds the necessary dependencies to interact with the external AI service.
type aiUseCase struct {
	aiAPIKey   string       // API key for the external AI service (e.g., Gemini API key)
	aiAPIURL   string       // Base URL for the external AI service endpoint
	httpClient *http.Client // HTTP client for making requests to the AI API
}

// NewAIUseCase creates a new instance of aiUseCase.
// It takes the AI API key and URL as parameters, which should be loaded from environment variables
// or a configuration management system in your main application.
func NewAIUseCase(aiAPIKey, aiAPIURL string) aidomain.IAIUseCase {
	return &aiUseCase{
		aiAPIKey:   aiAPIKey,
		aiAPIURL:   aiAPIURL,
		httpClient: &http.Client{Timeout: 30 * time.Second}, // Set a reasonable timeout for AI API calls
	}
}


func (uc *aiUseCase) GenerateContentSuggestions(ctx context.Context, req *aidomain.AIRequest) (*aidomain.AIResponse, error) {
	// Construct the prompt for the AI model based on user input
	prompt := "Generate blog content suggestions or enhancements."
	if req.Keywords != "" {
		prompt += fmt.Sprintf(" Focus on these keywords/topics: %s.", req.Keywords)
	}
	if req.ExistingContent != "" {
		prompt += fmt.Sprintf(" Improve or suggest enhancements for this existing content: %s", req.ExistingContent)
	}


	geminiReq := AI.GeminiRequest{ // Using the internal struct from aidomain
		Contents: []struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		}{
			{
				Parts: []struct {
					Text string `json:"text"`
				}{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: struct {
			ResponseMimeType string `json:"responseMimeType"`
		}{
			ResponseMimeType: "text/plain", // Requesting plain text output from the AI
		},
	}

	jsonPayload, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal AI request payload: %w", err)
	}

	fullAPIURL := fmt.Sprintf("%s?key=%s", uc.aiAPIURL, uc.aiAPIKey)

	// Create the HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", fullAPIURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request to AI API: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	var geminiResp AI.GeminiResponse // Using the internal struct from aidomain
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		resp, err := uc.httpClient.Do(httpReq)
		if err != nil {
			if i < maxRetries-1 {
				fmt.Printf("AI API request failed, retrying in %d seconds... (attempt %d/%d)\n", 1<<i, i+1, maxRetries)
				time.Sleep(time.Duration(1<<i) * time.Second) 
				continue
			}
			return nil, fmt.Errorf("failed to make request to AI API after %d retries: %w", maxRetries, err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read AI API response body: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("AI API returned non-OK status: %d - %s", resp.StatusCode, string(body))
		}

		err = json.Unmarshal(body, &geminiResp)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal AI API response: %w", err)
		}
		break 
	}

	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		return &aidomain.AIResponse{
			Suggestion: geminiResp.Candidates[0].Content.Parts[0].Text,
		}, nil
	}

	return &aidomain.AIResponse{
		Error: "No valid suggestion found in AI response. Please try again with different input.",
	}, nil
}
