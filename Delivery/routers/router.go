package routers

import (
	"github.com/Amaankaa/Blog-Starter-Project/Delivery/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRouter(controller *controllers.Controller) *gin.Engine {
	r := gin.Default()

	// Public routes
	r.POST("/register", controller.Register)
	r.POST("/login", controller.Login)

	//Use the AuthMiddleWare for protected routes
	return r
}