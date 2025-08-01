package blogpkg

import (
	"context"
)

// BlogUsecase defines the interface for blog use cases
type IBlogUsecase interface {
	CreateBlog(ctx context.Context, blog *Blog) (*Blog, error)
}
