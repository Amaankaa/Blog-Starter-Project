package usecases

import (
	"errors"
	"time"
	"context"

	blogpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/blog"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BlogUsecase struct {
	blogRepo blogpkg.IBlogRepository
}

func NewBlogUsecase(blogRepo blogpkg.IBlogRepository) *BlogUsecase {
	return &BlogUsecase{
		blogRepo: blogRepo,
	}
}
func (bu *BlogUsecase) CreateBlog(ctx context.Context, blog *blogpkg.Blog) (*blogpkg.Blog, error) {
	if blog == nil {
		return nil, errors.New("blog cannot be nil")
	}

	if blog.Title == "" {
		return nil, errors.New("blog title is required")
	}

	if blog.Content == "" {
		return nil, errors.New("blog content is required")
	}

	// Get user ID from context
	userID := ctx.Value("user_id")
	if userID == nil {
		return nil, errors.New("user ID not found in context")
	}

	authorIDStr, ok := userID.(string)
	if !ok {
		return nil, errors.New("user ID is not a string")
	}

	blog.AuthorID = authorIDStr
	blog.ID = primitive.NewObjectID().Hex()
	now := time.Now().Format(time.RFC3339)
	blog.CreatedAt = now
	blog.UpdatedAt = now

	createdBlog, err := bu.blogRepo.CreateBlog(blog)
	if err != nil {
		return nil, err
	}
	return createdBlog, nil
}
