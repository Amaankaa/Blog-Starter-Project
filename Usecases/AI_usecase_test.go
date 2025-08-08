package usecases_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

    aidomain "github.com/Amaankaa/Blog-Starter-Project/Domain/AI"
	usecases "github.com/Amaankaa/Blog-Starter-Project/Usecases"
	"github.com/stretchr/testify/suite"
)

// Define a custom mock for the http.RoundTripper interface to simulate API responses.
type mockRoundTripper struct {
	roundTripFunc func(*http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

type AIUsecaseTestSuite struct {
	suite.Suite
	ctx        context.Context
	mockClient *http.Client
	usecase    *usecases.AIUseCase
}

func TestAIUsecaseTestSuite(t *testing.T) {
	suite.Run(t, new(AIUsecaseTestSuite))
}

func (s *AIUsecaseTestSuite) SetupTest() {
	s.ctx = context.Background()

	aiAPIKey := "test-api-key"
	aiAPIURL := "https://api.test.com"

	s.usecase = usecases.NewAIUseCase(aiAPIKey, aiAPIURL).(*usecases.AIUseCase)
}

func (s *AIUsecaseTestSuite) TestGenerateContentSuggestions_Success() {
    s.mockClient = &http.Client{
        Timeout: 30 * time.Second,
        Transport: &mockRoundTripper{
            roundTripFunc: func(req *http.Request) (*http.Response, error) {
                responseBody := `{"candidates":[{"content":{"parts":[{"text":"This is a suggested blog post."}]}}]}`
                return &http.Response{
                    StatusCode: http.StatusOK,
                    Body:       io.NopCloser(strings.NewReader(responseBody)),
                }, nil
            },
        },
    }
    s.usecase.HTTPClient = s.mockClient

    req := &aidomain.AIRequest{
        Keywords: "Suggest a blog post about Go testing",
    }
    result, err := s.usecase.GenerateContentSuggestions(s.ctx, req)
    s.NoError(err)
    s.Equal("This is a suggested blog post.", result.Suggestion)
}
func (s *AIUsecaseTestSuite) TestGenerateContentSuggestions_EmptyPrompt() {
    s.mockClient = &http.Client{
        Timeout: 30 * time.Second,
        Transport: &mockRoundTripper{
            roundTripFunc: func(req *http.Request) (*http.Response, error) {
                return &http.Response{
                    StatusCode: http.StatusOK,
                    Body:       io.NopCloser(strings.NewReader(`{}`)),
                }, nil
            },
        },
    }
    s.usecase.HTTPClient = s.mockClient

    req := &aidomain.AIRequest{
        Keywords: "",
    }
    _, err := s.usecase.GenerateContentSuggestions(s.ctx, req)
    s.Error(err)
}
func (s *AIUsecaseTestSuite) TestGenerateContentSuggestions_APIError() {
    s.mockClient = &http.Client{
        Timeout: 30 * time.Second,
        Transport: &mockRoundTripper{
            roundTripFunc: func(req *http.Request) (*http.Response, error) {
                return nil, errors.New("network failure")
            },
        },
    }
    s.usecase.HTTPClient = s.mockClient

    req := &aidomain.AIRequest{
        Keywords: "Suggest a blog post",
    }
    _, err := s.usecase.GenerateContentSuggestions(s.ctx, req)
    s.Error(err)
    s.Contains(err.Error(), "network failure")
}

func (s *AIUsecaseTestSuite) TestGenerateContentSuggestions_APIStatusError() {
    s.mockClient = &http.Client{
        Timeout: 30 * time.Second,
        Transport: &mockRoundTripper{
            roundTripFunc: func(req *http.Request) (*http.Response, error) {
                return &http.Response{
                    StatusCode: http.StatusInternalServerError,
                    Body:       io.NopCloser(strings.NewReader(`{"error": {"message": "internal server error"}}`)),
                }, nil
            },
        },
    }
    s.usecase.HTTPClient = s.mockClient

    req := &aidomain.AIRequest{
        Keywords: "Suggest a blog post",
    }
    _, err := s.usecase.GenerateContentSuggestions(s.ctx, req)
    s.Error(err)
    s.Contains(err.Error(), "internal server error")
}

func (s *AIUsecaseTestSuite) TestGenerateContentSuggestions_InvalidAPIResponse() {
    s.mockClient = &http.Client{
        Timeout: 30 * time.Second,
        Transport: &mockRoundTripper{
            roundTripFunc: func(req *http.Request) (*http.Response, error) {
                responseBody := `{"invalid_key":"not a valid response"}`
                return &http.Response{
                    StatusCode: http.StatusOK,
                    Body:       io.NopCloser(strings.NewReader(responseBody)),
                }, nil
            },
        },
    }
    s.usecase.HTTPClient = s.mockClient

    req := &aidomain.AIRequest{
        Keywords: "Suggest a blog post",
    }
    _, err := s.usecase.GenerateContentSuggestions(s.ctx, req)
    s.Error(err)
    s.Contains(err.Error(), "failed to parse AI response")
}