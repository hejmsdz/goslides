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

	r.GET("/auth/me", h.Auth.AuthMiddleware, h.GetAuthMe)
	r.POST("/auth/google", h.PostAuthGoogle)
	r.POST("/auth/refresh", h.PostAuthRefresh)
	r.DELETE("/auth/refresh", h.DeleteAuthRefresh)
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

func (h *AuthHandler) GetAuthMe(c *gin.Context) {
	user, err := h.Users.GetUser(c.MustGet("userUUID").(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, dtos.NewAuthMeResponse(user))
}

func (h *AuthHandler) PostAuthGoogle(c *gin.Context) {
	var data dtos.AuthGoogleRequest

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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	accessToken, err := h.Auth.GenerateAccessToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to authenticate"})
		return
	}

	refreshToken, err := h.Auth.GenerateRefreshToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to authenticate"})
		return
	}

	c.JSON(http.StatusOK, dtos.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Name:         user.DisplayName,
	})
}

func (h *AuthHandler) PostAuthRefresh(c *gin.Context) {
	var data dtos.AuthRefreshRequest

	if err := c.ShouldBind(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	refreshToken, err := h.Auth.ValidateRefreshToken(data.RefreshToken)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token rejected"})
		return
	}

	accessToken, err := h.Auth.GenerateAccessToken(&refreshToken.User)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to authenticate"})
		return
	}

	c.JSON(http.StatusOK, dtos.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.Token,
		Name:         refreshToken.User.DisplayName,
	})
}

func (h *AuthHandler) DeleteAuthRefresh(c *gin.Context) {
	var data dtos.AuthRefreshRequest

	if err := c.ShouldBind(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.Auth.DeleteRefreshToken(data.RefreshToken)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token rejected"})
		return
	}

	c.Status(http.StatusNoContent)
}
