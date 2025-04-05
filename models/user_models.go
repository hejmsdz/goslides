package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UUID        uuid.UUID `gorm:"uniqueIndex"`
	Email       string    `gorm:"uniqueIndex"`
	DisplayName string
	IsAdmin     bool
	Teams       []*Team `gorm:"many2many:user_teams;"`
}

func (u *User) BeforeSave(tx *gorm.DB) (err error) {
	if u.UUID == uuid.Nil {
		u.UUID = uuid.New()
	}

	return nil
}
