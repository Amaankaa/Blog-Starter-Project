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
	protected.POST("/logout", authMiddleware.AuthMiddleware(), controller.Logout)

	// Blog routes
	protected.POST("/blog/create", blogController.CreateBlog)
	protected.GET("/blogs", blogController.GetAllBlogs)
	protected.GET("/blog/:id", blogController.GetBlogByID)
	
	return r
}
