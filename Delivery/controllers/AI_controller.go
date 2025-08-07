package controllers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	aidomain "github.com/Amaankaa/Blog-Starter-Project/Domain/AI"
)

type AIController struct {
	aiUseCase aidomain.IAIUseCase
}

func NewAIController(useCase aidomain.IAIUseCase) *AIController {
	return &AIController{
		aiUseCase: useCase,
	}
}

func (ctrl *AIController) SuggestContent(c *gin.Context) {
	var req aidomain.AIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, aidomain.AIResponse{Error: fmt.Sprintf("Invalid request payload: %v", err.Error())})
		return
	}
	keywords := strings.Split(req.Keywords, " ")
	if len(keywords) > 5 {
		c.JSON(http.StatusBadRequest, aidomain.AIResponse{Error: "Keywords must be 5 words or less."})
		return
	}

	// Character limits for keywords and existing content
	const maxKeywordsLength = 100
	const maxContentLength = 5000

	if len(req.Keywords) > maxKeywordsLength {
		c.JSON(http.StatusBadRequest, aidomain.AIResponse{Error: fmt.Sprintf("Keywords must be under %d characters.", maxKeywordsLength)})
		return
	}

	if len(req.ExistingContent) > maxContentLength {
		c.JSON(http.StatusBadRequest, aidomain.AIResponse{Error: fmt.Sprintf("Existing content must be under %d characters.", maxContentLength)})
		return
	}
	ctx, cancel := c.Request.Context(), func() {}
	if c.Request.Context() != nil {
		ctx, cancel = context.WithTimeout(c.Request.Context(), 60*time.Second)
	}
	defer cancel()

	resp, err := ctrl.aiUseCase.GenerateContentSuggestions(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, aidomain.AIResponse{Error: fmt.Sprintf("Failed to get AI suggestions: %v", err.Error())})
		return
	}

	c.JSON(http.StatusOK, resp)
}
