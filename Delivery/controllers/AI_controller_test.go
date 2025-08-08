package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Amaankaa/Blog-Starter-Project/Delivery/controllers"
	aidomain "github.com/Amaankaa/Blog-Starter-Project/Domain/AI"
	"github.com/Amaankaa/Blog-Starter-Project/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type AIControllerTestSuite struct {
	suite.Suite
	router         *gin.Engine
	mockAIUseCase  *mocks.IAIUseCase
	controller     *controllers.AIController
}

func TestAIControllerTestSuite(t *testing.T) {
	suite.Run(t, new(AIControllerTestSuite))
}

func (s *AIControllerTestSuite) SetupTest() {
	s.mockAIUseCase = new(mocks.IAIUseCase)
	s.controller = controllers.NewAIController(s.mockAIUseCase)
	
	gin.SetMode(gin.TestMode)
	s.router = gin.Default()
	s.router.POST("/ai/suggest-content", s.controller.SuggestContent)
}

func (s *AIControllerTestSuite) TestSuggestContent_Success() {
	reqPayload := aidomain.AIRequest{
		Keywords: "golang testing",
		ExistingContent: "Some existing content about Go.",
	}
	expectedResponse := aidomain.AIResponse{
		Suggestion: "New content suggestion 1\nNew content suggestion 2",
	}
	
	s.mockAIUseCase.On("GenerateContentSuggestions", mock.Anything, &reqPayload).Return(&expectedResponse, nil)
	
	jsonPayload, _ := json.Marshal(reqPayload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/ai/suggest-content", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	
	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusOK, w.Code)
	var response aidomain.AIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(s.T(), expectedResponse, response)
	s.mockAIUseCase.AssertExpectations(s.T())
}

func (s *AIControllerTestSuite) TestSuggestContent_InvalidJSONPayload() {
	
	invalidPayload := `{"keywords": "golang testing", "existing_content": 123}` // existing_content should be a string
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/ai/suggest-content", bytes.NewBufferString(invalidPayload))
	req.Header.Set("Content-Type", "application/json")
	
	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
	var response aidomain.AIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(s.T(), response.Error, "Invalid request payload")
	s.mockAIUseCase.AssertNotCalled(s.T(), "GenerateContentSuggestions")
}

func (s *AIControllerTestSuite) TestSuggestContent_TooManyKeywords() {
	
	reqPayload := aidomain.AIRequest{
		Keywords: "golang testing microservices rest api databases", 
	}
	
	jsonPayload, _ := json.Marshal(reqPayload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/ai/suggest-content", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	
	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
	var response aidomain.AIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(s.T(), response.Error, "Keywords must be 5 words or less.")
	s.mockAIUseCase.AssertNotCalled(s.T(), "GenerateContentSuggestions")
}

func (s *AIControllerTestSuite) TestSuggestContent_KeywordsTooLong() {
	
	longKeyword := strings.Repeat("a", 101)
	reqPayload := aidomain.AIRequest{
		Keywords: longKeyword,
	}
	
	jsonPayload, _ := json.Marshal(reqPayload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/ai/suggest-content", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")

	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
	var response aidomain.AIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(s.T(), response.Error, "Keywords must be under 100 characters.")
	s.mockAIUseCase.AssertNotCalled(s.T(), "GenerateContentSuggestions")
}

func (s *AIControllerTestSuite) TestSuggestContent_ExistingContentTooLong() {
	longContent := strings.Repeat("a", 5001) 
	reqPayload := aidomain.AIRequest{
		Keywords: "go",
		ExistingContent: longContent,
	}
	
	jsonPayload, _ := json.Marshal(reqPayload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/ai/suggest-content", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	
	s.router.ServeHTTP(w, req)
	
	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
	var response aidomain.AIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(s.T(), response.Error, "Existing content must be under 5000 characters.")
	s.mockAIUseCase.AssertNotCalled(s.T(), "GenerateContentSuggestions")
}

func (s *AIControllerTestSuite) TestSuggestContent_UseCaseError() {

	reqPayload := aidomain.AIRequest{
		Keywords: "golang testing",
	}
	
	s.mockAIUseCase.On("GenerateContentSuggestions", mock.Anything, &reqPayload).Return(&aidomain.AIResponse{}, errors.New("something went wrong"))
	
	jsonPayload, _ := json.Marshal(reqPayload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/ai/suggest-content", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	
	s.router.ServeHTTP(w, req)
	
	assert.Equal(s.T(), http.StatusInternalServerError, w.Code)
	var response aidomain.AIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(s.T(), response.Error, "Failed to get AI suggestions: something went wrong")
	s.mockAIUseCase.AssertExpectations(s.T())
}

func (s *AIControllerTestSuite) TestSuggestContent_ContextTimeout() {

	reqPayload := aidomain.AIRequest{
		Keywords: "golang testing",
	}
	
	s.mockAIUseCase.On("GenerateContentSuggestions", mock.Anything, &reqPayload).
		Run(func(args mock.Arguments) {
			ctx := args.Get(0).(context.Context)
			select {
			case <-time.After(2 * time.Second):
				// This case should not be hit if the timeout works
			case <-ctx.Done():
				// This is the expected path
			}
		}).
		Return(&aidomain.AIResponse{}, context.DeadlineExceeded)

	jsonPayload, _ := json.Marshal(reqPayload)
	
	w := httptest.NewRecorder()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	req, _ := http.NewRequest(http.MethodPost, "/ai/suggest-content", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)

	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusInternalServerError, w.Code)
	var response aidomain.AIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(s.T(), response.Error, "context deadline exceeded")
	s.mockAIUseCase.AssertExpectations(s.T())
}
