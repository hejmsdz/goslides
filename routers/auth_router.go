package routers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hejmsdz/goslides/di"
	"github.com/hejmsdz/goslides/dtos"
	"github.com/hejmsdz/goslides/services"
)

func RegisterAuthRoutes(r gin.IRouter, dic *di.Container) {
	h := NewAuthHandler(dic)

	r.POST("/auth/google", h.PostGoogleAuth)
}

type AuthHandler struct {
	Users *services.UsersService
	Auth  *services.AuthService
}

func NewAuthHandler(dic *di.Container) *AuthHandler {
	return &AuthHandler{
		Users: dic.Users,
		Auth:  dic.Auth,
	}
}

func (h *AuthHandler) PostGoogleAuth(c *gin.Context) {
	var data dtos.GoogleAuthRequest

	if err := c.ShouldBind(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	email, err := h.Auth.GetEmailFromGoogleIDToken(c.Request.Context(), data.IDToken)
	if err != nil {
		log.Printf("Failed to read Google ID token: %+v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	user, err := h.Users.GetUserByEmail(email)
	if err != nil {
		log.Printf("User with email %s not found", email)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credentials"})
		return
	}

	signedToken, err := h.Auth.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to authenticate"})
		return
	}

	c.JSON(http.StatusOK, dtos.GoogleAuthResponse{Token: signedToken, Name: user.DisplayName})
}
