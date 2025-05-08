package services

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/andybalholm/cascadia"
	"github.com/hejmsdz/goslides/dtos"
	"github.com/hejmsdz/goslides/repos"
	"golang.org/x/net/html"
)

type LiturgyService struct {
	repo repos.LiturgyRepo
}

func getAttributeValue(node *html.Node, key string) string {
	for _, attribute := range node.Attr {
		if attribute.Key == key {
			return attribute.Val
		}
	}
	return ""
}

func findTabId(doc *html.Node, label string) string {
	rootTabSelector, _ := cascadia.Parse("article > .nav-tabs li:first-child button")
	rootTab := cascadia.Query(doc, rootTabSelector)
	rootTabId := getAttributeValue(rootTab, "data-bs-target")

	tabSelector, _ := cascadia.Parse(fmt.Sprintf("%s ul[role=tablist] li button", rootTabId))
	tabs := cascadia.QueryAll(doc, tabSelector)
	for _, tab := range tabs {
		if tab.FirstChild.Data == label {
			return getAttributeValue(tab, "data-bs-target")
		}
	}
	return ""
}

func getPsalm(doc *html.Node) (string, bool) {
	tabId := findTabId(doc, "Psalm")
	if tabId == "" {
		return "", false
	}

	selector, _ := cascadia.Parse(fmt.Sprintf("%s h4 em", tabId))
	node := cascadia.Query(doc, selector)
	if node == nil {
		return "", false
	}

	return strings.TrimSpace(node.FirstChild.Data), true
}

func getAcclamation(doc *html.Node) (string, string, bool) {
	tabId := findTabId(doc, "Aklamacja")
	if tabId == "" {
		return "", "", false
	}

	alleluiaSelector, _ := cascadia.Parse(fmt.Sprintf("%s h4 em", tabId))
	verseSelector, _ := cascadia.Parse(fmt.Sprintf("%s h4 ~ p", tabId))
	alleluiaNode := cascadia.Query(doc, alleluiaSelector)
	verseNodes := cascadia.QueryAll(doc, verseSelector)
	if alleluiaNode == nil || len(verseNodes) == 0 {
		return "", "", false
	}

	alleluia := alleluiaNode.FirstChild.Data

	verse := ""
	for _, verseNode := range verseNodes {
		if verseNode.FirstChild.Type != html.NodeType(html.TextNode) {
			continue
		}
		for verseContent := verseNode.FirstChild; verseContent != nil; verseContent = verseContent.NextSibling {
			if verseContent.Type == html.NodeType(html.TextNode) {
				verse += verseContent.Data
			}
		}
	}

	return strings.TrimSpace(alleluia), strings.TrimSpace(verse), true
}

func (l LiturgyService) fetchLiturgy(date string) (dtos.LiturgyItems, bool) {
	liturgy := dtos.LiturgyItems{}

	url := fmt.Sprintf("https://niezbednik.niedziela.pl/liturgia/%s", date)
	res, err := http.Get(url)
	if err != nil {
		return liturgy, false
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return liturgy, false
	}

	doc, err := html.Parse(res.Body)
	if err != nil {
		return liturgy, false
	}

	var psalmOk, acclamationOk bool
	liturgy.Psalm, psalmOk = getPsalm(doc)
	liturgy.Acclamation, liturgy.AcclamationVerse, acclamationOk = getAcclamation(doc)

	return liturgy, psalmOk && acclamationOk
}

func NewLiturgyService(repo repos.LiturgyRepo) *LiturgyService {
	return &LiturgyService{
		repo: repo,
	}
}

func (l *LiturgyService) GetDay(date string) (dtos.LiturgyItems, bool) {
	liturgy, ok := l.repo.GetDay(date)
	if ok {
		return liturgy, true
	}

	liturgy, ok = l.fetchLiturgy(date)
	if ok {
		l.repo.StoreDay(date, liturgy)
	}

	return liturgy, ok
}
