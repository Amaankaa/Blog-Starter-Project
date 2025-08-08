package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	blogpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/blog"
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
	if blog == nil {
		return nil, errors.New("blog not found")
	}
	userID := ctx.Value("user_id")
	if userID == nil {
		return blog, nil
	}

	// Increment view count only if user ID is present
	if _, ok := userID.(string); !ok {
		return blog, nil
	}
	if userID == blog.AuthorID {
		return blog, nil
	}

	err = bu.blogRepo.UpdateViewCount(ctx, id)
	if err != nil {
		return nil, err
	}

	blog.Views++
	return blog, nil
}

// GetAllBlogs returns paginated blogs
func (bu *BlogUsecase) GetAllBlogs(ctx context.Context, pagination blogpkg.PaginationRequest) (blogpkg.PaginationResponse, error) {
	// Set default values if not provided
	pagination = normalizePagination(pagination)

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
	existingBlog, err := bu.blogRepo.FindBlogByID(id)
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

	blog, err := bu.blogRepo.FindBlogByID(id)
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

func (bu *BlogUsecase) SearchBlogs(ctx context.Context, query string, pagination blogpkg.PaginationRequest) (blogpkg.PaginationResponse, error) {
	if query == "" {
		return blogpkg.PaginationResponse{}, errors.New("search query cannot be empty")
	}

	pagination = normalizePagination(pagination)

	result, err := bu.blogRepo.SearchBlogs(ctx, query, pagination)
	if err != nil {
		return blogpkg.PaginationResponse{}, err
	}
	return result, nil
}

func (bu *BlogUsecase) FilterByTags(ctx context.Context, tags []string, pagination blogpkg.PaginationRequest) (blogpkg.PaginationResponse, error) {
	if len(tags) == 0 {
		return blogpkg.PaginationResponse{}, errors.New("tags cannot be empty")
	}

	pagination = normalizePagination(pagination)

	// Call the repository to filter blogs by tags
	if len(tags) > 5 {
		return blogpkg.PaginationResponse{}, errors.New("too many tags, maximum is 5")
	}

	for _, tag := range tags {
		if tag == "" {
			return blogpkg.PaginationResponse{}, errors.New("tag cannot be empty")
		}
	}

	result, err := bu.blogRepo.FilterByTags(ctx, tags, pagination)
	if err != nil {
		return blogpkg.PaginationResponse{}, err
	}
	return result, nil
}

// helper function
func normalizePagination(p blogpkg.PaginationRequest) blogpkg.PaginationRequest {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Limit <= 0 {
		p.Limit = 10
	}
	if p.Limit > 100 {
		p.Limit = 100
	}
	return p
}

func (bu *BlogUsecase) ToggleLike(ctx context.Context, blogID string, userID string) error {
	blog, err := bu.blogRepo.FindBlogByID(blogID)
	if err != nil {
		return err
	}

	if blog == nil {
		return errors.New("blog not found")
	}

	// Check if aleardy liked
	for _, id := range blog.Likes {
		if id == userID {
			return bu.blogRepo.RemoveLike(ctx, blogID, userID)
		}
	}

	return bu.blogRepo.AddLike(ctx, blogID, userID)
}

func (bu *BlogUsecase) AddComment(ctx context.Context, comment *blogpkg.Comment, blogID string) (*blogpkg.Comment, error) {
	exists, err := bu.blogRepo.FindBlogByID(blogID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if blog exists: %w", err)
	}
	if exists == nil {
		return nil, errors.New("blog not found")
	}

	newComment := &blogpkg.Comment{
		BlogID:    comment.BlogID,
		UserID:    comment.UserID,
		Content:   comment.Content,
		CreatedAt: time.Now(),
	}

	return bu.blogRepo.AddComment(ctx, newComment)
}
