package blogpkg

import "context"

// BlogRepository interface defines the methods
type IBlogRepository interface {
	CreateBlog(blog *Blog) (*Blog, error)
	GetBlogByID(id string) (*Blog, error)
	GetAllBlogs(ctx context.Context, pagination PaginationRequest) (PaginationResponse, error)
	UpdateBlog(id string, blog *Blog) (*Blog, error)
	DeleteBlog(id string) error
}
