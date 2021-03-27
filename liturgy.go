package main

import (
	"fmt"
	"net/http"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

type Liturgy struct {
	Psalm            string
	Acclamation      string
	AcclamationVerse string
}

type LiturgyDB struct {
	Days map[string]Liturgy
}

const PSALM = "PSALM"
const ACCLAMATION = "ACCLAMATION"

func getAttributeValue(node *html.Node, key string) string {
	for _, attribute := range node.Attr {
		if attribute.Key == key {
			return attribute.Val
		}
	}
	return ""
}

func findTabId(doc *html.Node, label string) string {
	rootTabSelector, _ := cascadia.Parse("article > .nav-tabs li:first-child a")
	rootTab := cascadia.Query(doc, rootTabSelector)
	rootTabId := getAttributeValue(rootTab, "href")

	tabSelector, _ := cascadia.Parse(fmt.Sprintf("%s a[data-toggle=tab]", rootTabId))
	tabs := cascadia.QueryAll(doc, tabSelector)
	for _, tab := range tabs {
		if tab.FirstChild.Data == label {
			return getAttributeValue(tab, "href")
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

	return node.FirstChild.Data, true
}

func getAcclamation(doc *html.Node) (string, string, bool) {
	tabId := findTabId(doc, "Aklamacja")
	if tabId == "" {
		return "", "", false
	}

	alleluiaSelector, _ := cascadia.Parse(fmt.Sprintf("%s h4 em", tabId))
	verseSelector, _ := cascadia.Parse(fmt.Sprintf("%s p", tabId))
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

	return alleluia, verse, true
}

func GetLiturgy(date string) (Liturgy, bool) {
	liturgy := Liturgy{}

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

func (ldb *LiturgyDB) Initialize() {
	ldb.Days = make(map[string]Liturgy)
}

func (ldb LiturgyDB) GetDay(date string) (Liturgy, bool) {
	liturgy, ok := ldb.Days[date]
	if ok {
		return liturgy, true
	}

	liturgy, ok = GetLiturgy(date)
	if ok {
		ldb.Days[date] = liturgy
	}

	return liturgy, ok
}
