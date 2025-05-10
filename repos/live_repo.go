package repos

import (
	"time"

	"github.com/hejmsdz/goslides/dtos"
	"github.com/hejmsdz/goslides/models"
)

type LiveSessionRepo interface {
	GetSession(key string) (*models.LiveSession, error)
	CreateSession(session *models.LiveSession) (string, error)
	UpdateSession(key string, session *models.LiveSession) error
	ChangeSessionPage(key string, currentPage int) error
	DeleteSession(key string) error
	PublishEvent(key string, event *dtos.Event) error
	Subscribe(key string) (chan *dtos.Event, func(), error)
	CleanUp(minUpdatedAt time.Time) ([]string, error)
}
