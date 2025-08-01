package controllers

import (
	"context"
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
