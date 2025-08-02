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

// ParseInt64 safely parses a string to int64
func ParseInt64(s string) (int64, error) {
	var v int64
	_, err := fmt.Sscan(s, &v)
	return v, err
}
