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
}

func (s *User) BeforeSave(tx *gorm.DB) (err error) {
	if s.UUID == uuid.Nil {
		s.UUID = uuid.New()
	}

	return nil
}
