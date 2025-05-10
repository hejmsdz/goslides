package repos

import "time"

type NonceRepo interface {
	GetUserIdFromNonce(token string) (uint, error)
	CreateNonce(token string, userID uint, expiration time.Duration) error
}
