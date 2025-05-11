package dtos

import (
	"errors"

	"github.com/hejmsdz/goslides/common"
	"github.com/hejmsdz/goslides/models"
)

type JsonObject map[string]interface{}

type Event struct {
	Type string
	Data JsonObject
}

type LiveSessionRequest struct {
	Deck        DeckRequest `json:"deck"`
	CurrentPage int         `json:"currentPage"`
}

func (ls LiveSessionRequest) Validate() error {
	if ls.CurrentPage < 0 {
		return errors.New("invalid current page")
	}

	return ls.Deck.Validate()
}

type LiveSessionResponse struct {
	ID    string `json:"id"`
	URL   string `json:"url"`
	Token string `json:"token"`
}

func NewLiveSessionResponse(id string, token string) LiveSessionResponse {
	return LiveSessionResponse{
		ID:    id,
		URL:   common.GetFrontendURL(id),
		Token: token,
	}
}

type LiveSessionStatusResponse struct {
	URL         string `json:"url"`
	CurrentPage int    `json:"currentPage"`
}

func NewLiveSessionStatusResponse(session *models.LiveSession) LiveSessionStatusResponse {
	return LiveSessionStatusResponse{
		URL:         session.URL,
		CurrentPage: session.CurrentPage,
	}
}
