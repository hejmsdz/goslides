package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hejmsdz/goslides/common"
	"github.com/hejmsdz/goslides/di"
	"github.com/hejmsdz/goslides/dtos"
	"github.com/hejmsdz/goslides/services"
)

func RegisterUsersRoutes(r gin.IRouter, dic *di.Container) {
	h := NewUsersHandler(dic)

	r.GET("/users/me", h.Auth.AuthMiddleware, h.GetUserMe)
}

type UsersHandler struct {
	Users *services.UsersService
	Auth  *services.AuthService
}

func NewUsersHandler(dic *di.Container) *UsersHandler {
	return &UsersHandler{
		Users: dic.Users,
		Auth:  dic.Auth,
	}
}

func (h *UsersHandler) GetUserMe(c *gin.Context) {
	user := h.Auth.GetCurrentUser(c)
	if user == nil {
		common.ReturnAPIError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}
	c.JSON(http.StatusOK, dtos.NewUserMeResponse(user))
}
