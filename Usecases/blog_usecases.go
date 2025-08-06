package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

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
	blog.CreatedAt = time.Now()
	blog.UpdatedAt = time.Now()

	createdBlog, err := bu.blogRepo.CreateBlog(blog)
	if err != nil {
		return nil, err
	}
	return createdBlog, nil
}

// GetBlogByID returns a blog by its ID
func (bu *BlogUsecase) GetBlogByID(ctx context.Context, id string) (*blogpkg.Blog, error) {
	if id == "" {
		return nil, errors.New("blog ID is required")
	}
	blog, err := bu.blogRepo.GetBlogByID(id)
	if err != nil {
		return nil, err
	}
	return blog, nil
}

// GetAllBlogs returns paginated blogs
func (bu *BlogUsecase) GetAllBlogs(ctx context.Context, pagination blogpkg.PaginationRequest) (blogpkg.PaginationResponse, error) {
   // Set default values if not provided
   if pagination.Page <= 0 {
       pagination.Page = 1
   }
   if pagination.Limit <= 0 {
       pagination.Limit = 10
   }
   if pagination.Limit > 100 {
       pagination.Limit = 100 // Limit max page size
   }

	result, err := bu.blogRepo.GetAllBlogs(ctx, pagination)
	if err != nil {
		return blogpkg.PaginationResponse{}, err
	}
	return result, nil
}

// UpdateBlog updates an existing blog
func (bu *BlogUsecase) UpdateBlog(ctx context.Context, id string, blog *blogpkg.Blog) (*blogpkg.Blog, error) {
	if id == "" {
		return nil, errors.New("blog ID is required")
	}

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
	if !ok || authorIDStr == "" {
		return nil, errors.New("invalid user ID in context")
	}

	// Fetch existing blog to validate ownership
	existingBlog, err := bu.blogRepo.GetBlogByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch existing blog: %w", err)
	}
	if existingBlog == nil {
		return nil, errors.New("blog not found")
	}
	if existingBlog.AuthorID != authorIDStr {
		return nil, errors.New("unauthorized to update this blog")
	}

	blog.AuthorID = authorIDStr
	blog.ID = existingBlog.ID
	blog.CreatedAt = existingBlog.CreatedAt
	blog.UpdatedAt = time.Now()

	updatedBlog, err := bu.blogRepo.UpdateBlog(id, blog)
	if err != nil {
		return nil, err
		}
	return updatedBlog, nil
}

// DeleteBlog deletes a blog by its ID
func (bu *BlogUsecase) DeleteBlog(ctx context.Context, id string) error {
	userID := ctx.Value("user_id")
	if userID == nil {
		return errors.New("user ID not found in context")
	}
	authorID, ok := userID.(string)
	if !ok || authorID == "" {
		return errors.New("invalid user ID")
	}

	blog, err := bu.blogRepo.GetBlogByID(id)
	if err != nil {
		return fmt.Errorf("failed to fetch blog: %w", err)
	}
	if blog == nil {
		return errors.New("blog not found")
	}
	if blog.AuthorID != authorID {
		return errors.New("unauthorized to delete this blog")
	}

	err = bu.blogRepo.DeleteBlog(id)
	if err != nil {
		return err
	}
	return nil
}
