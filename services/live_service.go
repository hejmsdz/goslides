package services

import (
	"crypto/subtle"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/hejmsdz/goslides/common"
	"github.com/hejmsdz/goslides/dtos"
)

type LiveService struct {
	sessions map[string]*LiveSession
}

type MemberChannel chan dtos.Event

type LiveSession struct {
	Deck            dtos.DeckRequest `json:"deck"`
	CurrentPage     int              `json:"currentPage"`
	Token           string
	members         []MemberChannel
	expirationTimer *time.Timer
}

func NewLiveService() *LiveService {
	return &LiveService{
		sessions: make(map[string]*LiveSession),
	}
}

func (l *LiveService) GetSession(id string) (*LiveSession, bool) {
	session, ok := l.sessions[id]

	return session, ok
}

func (l *LiveService) ValidateToken(id string, token string) bool {
	session, ok := l.GetSession(id)
	if !ok {
		return false
	}

	correctTokenBytes := []byte(session.Token)
	tokenBytes := []byte(token)

	return subtle.ConstantTimeCompare(correctTokenBytes, tokenBytes) == 1
}

func (l *LiveService) GenerateLiveSessionId() string {
	for {
		id := fmt.Sprintf("%04d", rand.IntN(10000))
		if _, ok := l.sessions[id]; !ok {
			return id
		}
	}
}

func (l *LiveService) CreateSession(id string, input dtos.LiveSessionRequest) *LiveSession {
	session := &LiveSession{
		Deck:        input.Deck,
		CurrentPage: input.CurrentPage,
		Token:       common.GetRandomString(16),
		members:     make([]MemberChannel, 0),
	}

	l.sessions[id] = session
	l.ExtendSessionTime(id)

	return session
}

func (l *LiveService) ExtendSessionTime(id string) {
	session, ok := l.GetSession(id)
	if !ok {
		return
	}

	if session.expirationTimer != nil {
		session.expirationTimer.Stop()
	}

	session.expirationTimer = time.AfterFunc(2*time.Hour, func() {
		l.DeleteSession(id)
	})
}

func (l *LiveService) DeleteSession(id string) {
	session, ok := l.GetSession(id)
	if !ok {
		return
	}

	for _, member := range session.members {
		close(member)
	}

	delete(l.sessions, id)
}

func (l *LiveService) UpdateSession(id string, input dtos.LiveSessionRequest) {
	session, ok := l.GetSession(id)
	if !ok {
		return
	}

	session.Deck = input.Deck
	session.CurrentPage = input.CurrentPage

	for _, member := range session.members {
		member <- dtos.Event{
			Type: "start",
			Data: dtos.JsonObject{"deck": input.Deck, "currentPage": input.CurrentPage},
		}
	}
}

func (l *LiveService) ChangeSessionPage(id string, currentPage int) {
	session, ok := l.GetSession(id)
	if !ok {
		return
	}

	for _, member := range session.members {
		member <- dtos.Event{
			Type: "changePage",
			Data: dtos.JsonObject{"page": currentPage},
		}
	}
}

func (ls *LiveSession) AddMember() MemberChannel {
	memberChannel := make(MemberChannel)
	ls.members = append(ls.members, memberChannel)

	return memberChannel
}

func (ls *LiveSession) RemoveMember(memberChannel MemberChannel) {
	close(memberChannel)

	for i, v := range ls.members {
		if v == memberChannel {
			ls.members = append(ls.members[:i], ls.members[i+1:]...)
		}
	}
}
