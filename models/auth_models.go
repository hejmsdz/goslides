package models

import (
	"time"

	"github.com/hejmsdz/goslides/common"
	"gorm.io/gorm"
)

type RefreshToken struct {
	gorm.Model
	Token     string `gorm:"uniqueIndex"`
	ExpiresAt time.Time
	UserID    uint `gorm:"not null"`
	User      User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

const refreshTokenLength = 32
const refreshTokenValidity = time.Duration(time.Hour * 24 * 30)

func NewRefreshToken(userID uint) *RefreshToken {
	token, err := common.GetSecureRandomString(refreshTokenLength)
	if err != nil {
		return nil
	}

	rt := RefreshToken{
		Token:     token,
		ExpiresAt: time.Now().Add(refreshTokenValidity),
		UserID:    userID,
	}

	return &rt
}

func (rt *RefreshToken) Regenerate() error {
	newToken, err := common.GetSecureRandomString(refreshTokenLength)
	if err != nil {
		return err
	}

	rt.Token = newToken
	rt.ExpiresAt = time.Now().Add(refreshTokenValidity)

	return nil
}

type Nonce struct {
	gorm.Model
	Token     string `gorm:"uniqueIndex"`
	ExpiresAt time.Time
	UserID    uint  `gorm:"not null"`
	User      *User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

const nonceLength = 32
const nonceValidity = time.Duration(time.Second * 30)

func NewNonce(userID uint) *Nonce {
	token, err := common.GetSecureRandomString(nonceLength)
	if err != nil {
		return nil
	}

	nonce := Nonce{
		Token:     token,
		ExpiresAt: time.Now().Add(nonceValidity),
		UserID:    userID,
	}

	return &nonce
}
