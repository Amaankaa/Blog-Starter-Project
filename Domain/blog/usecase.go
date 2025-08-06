package blogpkg

import (
	"context"
)

// BlogUsecase defines the interface for blog use cases
type IBlogUsecase interface {
	CreateBlog(ctx context.Context, blog *Blog) (*Blog, error)
	GetBlogByID(ctx context.Context, id string) (*Blog, error)
	GetAllBlogs(ctx context.Context, pagination PaginationRequest) (PaginationResponse, error)
	UpdateBlog(ctx context.Context, id string, blog *Blog) (*Blog, error)
	DeleteBlog(ctx context.Context, id string) error
	SearchBlogs(ctx context.Context, query string, pagination PaginationRequest) (PaginationResponse, error)
	FilterByTags(ctx context.Context, tags []string, pagination PaginationRequest) (PaginationResponse, error)
}
