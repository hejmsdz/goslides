package services

type LiveSessionStore interface {
	GetSession(id string) (*LiveSession, error)
	CreateSession(session *LiveSession) error
	UpdateSession(session *LiveSession) error
	DeleteSession(id string) error
}
