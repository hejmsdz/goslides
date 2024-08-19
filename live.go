package main

type JsonObject map[string]interface{}

type Event struct {
	Type string
	Data map[string]interface{}
}

type LiveSession struct {
	Deck        Deck `json:"deck"`
	CurrentPage int  `json:"currentPage"`
	members     []chan Event
}

var LiveSessions = map[string]*LiveSession{}

func (ls *LiveSession) ReplaceDeck(deck Deck, currentPage int) {
	ls.Deck = deck
	ls.CurrentPage = currentPage

	for _, member := range ls.members {
		member <- Event{"start", JsonObject{"deck": deck, "currentPage": currentPage}}
	}
}

func (ls *LiveSession) ChangePage(currentPage int) {
	ls.CurrentPage = currentPage

	for _, member := range ls.members {
		member <- Event{"changePage", JsonObject{"page": currentPage}}
	}
}

func (ls *LiveSession) AddMember() chan Event {
	memberChannel := make(chan Event)
	ls.members = append(ls.members, memberChannel)

	return memberChannel
}

func (ls *LiveSession) RemoveMember(memberChannel chan Event) {
	close(memberChannel)

	for i, v := range ls.members {
		if v == memberChannel {
			ls.members = append(ls.members[:i], ls.members[i+1:]...)
		}
	}
}
