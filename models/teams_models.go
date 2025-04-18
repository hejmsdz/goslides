package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/hejmsdz/goslides/common"
	"gorm.io/gorm"
)

type Team struct {
	gorm.Model
	UUID  uuid.UUID `gorm:"uniqueIndex"`
	Name  string
	Users []*User `gorm:"many2many:user_teams;"`
}

func (t *Team) BeforeSave(tx *gorm.DB) (err error) {
	if t.UUID == uuid.Nil {
		t.UUID = uuid.New()
	}

	return nil
}

type Invitation struct {
	gorm.Model
	TeamID    uint
	Team      *Team
	Token     string    `gorm:"not null;unique"`
	ExpiresAt time.Time `gorm:"not null"`
}

func (i *Invitation) BeforeSave(tx *gorm.DB) (err error) {
	if i.Token == "" {
		token, err := common.GetSecureRandomString(32)
		if err != nil {
			return err
		}

		i.Token = token
	}

	return nil
}
