package dtos

import (
	"errors"
	"regexp"

	"github.com/hejmsdz/goslides/core"
)

type DeckRequest struct {
	Date          string     `json:"date"`
	Items         []DeckItem `json:"items"`
	Hints         bool       `json:"hints"`
	Ratio         string     `json:"ratio"`
	FontSize      int        `json:"fontSize"`
	VerticalAlign string     `json:"verticalAlign"`
	Format        string     `json:"format"`
	Contents      bool       `json:"contents"`
}

type DeckItem struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Contents []string `json:"contents"`
	Order    []int    `json:"order"`
}

type DeckResponse struct {
	URL      string              `json:"url"`
	Contents []core.ContentSlide `json:"contents"`
}

func NewDeckResponse(url string, contents []core.ContentSlide) DeckResponse {
	return DeckResponse{
		URL:      url,
		Contents: contents,
	}
}

var dateRegexp = regexp.MustCompile(`^20\d\d-[0-1]\d-[0-3]\d$`)

func (d DeckRequest) Validate() error {
	if !dateRegexp.MatchString(d.Date) {
		return errors.New("invalid date")
	}

	if len(d.Items) == 0 {
		return errors.New("items are empty")
	}

	if len(d.Items) > 100 {
		return errors.New("too many items")
	}

	if d.FontSize > 0 && d.FontSize < 36 {
		return errors.New("font size too small")
	}

	if d.FontSize > 72 {
		return errors.New("font size too large")
	}

	if d.Ratio != "" && d.Ratio != "16:9" && d.Ratio != "4:3" {
		return errors.New("unsupported aspect ratio")
	}

	if d.VerticalAlign != "" && d.VerticalAlign != "top" && d.VerticalAlign != "bottom" && d.VerticalAlign != "center" {
		return errors.New("unsupported vertical align")
	}

	for _, item := range d.Items {
		if err := item.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (i DeckItem) Validate() error {
	// TODO

	return nil
}
