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
	Teams   *services.TeamsService
}

func NewContainer(db *gorm.DB) *Container {
	users := services.NewUsersService(db)
	idTokenValidator := services.NewGoogleIDTokenValidator()
	auth := services.NewAuthService(db, users, idTokenValidator)
	teams := services.NewTeamsService(db)
	songs := services.NewSongsService(db, auth, teams)
	liturgy := services.NewLiturgyService()

	return &Container{
		DB:      db,
		Auth:    auth,
		Songs:   songs,
		Liturgy: liturgy,
		Live:    services.NewLiveService(songs, liturgy),
		Users:   users,
		Teams:   teams,
	}
}

func NewTestContainer(db *gorm.DB) *Container {
	users := services.NewUsersService(db)
	idTokenValidator := services.NewMockIDTokenValidator()
	auth := services.NewAuthService(db, users, idTokenValidator)
	teams := services.NewTeamsService(db)
	songs := services.NewSongsService(db, auth, teams)
	liturgy := services.NewLiturgyService()

	return &Container{
		DB:      db,
		Auth:    auth,
		Songs:   songs,
		Liturgy: liturgy,
		Live:    services.NewLiveService(songs, liturgy),
		Users:   users,
		Teams:   teams,
	}
}
