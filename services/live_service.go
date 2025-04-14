package services

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"math/rand/v2"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/hejmsdz/goslides/common"
	"github.com/hejmsdz/goslides/core"
	"github.com/hejmsdz/goslides/dtos"
	"github.com/hejmsdz/goslides/models"
)

type LiveService struct {
	sessions map[string]*LiveSession
	Songs    *SongsService
	Liturgy  *LiturgyService
}

type MemberChannel chan dtos.Event

type LiveSession struct {
	URL             string `json:"url"`
	CurrentPage     int    `json:"currentPage"`
	Token           string `json:"-"`
	fileName        string
	members         []MemberChannel
	expirationTimer *time.Timer
}

func NewLiveService(songs *SongsService, liturgy *LiturgyService) *LiveService {
	return &LiveService{
		sessions: make(map[string]*LiveSession),
		Songs:    songs,
		Liturgy:  liturgy,
	}
}

func (l *LiveService) GetSession(key string) (*LiveSession, bool) {
	session, ok := l.sessions[key]

	return session, ok
}

func (l *LiveService) ValidateToken(key string, token string) bool {
	session, ok := l.GetSession(key)
	if !ok {
		return false
	}

	correctTokenBytes := []byte(session.Token)
	tokenBytes := []byte(token)

	return subtle.ConstantTimeCompare(correctTokenBytes, tokenBytes) == 1
}

var sessionKeyRegexp = regexp.MustCompile(`^\d{4}$`)

func (l *LiveService) ValidateLiveSessionKey(key string) bool {
	return sessionKeyRegexp.MatchString(key)
}

func (l *LiveService) GenerateLiveSessionKey() string {
	for {
		key := fmt.Sprintf("%04d", rand.IntN(10000))
		if _, ok := l.sessions[key]; !ok {
			return key
		}
	}
}

func (l *LiveService) GenerateLiveSessionDeck(input dtos.LiveSessionRequest, user *models.User) (string, error) {
	textDeck, ok := BuildTextSlides(input.Deck, l.Songs, l.Liturgy, user)
	if !ok {
		return "", errors.New("failed to build text deck")
	}

	file, _, err := core.BuildPDF(textDeck, GetPageConfig(input.Deck))
	if err != nil {
		return "", err
	}

	fileName := uuid.New().String() + ".pdf"
	return fileName, common.SavePublicFile(file, fileName)
}

func (l *LiveService) CreateSession(key string, input dtos.LiveSessionRequest, user *models.User) (*LiveSession, error) {
	fileName, err := l.GenerateLiveSessionDeck(input, user)
	if err != nil {
		return nil, err
	}

	token, err := common.GetSecureRandomString(16)
	if err != nil {
		return nil, err
	}

	session := &LiveSession{
		URL:         common.GetPublicURL(fileName),
		CurrentPage: input.CurrentPage,
		Token:       token,
		fileName:    fileName,
		members:     make([]MemberChannel, 0),
	}

	l.sessions[key] = session
	l.ExtendSessionTime(key)

	return session, nil
}

func (l *LiveService) ExtendSessionTime(key string) {
	session, ok := l.GetSession(key)
	if !ok {
		return
	}

	if session.expirationTimer != nil {
		session.expirationTimer.Stop()
	}

	session.expirationTimer = time.AfterFunc(2*time.Hour, func() {
		l.DeleteSession(key)
	})
}

func (l *LiveService) DeleteSession(key string) {
	session, ok := l.GetSession(key)
	if !ok {
		return
	}

	for _, member := range session.members {
		close(member)
	}

	common.DeletePublicFile(session.fileName)

	delete(l.sessions, key)
}

func (l *LiveService) UpdateSession(key string, input dtos.LiveSessionRequest, user *models.User) error {
	session, ok := l.GetSession(key)
	if !ok {
		return errors.New("session not found")
	}

	fileName, err := l.GenerateLiveSessionDeck(input, user)
	if err != nil {
		return err
	}

	common.DeletePublicFile(session.fileName)

	session.fileName = fileName
	session.URL = common.GetPublicURL(fileName)
	session.CurrentPage = input.CurrentPage

	for _, member := range session.members {
		member <- dtos.Event{
			Type: "start",
			Data: dtos.JsonObject{
				"url":         session.URL,
				"currentPage": session.CurrentPage,
			},
		}
	}

	return nil
}

func (l *LiveService) ChangeSessionPage(key string, currentPage int) {
	session, ok := l.GetSession(key)
	if !ok {
		return
	}

	session.CurrentPage = currentPage

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
