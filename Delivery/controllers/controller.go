package controllers

import (
	"net/http"
	"github.com/Amaankaa/Blog-Starter-Project/Domain"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	userUsecase domain.UserUsecase
}

func NewController(userUsecase domain.UserUsecase) *Controller {
	return &Controller{
		userUsecase: userUsecase,
	}
}

// User Controllers
func (ctrl *Controller) Register(c *gin.Context) {
	var user domain.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	createdUser, err := ctrl.userUsecase.RegisterUser(user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, createdUser)
}