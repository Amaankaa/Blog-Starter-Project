package AI

import "context"

type IAIUseCase interface {
	GenerateContentSuggestions(ctx context.Context, req *AIRequest) (*AIResponse, error)
}
