package AI

// AIRequest defines the structure for the input to the AI content generation.
type AIRequest struct {
	Keywords        string `json:"keywords"`         // User-provided keywords or topics
	ExistingContent string `json:"existing_content"` // Optional: Existing blog content for suggestions/enhancements
}

// AIResponse defines the structure for the output from the AI content generation.
type AIResponse struct {
	Suggestion string `json:"suggestion"`      // AI-generated content suggestion or enhancement
	Error      string `json:"error,omitempty"` // Optional: Error message if something went wrong
}

// Internal structure for the AI API request payload (e.g., for Google Gemini)
type GeminiRequest struct {
	Contents []struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"contents"`
	GenerationConfig struct {
		ResponseMimeType string `json:"responseMimeType"`
	} `json:"generationConfig"`
}

// Internal structure for the AI API response payload (e.g., from Google Gemini)
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}
