package models

import (
	"github.com/google/uuid"
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
