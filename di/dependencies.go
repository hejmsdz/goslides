package di

import (
	"github.com/hejmsdz/goslides/services"
	"gorm.io/gorm"
)

type Container struct {
	DB      *gorm.DB
	Auth    *services.AuthService
	Songs   *services.SongsService
	Liturgy *services.LiturgyService
	Live    *services.LiveService
	Users   *services.UsersService
}

func NewContainer(db *gorm.DB) *Container {
	users := services.NewUsersService(db)
	idTokenValidator := services.NewGoogleIDTokenValidator()

	return &Container{
		DB:      db,
		Auth:    services.NewAuthService(db, users, idTokenValidator),
		Songs:   services.NewSongsService(db),
		Liturgy: services.NewLiturgyService(),
		Live:    services.NewLiveService(),
		Users:   users,
	}
}

func NewTestContainer(db *gorm.DB) *Container {
	users := services.NewUsersService(db)
	idTokenValidator := services.NewMockIDTokenValidator()

	return &Container{
		DB:      db,
		Auth:    services.NewAuthService(db, users, idTokenValidator),
		Songs:   services.NewSongsService(db),
		Liturgy: services.NewLiturgyService(),
		Live:    services.NewLiveService(),
		Users:   users,
	}
}
