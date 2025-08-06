package blogpkg

import "context"

// BlogRepository interface defines the methods
type IBlogRepository interface {
	CreateBlog(blog *Blog) (*Blog, error)
	GetBlogByID(id string) (*Blog, error)
	GetAllBlogs(ctx context.Context, pagination PaginationRequest) (PaginationResponse, error)
	UpdateBlog(id string, blog *Blog) (*Blog, error)
	DeleteBlog(id string) error
	SearchBlogs(ctx context.Context, query string, pagination PaginationRequest) (PaginationResponse, error)
	FilterByTags(ctx context.Context, tags []string, pagination PaginationRequest) (PaginationResponse, error)
	AddLike(ctx context.Context, blogID string, userID string) error
	RemoveLike(ctx context.Context, blogID string, userID string) error
}
