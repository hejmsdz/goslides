package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hejmsdz/goslides/common"
	"github.com/hejmsdz/goslides/di"
	"github.com/hejmsdz/goslides/dtos"
	"github.com/hejmsdz/goslides/services"
	"github.com/pkg/errors"
)

func RegisterAuthRoutes(r gin.IRouter, dic *di.Container) {
	h := NewAuthHandler(dic)

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

func (h *AuthHandler) PostAuthGoogle(c *gin.Context) {
	var data dtos.AuthGoogleRequest

	if err := c.ShouldBind(&data); err != nil {
		common.ReturnBadRequestError(c, err)
		return
	}

	email, err := h.Auth.GetEmailFromGoogleIDToken(c.Request.Context(), data.IDToken)
	if err != nil {
		common.ReturnAPIError(c, http.StatusUnauthorized, "invalid credentials", errors.Wrap(err, "failed to get email from google id token"))
		return
	}

	user, err := h.Users.GetUserByEmail(email)
	if err != nil {
		common.ReturnAPIError(c, http.StatusUnauthorized, "invalid credentials", errors.Wrapf(err, "user with email %s not found", email))
		return
	}

	accessToken, err := h.Auth.GenerateAccessToken(user)
	if err != nil {
		common.ReturnAPIError(c, http.StatusInternalServerError, "failed to authenticate", errors.Wrap(err, "failed to generate access token"))
		return
	}

	refreshToken, err := h.Auth.GenerateRefreshToken(user)
	if err != nil {
		common.ReturnAPIError(c, http.StatusInternalServerError, "failed to authenticate", errors.Wrap(err, "failed to generate refresh token"))
		return
	}

	c.JSON(http.StatusOK, dtos.NewAuthResponse(accessToken, refreshToken, user))
}

func (h *AuthHandler) PostAuthRefresh(c *gin.Context) {
	var data dtos.AuthRefreshRequest

	if err := c.ShouldBind(&data); err != nil {
		common.ReturnBadRequestError(c, err)
		return
	}

	refreshToken, err := h.Auth.ValidateRefreshToken(data.RefreshToken)
	if err != nil {
		common.ReturnAPIError(c, http.StatusUnauthorized, "refresh token rejected", errors.Wrap(err, "failed to validate refresh token"))
		return
	}

	accessToken, err := h.Auth.GenerateAccessToken(&refreshToken.User)
	if err != nil {
		common.ReturnAPIError(c, http.StatusInternalServerError, "failed to authenticate", errors.Wrap(err, "failed to generate access token"))
		return
	}

	c.JSON(http.StatusOK, dtos.NewAuthResponse(accessToken, refreshToken.Token, &refreshToken.User))
}

func (h *AuthHandler) DeleteAuthRefresh(c *gin.Context) {
	var data dtos.AuthRefreshRequest

	if err := c.ShouldBind(&data); err != nil {
		common.ReturnBadRequestError(c, err)
		return
	}

	err := h.Auth.DeleteRefreshToken(data.RefreshToken)
	if err != nil {
		common.ReturnAPIError(c, http.StatusUnauthorized, "refresh token rejected", errors.Wrap(err, "failed to delete refresh token"))
		return
	}

	c.Status(http.StatusNoContent)
}
