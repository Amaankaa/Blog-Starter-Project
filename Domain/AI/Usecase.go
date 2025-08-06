package AI

import "context"

// IAIUseCase defines the interface for AI-related business logic within the Domain layer.
// This interface is implemented by a concrete use case in the 'usecases' directory.
type IAIUseCase interface {
	GenerateContentSuggestions(ctx context.Context, req *AIRequest) (*AIResponse, error)
}
