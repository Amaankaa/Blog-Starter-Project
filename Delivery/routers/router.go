package routers

import (
	"github.com/Amaankaa/Blog-Starter-Project/Delivery/controllers"
	infrastructure "github.com/Amaankaa/Blog-Starter-Project/Infrastructure"

	"github.com/gin-gonic/gin"
)

func SetupRouter(controller *controllers.Controller, blogController *controllers.BlogController, authMiddleware *infrastructure.AuthMiddleware, aiController *controllers.AIController, aiRateLimiter gin.HandlerFunc) *gin.Engine {
	r := gin.Default()

	// Public routes
	r.POST("/register", controller.Register)
	r.POST("/login", controller.Login)
	r.POST("/forgot-password", controller.ForgotPassword)
	r.POST("/verify-otp", controller.VerifyOTP)
	r.POST("/reset-password", controller.ResetPassword)

	aiGroup := r.Group("/ai")
	aiGroup.Use(aiRateLimiter)
	{
		aiGroup.POST("/suggest-content", aiController.SuggestContent)
	}
	// Protected routes
	protected := r.Group("")
	protected.Use(authMiddleware.AuthMiddleware())

	//User routes
	protected.POST("/logout", controller.Logout)

	// Admin routes for user promotion and demotion
	admin := protected.Group("")
	admin.Use(authMiddleware.AdminOnly())
	admin.PUT("/user/:id/promote", controller.PromoteUser)
	admin.PUT("/user/:id/demote", controller.DemoteUser)

	// Blog routes (Public)
	r.GET("/blogs", blogController.GetAllBlogs)
	r.GET("/blogs/:id", blogController.GetBlogByID)
	r.GET("/blogs/search", blogController.SearchBlogs)
	r.GET("/blogs/filter", blogController.FilterByTags)
	
	// Blog routes (Protected)
	protected.POST("/blogs/create", blogController.CreateBlog)
	protected.PUT("/blogs/:id", blogController.UpdateBlog)
	protected.DELETE("/blogs/:id", blogController.DeleteBlog)
	protected.PATCH("/blogs/:id/like", blogController.LikeBlog)
	protected.POST("/blogs/:id/comment", blogController.AddComment)

	// AI routes
	

	return r
}
