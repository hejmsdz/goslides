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
	songs := services.NewSongsService(db)
	liturgy := services.NewLiturgyService()

	return &Container{
		DB:      db,
		Auth:    services.NewAuthService(db, users, idTokenValidator),
		Songs:   songs,
		Liturgy: liturgy,
		Live:    services.NewLiveService(songs, liturgy),
		Users:   users,
	}
}

func NewTestContainer(db *gorm.DB) *Container {
	users := services.NewUsersService(db)
	idTokenValidator := services.NewMockIDTokenValidator()
	songs := services.NewSongsService(db)
	liturgy := services.NewLiturgyService()

	return &Container{
		DB:      db,
		Auth:    services.NewAuthService(db, users, idTokenValidator),
		Songs:   songs,
		Liturgy: liturgy,
		Live:    services.NewLiveService(songs, liturgy),
		Users:   users,
	}
}
