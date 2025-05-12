package di

import (
	"github.com/hejmsdz/goslides/repos"
	"github.com/hejmsdz/goslides/services"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Container struct {
	DB      *gorm.DB
	Auth    *services.AuthService
	Songs   *services.SongsService
	Liturgy *services.LiturgyService
	Deck    *services.DeckService
	Live    *services.LiveService
	Users   *services.UsersService
	Teams   *services.TeamsService
}

func NewContainer(db *gorm.DB, redis *redis.Client) *Container {

	users := services.NewUsersService(db)
	idTokenValidator := services.NewGoogleIDTokenValidator()
	nonceRepo := repos.NewRedisNonceRepo(redis)
	auth := services.NewAuthService(db, users, idTokenValidator, nonceRepo)
	teams := services.NewTeamsService(db)
	songs := services.NewSongsService(db, auth, teams)
	liturgyRepo := repos.NewRedisLiturgyRepo(redis)
	liturgy := services.NewLiturgyService(liturgyRepo)
	liveRepo := repos.NewRedisLiveRepo(redis)
	deck := services.NewDeckService(songs, liturgy)

	return &Container{
		DB:      db,
		Auth:    auth,
		Songs:   songs,
		Liturgy: liturgy,
		Deck:    deck,
		Live:    services.NewLiveService(songs, liturgy, deck, liveRepo),
		Users:   users,
		Teams:   teams,
	}
}

func NewTestContainer(db *gorm.DB) *Container {
	users := services.NewUsersService(db)
	idTokenValidator := services.NewMockIDTokenValidator()
	nonceRepo := repos.NewSQLNonceRepo(db) // repos.NewRedisNonceRepo(redis)
	auth := services.NewAuthService(db, users, idTokenValidator, nonceRepo)
	teams := services.NewTeamsService(db)
	songs := services.NewSongsService(db, auth, teams)
	liturgy := services.NewLiturgyService(repos.NewMemoryLiturgyRepo())
	deck := services.NewDeckService(songs, liturgy)

	return &Container{
		DB:      db,
		Auth:    auth,
		Songs:   songs,
		Liturgy: liturgy,
		Deck:    deck,
		Live:    services.NewLiveService(songs, liturgy, deck, repos.NewMemoryLiveRepo()),
		Users:   users,
		Teams:   teams,
	}
}
