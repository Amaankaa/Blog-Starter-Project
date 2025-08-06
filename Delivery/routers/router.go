package routers

import (
	"github.com/Amaankaa/Blog-Starter-Project/Delivery/controllers"
	infrastructure "github.com/Amaankaa/Blog-Starter-Project/Infrastructure"

	"github.com/gin-gonic/gin"
)

func SetupRouter(controller *controllers.Controller, blogController *controllers.BlogController, authMiddleware *infrastructure.AuthMiddleware) *gin.Engine {
	r := gin.Default()

	// Public routes
	r.POST("/register", controller.Register)
	r.POST("/login", controller.Login)
	r.POST("/forgot-password", controller.ForgotPassword)
	r.POST("/verify-otp", controller.VerifyOTP)
	r.POST("/reset-password", controller.ResetPassword)

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

	// Blog routes
	protected.POST("/blog/create", blogController.CreateBlog)
	protected.GET("/blogs", blogController.GetAllBlogs)
	protected.GET("/blog/:id", blogController.GetBlogByID)
	protected.PUT("/blog/:id", blogController.UpdateBlog)
	protected.DELETE("/blog/:id", blogController.DeleteBlog)
	protected.GET("/blog/search", blogController.SearchBlogs)
	protected.GET("/blog/filter", blogController.FilterByTags)

	return r
}
