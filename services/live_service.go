package services

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/hejmsdz/goslides/common"
	"github.com/hejmsdz/goslides/core"
	"github.com/hejmsdz/goslides/dtos"
	"github.com/hejmsdz/goslides/models"
	"github.com/hejmsdz/goslides/repos"
)

type LiveService struct {
	repo            repos.LiveSessionRepo
	Songs           *SongsService
	Liturgy         *LiturgyService
	maxIdleDuration time.Duration
}

type MemberChannel chan *dtos.Event

func NewLiveService(songs *SongsService, liturgy *LiturgyService, repo repos.LiveSessionRepo) *LiveService {
	maxIdleDuration, err := time.ParseDuration(os.Getenv("LIVE_SESSION_MAX_IDLE_DURATION"))
	if err != nil {
		panic(fmt.Sprintf("failed to read LIVE_SESSION_MAX_IDLE_DURATION: %s", err.Error()))
	}

	return &LiveService{
		repo:            repo,
		Songs:           songs,
		Liturgy:         liturgy,
		maxIdleDuration: maxIdleDuration,
	}
}

func (l *LiveService) GetSession(key string) (*models.LiveSession, bool) {
	session, err := l.repo.GetSession(key)
	if err != nil {
		return nil, false
	}

	return session, true
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

func (l *LiveService) CreateSession(input dtos.LiveSessionRequest, user *models.User) (string, *models.LiveSession, error) {
	fileName, err := l.GenerateLiveSessionDeck(input, user)
	if err != nil {
		return "", nil, err
	}

	token, err := common.GetSecureRandomString(16)
	if err != nil {
		return "", nil, err
	}

	session := &models.LiveSession{
		URL:         common.GetPublicURL(fileName),
		CurrentPage: input.CurrentPage,
		Token:       token,
		FileName:    fileName,
		UpdatedAt:   time.Now(),
	}

	key, err := l.repo.CreateSession(session)
	if err != nil {
		return "", nil, err
	}

	return key, session, nil
}

func (l *LiveService) DeleteSession(key string) error {
	session, ok := l.GetSession(key)
	if !ok {
		return errors.New("session not found")
	}

	l.repo.PublishEvent(key, &dtos.Event{
		Type: "delete",
		Data: dtos.JsonObject{},
	})

	err := l.repo.DeleteSession(key)
	if err != nil {
		return err
	}

	common.DeletePublicFile(session.FileName)

	return nil
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

	common.DeletePublicFile(session.FileName)

	session.FileName = fileName
	session.URL = common.GetPublicURL(fileName)
	session.CurrentPage = input.CurrentPage
	session.UpdatedAt = time.Now()

	err = l.repo.UpdateSession(key, session)
	if err != nil {
		return err
	}

	l.repo.PublishEvent(key, &dtos.Event{
		Type: "start",
		Data: dtos.JsonObject{
			"url":         session.URL,
			"currentPage": session.CurrentPage,
		},
	})

	return nil
}

func (l *LiveService) ChangeSessionPage(key string, currentPage int) {
	l.repo.ChangeSessionPage(key, currentPage)
	l.repo.PublishEvent(key, &dtos.Event{
		Type: "changePage",
		Data: dtos.JsonObject{"page": currentPage},
	})
}

func (l *LiveService) Subscribe(key string) (MemberChannel, func(), error) {
	return l.repo.Subscribe(key)
}

func (l *LiveService) CleanUp() int {
	minUpdatedAt := time.Now().Add(-1 * l.maxIdleDuration)
	filenames, err := l.repo.CleanUp(minUpdatedAt)
	if err != nil {
		return 0
	}

	for _, filename := range filenames {
		common.DeletePublicFile(filename)
	}

	return len(filenames)
}
