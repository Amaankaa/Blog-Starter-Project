package usecases

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Amaankaa/Blog-Starter-Project/Domain/AI"
	aidomain "github.com/Amaankaa/Blog-Starter-Project/Domain/AI"
)

type aiUseCase struct {
	aiAPIKey   string
	aiAPIURL   string
	httpClient *http.Client
}

// NewAIUseCase takes the AI API key and URL as parameters, which should be loaded from environment variables.
func NewAIUseCase(aiAPIKey, aiAPIURL string) aidomain.IAIUseCase {
	return &aiUseCase{
		aiAPIKey:   aiAPIKey,
		aiAPIURL:   aiAPIURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
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

	geminiReq := AI.GeminiRequest{
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
			ResponseMimeType: "text/plain",
		},
	}

	jsonPayload, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal AI request payload: %w", err)
	}

	fullAPIURL := fmt.Sprintf("%s?key=%s", uc.aiAPIURL, uc.aiAPIKey)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", fullAPIURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request to AI API: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	var geminiResp AI.GeminiResponse
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		resp, err := uc.httpClient.Do(httpReq)
		if err != nil {
			if i < maxRetries-1 {
				log.Printf("AI API request failed, retrying in %d seconds... (attempt %d/%d)\n", 1<<i, i+1, maxRetries)
				time.Sleep(time.Duration(1<<i) * time.Second)
				continue
			}
			return nil, fmt.Errorf("failed to make request to AI API after %d retries: %w", maxRetries, err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
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
