package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	blogpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/blog"
	"github.com/gin-gonic/gin"
)

type BlogController struct {
	blogUsecase blogpkg.IBlogUsecase
}

func NewBlogController(blogUsecase blogpkg.IBlogUsecase) *BlogController {
	return &BlogController{
		blogUsecase: blogUsecase,
	}
}

// CreateBlog handles the creation of a new blog post
func (bc *BlogController) CreateBlog(c *gin.Context) {
	var blog blogpkg.Blog
	if err := c.ShouldBindJSON(&blog); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Create context with timeout and pass Gin context values
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Transfer Gin context values to standard context
	if userID, exists := c.Get("user_id"); exists {
		ctx = context.WithValue(ctx, "user_id", userID)
	}
	createdBlog, err := bc.blogUsecase.CreateBlog(ctx, &blog)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdBlog)
}

// GetBlogByID handles fetching a blog by its ID
func (bc *BlogController) GetBlogByID(c *gin.Context) {
	id := c.Param("id")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	blog, err := bc.blogUsecase.GetBlogByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, blog)
}

// GetAllBlogs handles fetching all blogs with pagination
func (bc *BlogController) GetAllBlogs(c *gin.Context) {
	page := 1
	limit := 10
	if p := c.Query("page"); p != "" {
		if v, err := ParseInt64(p); err == nil && v > 0 {
			page = int(v)
		}
	}
	if ps := c.Query("limit"); ps != "" {
		if v, err := ParseInt64(ps); err == nil && v > 0 {
			limit = int(v)
		}
	}

	pagination := blogpkg.PaginationRequest{
		Page:  page,
		Limit: limit,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := bc.blogUsecase.GetAllBlogs(ctx, pagination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// UpdateBlog handles updating an existing blog post
func (bc *BlogController) UpdateBlog(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Blog ID is required"})
		return
	}

	var blog blogpkg.Blog
	if err := c.ShouldBindJSON(&blog); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()


	updatedBlog, err := bc.blogUsecase.UpdateBlog(ctx, id, &blog)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updatedBlog)
}

// DeleteBlog handles deleting a blog post
func (bc *BlogController) DeleteBlog(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Blog ID is required"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := bc.blogUsecase.DeleteBlog(ctx, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (bc *BlogController) SearchBlogs(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	// Parse pagination parameters
	page, limit := parsePaginationParams(c, 1, 10)
	pagination := blogpkg.PaginationRequest{
		Page:  page,
		Limit: limit,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := bc.blogUsecase.SearchBlogs(ctx, query, pagination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// Helper function
func parsePaginationParams(c *gin.Context, defaultPage, defaultLimit int) (int, int) {
	page := defaultPage
	limit := defaultLimit

	if p := c.Query("page"); p != "" {
		if v, err := ParseInt64(p); err == nil && v > 0 {
			page = int(v)
		}
	}

	if ps := c.Query("limit"); ps != "" {
		if v, err := ParseInt64(ps); err == nil && v > 0 {
			limit = int(v)
		}
	}

	return page, limit
}

// FilterByTags handles filtering blogs by tags
func (bc *BlogController) FilterByTags(c *gin.Context) {
	tags := c.QueryArray("tags")
	if len(tags) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one tag is required"})
		return
	}

	page, limit := parsePaginationParams(c, 1, 10)

	pagination := blogpkg.PaginationRequest{
		Page:  page,
		Limit: limit,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := bc.blogUsecase.FilterByTags(ctx, tags, pagination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func ParseInt64(s string) (int64, error) {
	var v int64
	_, err := fmt.Sscan(s, &v)
	return v, err
}

func (bc *BlogController) LikeBlog(c *gin.Context) {
	blogID := c.Param("id")
	if blogID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Blog ID is required"})
		return
	}
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := bc.blogUsecase.ToggleLike(ctx, blogID, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Toggled like successfully"})
}

func (bc *BlogController) AddComment(c *gin.Context) {
	blogID := c.Param("id")
	if blogID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Blog ID is required"})
		return
	}
	var req blogpkg.AddCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	comment := &blogpkg.Comment{
		UserID:  userID.(string),
		Content: req.Content,
	}
	createdComment, err := bc.blogUsecase.AddComment(ctx, comment, blogID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Comment added successfully", "comment": createdComment})
}
