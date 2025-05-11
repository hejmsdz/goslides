package repos

import (
	"net/http"
	"time"

	"github.com/hejmsdz/goslides/common"
	"github.com/hejmsdz/goslides/models"
	"gorm.io/gorm"
)

type SQLNonceRepo struct {
	db *gorm.DB
}

func NewSQLNonceRepo(db *gorm.DB) *SQLNonceRepo {
	return &SQLNonceRepo{db: db}
}

func (r *SQLNonceRepo) GetUserIdFromNonce(token string) (uint, error) {
	var nonce models.Nonce
	result := r.db.Where("token = ? AND expires_at > current_timestamp", token).Take(&nonce)
	if result.Error != nil {
		return 0, common.NewAPIError(http.StatusUnauthorized, "invalid nonce", result.Error)
	}

	r.db.Delete(&nonce)

	return nonce.UserID, nil
}

func (r *SQLNonceRepo) CreateNonce(nonce string, userID uint, expiration time.Duration) error {
	return r.db.Create(&models.Nonce{
		Token:     nonce,
		UserID:    userID,
		ExpiresAt: time.Now().Add(expiration),
	}).Error
}
