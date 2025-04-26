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
	r.PATCH("/users/me", h.Auth.AuthMiddleware, h.PatchUserMe)
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

func (h *UsersHandler) PatchUserMe(c *gin.Context) {
	user := h.Auth.GetCurrentUser(c)
	if user == nil {
		common.ReturnAPIError(c, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	var input dtos.UserUpdateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		common.ReturnAPIError(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	if err := input.Validate(); err != nil {
		common.ReturnAPIError(c, http.StatusUnprocessableEntity, "invalid request", err)
		return
	}

	err := h.Users.UpdateUser(user, input)
	if err != nil {
		common.ReturnAPIError(c, http.StatusInternalServerError, "failed to update user", err)
		return
	}

	c.JSON(http.StatusOK, dtos.NewUserMeResponse(user))
}
