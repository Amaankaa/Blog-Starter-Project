package controllers

import (
	"os"
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

func (ctrl *Controller) Login(c *gin.Context) {
	var input struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	user, accessToken, refreshToken, err := ctrl.userUsecase.LoginUser(ctx, input.Login, input.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":          user,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (ctrl *Controller) RefreshToken(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.ShouldBindJSON(&body); err != nil || body.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh token required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	newTokens, err := ctrl.userUsecase.RefreshToken(ctx, body.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, newTokens)
}

func (ctrl *Controller) ForgotPassword(c *gin.Context) {
    var req struct{ Email string }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()

    if err := ctrl.userUsecase.SendResetOTP(ctx, req.Email); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "OTP sent"})
}

func (ctrl *Controller) VerifyOTP(c *gin.Context) {
    var req struct {
        Email string `json:"email"`
        OTP   string `json:"otp"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()

    if err := ctrl.userUsecase.VerifyOTP(ctx, req.Email, req.OTP); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "OTP verified, you can reset your password."})
}

func (ctrl *Controller) ResetPassword(c *gin.Context) {
    var req struct {
        Email       string `json:"email"`
        NewPassword string `json:"new_password"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()

    if err := ctrl.userUsecase.ResetPassword(ctx, req.Email, req.NewPassword); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Password reset successful"})
}

func (ctrl *Controller) Logout(c *gin.Context) {
	userID := c.GetString("userID") // assuming middleware sets this
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	err := ctrl.userUsecase.Logout(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Logout failed"})
		return
	}
	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	if cookieDomain == "" {
		cookieDomain = "localhost"
	}

	// Optionally clear tokens from client
	c.SetCookie("access_token", "", -1, "/", cookieDomain, false, true)
	c.SetCookie("refresh_token", "", -1, "/", cookieDomain, false, true)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
