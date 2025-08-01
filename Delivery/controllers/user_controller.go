package controllers

import (
	"context"
	"net/http"
	"time"

	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	userUsecase userpkg.IUserUsecase
}

func NewController(userUsecase userpkg.IUserUsecase) *Controller {
	return &Controller{
		userUsecase: userUsecase,
	}
}

// User Controllers
func (ctrl *Controller) Register(c *gin.Context) {
	var user userpkg.User

	// 1. Parse JSON input
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. Create context with timeout (e.g. 20s)
	//We needed a longer timeout due to MailboxLayer to verify the emails validity
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()

	// 3. Call the usecase
	createdUser, err := ctrl.userUsecase.RegisterUser(ctx, user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 4. Success response
	c.JSON(http.StatusCreated, createdUser)
}