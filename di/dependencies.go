package di

import (
	"github.com/hejmsdz/goslides/services"
	"gorm.io/gorm"
)

type Container struct {
	DB      *gorm.DB
	Songs   *services.SongsService
	Liturgy *services.LiturgyService
	Live    *services.LiveService
}

func NewContainer(db *gorm.DB) *Container {
	return &Container{
		DB:      db,
		Songs:   services.NewSongsService(db),
		Liturgy: services.NewLiturgyService(),
		Live:    services.NewLiveService(),
	}
}
