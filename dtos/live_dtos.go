package dtos

import (
	"errors"
	"fmt"

	"github.com/hejmsdz/goslides/common"
)

type JsonObject map[string]interface{}

type Event struct {
	Type string
	Data JsonObject
}

type LiveSessionDTO struct {
	Deck        DeckRequest `json:"deck"`
	CurrentPage int         `json:"currentPage"`
}

func (ls LiveSessionDTO) Validate() error {
	if ls.CurrentPage < 0 {
		return errors.New("invalid current page")
	}

	return ls.Deck.Validate()
}

type LiveSessionRequest = LiveSessionDTO

type LiveSessionResponse struct {
	ID    string `json:"id"`
	URL   string `json:"url"`
	Token string `json:"token"`
}

func NewLiveSessionResponse(id string, token string) LiveSessionResponse {
	return LiveSessionResponse{
		ID:    id,
		URL:   common.GetFrontendURL(fmt.Sprintf("live/%s", id)),
		Token: token,
	}
}
