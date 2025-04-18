package routers

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/hejmsdz/goslides/common"
	"github.com/hejmsdz/goslides/core"
	"github.com/hejmsdz/goslides/di"
	"github.com/hejmsdz/goslides/dtos"
	"github.com/hejmsdz/goslides/services"
)

func RegisterDeckRoutes(r gin.IRouter, dic *di.Container) {
	h := NewDeckHandler(dic)

	r.POST("/deck", h.Auth.OptionalAuthMiddleware, h.PostDeck)
}

type DeckHandler struct {
	Songs   *services.SongsService
	Liturgy *services.LiturgyService
	Auth    *services.AuthService
}

func NewDeckHandler(dic *di.Container) *DeckHandler {
	return &DeckHandler{
		Songs:   dic.Songs,
		Liturgy: dic.Liturgy,
		Auth:    dic.Auth,
	}
}

func (h *DeckHandler) PostDeck(c *gin.Context) {
	var deck dtos.DeckRequest
	if err := c.ShouldBind(&deck); err != nil {
		common.ReturnBadRequestError(c, err)
		return
	}

	if err := deck.Validate(); err != nil {
		common.ReturnAPIError(c, http.StatusUnprocessableEntity, "validation failed", err)
		return
	}

	user := h.Auth.GetCurrentUser(c)

	textDeck, ok := services.BuildTextSlides(deck, h.Songs, h.Liturgy, user)
	if !ok {
		common.ReturnAPIError(c, http.StatusInternalServerError, "failed to get lyrics", nil)
		return
	}

	extension := ""
	var file io.Reader
	var contents []core.ContentSlide
	var err error

	switch deck.Format {
	case "txt":
		extension = ".txt"
		text := core.Tugalize(textDeck)
		file = strings.NewReader(text)

	default:
		extension = ".pdf"
		file, contents, err = core.BuildPDF(textDeck, services.GetPageConfig(deck))
		if err != nil {
			common.ReturnError(c, err)
			return
		}
	}

	fileName := uuid.New().String() + extension
	common.SaveTemporaryFile(file, fileName)

	if !deck.Contents {
		contents = nil
	}

	resp := dtos.NewDeckResponse(common.GetPublicURL(fileName), contents)
	c.JSON(http.StatusOK, resp)
}
