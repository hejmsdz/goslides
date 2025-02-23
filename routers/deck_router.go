package routers

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/hejmsdz/goslides/common"
	"github.com/hejmsdz/goslides/core"
	"github.com/hejmsdz/goslides/di"
	"github.com/hejmsdz/goslides/dtos"
	"github.com/hejmsdz/goslides/services"
)

func RegisterDeckRoutes(r gin.IRouter, dic *di.Container) {
	h := NewDeckHandler(dic)

	r.POST("/deck", h.PostDeck)
}

type DeckHandler struct {
	Songs   services.SongsService
	Liturgy services.LiturgyService
}

func NewDeckHandler(dic *di.Container) *DeckHandler {
	return &DeckHandler{
		Songs:   *dic.Songs,
		Liturgy: *dic.Liturgy,
	}
}

func (h *DeckHandler) PostDeck(c *gin.Context) {
	var deck dtos.DeckRequest
	if err := c.ShouldBind(&deck); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := deck.Validate(); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error})
		return
	}

	textDeck, ok := services.BuildTextSlides(deck, h.Songs, h.Liturgy)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get lyrics"})
		return
	}

	uid := common.GetRandomString(6)
	extension := ""
	var file io.Reader
	var contents []core.ContentSlide
	var err error

	switch deck.Format {
	case "png+zip":
		extension = ".zip"
		file, contents, err = core.BuildImages(textDeck, deck.GetPageConfig())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

	case "txt":
		extension = ".txt"
		text := core.Tugalize(textDeck)
		file = strings.NewReader(text)

	default:
		extension = ".pdf"
		file, contents, err = core.BuildPDF(textDeck, services.GetPageConfig(deck))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	fileName := uid + extension
	common.SaveTemporaryFile(file, fileName)

	if !deck.Contents {
		contents = nil
	}

	resp := dtos.NewDeckResponse(common.GetPublicURL(c, fileName), contents)
	c.JSON(http.StatusOK, resp)
}
