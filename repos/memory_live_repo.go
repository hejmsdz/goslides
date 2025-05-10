package repos

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/hejmsdz/goslides/dtos"
	"github.com/hejmsdz/goslides/models"
)

type memoryLiveSession struct {
	models.LiveSession
	subscribers []chan *dtos.Event
}

type MemoryLiveRepo struct {
	sessions map[string]*memoryLiveSession
}

func NewMemoryLiveRepo() *MemoryLiveRepo {
	return &MemoryLiveRepo{
		sessions: make(map[string]*memoryLiveSession),
	}
}

func (r *MemoryLiveRepo) CleanUp(minUpdatedAt time.Time) ([]string, error) {
	filenames := make([]string, 0)

	for key, session := range r.sessions {
		if session.UpdatedAt.Before(minUpdatedAt) && len(session.subscribers) == 0 {
			filenames = append(filenames, session.FileName)
			delete(r.sessions, key)
		}
	}

	return filenames, nil
}

func (r *MemoryLiveRepo) GetSession(key string) (*models.LiveSession, error) {
	session, ok := r.sessions[key]
	if !ok {
		return nil, errors.New("session not found")
	}

	return &session.LiveSession, nil
}

func (r *MemoryLiveRepo) CreateSession(session *models.LiveSession) (string, error) {
	for {
		key := fmt.Sprintf("%04d", rand.Intn(10000))
		if _, ok := r.sessions[key]; !ok {
			r.sessions[key] = &memoryLiveSession{
				LiveSession: *session,
				subscribers: make([]chan *dtos.Event, 0),
			}
			return key, nil
		}
	}
}

func (r *MemoryLiveRepo) ChangeSessionPage(key string, currentPage int) error {
	r.sessions[key].CurrentPage = currentPage
	return nil
}

func (r *MemoryLiveRepo) Subscribe(key string) (chan *dtos.Event, func(), error) {
	ch := make(chan *dtos.Event)
	session, ok := r.sessions[key]

	if !ok {
		return nil, nil, errors.New("session not found")
	}

	session.subscribers = append(session.subscribers, ch)

	return ch, func() {
		close(ch)
		for i, s := range session.subscribers {
			if s == ch {
				session.subscribers = append(session.subscribers[:i], session.subscribers[i+1:]...)
			}
		}
	}, nil
}

func (r *MemoryLiveRepo) PublishEvent(key string, event *dtos.Event) error {
	session, ok := r.sessions[key]

	if !ok {
		return errors.New("session not found")
	}

	for _, s := range session.subscribers {
		s <- event
	}
	return nil
}

func (r *MemoryLiveRepo) DeleteSession(key string) error {
	delete(r.sessions, key)
	return nil
}

func (r *MemoryLiveRepo) UpdateSession(key string, session *models.LiveSession) error {
	r.sessions[key].LiveSession = *session
	return nil
}
