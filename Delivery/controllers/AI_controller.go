package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time" // Added for context timeout in main.go snippet

	"github.com/gin-gonic/gin" // Assuming Gin framework for routing

	aidomain "github.com/Amaankaa/Blog-Starter-Project/Domain/AI"
	
)

// AIController handles HTTP requests related to AI features.
type AIController struct {
	aiUseCase aidomain.IAIUseCase
}

// NewAIController creates a new instance of AIController.
func NewAIController(useCase aidomain.IAIUseCase) *AIController {
	return &AIController{
		aiUseCase: useCase,
	}
}


// @Router /ai/suggest-content [post]
func (ctrl *AIController) SuggestContent(c *gin.Context) {
	var req aidomain.AIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, aidomain.AIResponse{Error: fmt.Sprintf("Invalid request payload: %v", err.Error())})
		return
	}

	ctx, cancel := c.Request.Context(), func() {} // Initialize cancel to a no-op function
	if c.Request.Context() != nil {
		ctx, cancel = context.WithTimeout(c.Request.Context(), 60*time.Second) // Increased timeout for AI calls
	}
	defer cancel()

	// Call the AI use case to get suggestions
	resp, err := ctrl.aiUseCase.GenerateContentSuggestions(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, aidomain.AIResponse{Error: fmt.Sprintf("Failed to get AI suggestions: %v", err.Error())})
		return
	}

	// Return the AI-generated suggestion
	c.JSON(http.StatusOK, resp)
}

