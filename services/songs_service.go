package services

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hejmsdz/goslides/common"
	"github.com/hejmsdz/goslides/dtos"
	"github.com/hejmsdz/goslides/models"
	"gorm.io/gorm"
)

type SongsService struct {
	db    *gorm.DB
	auth  *AuthService
	teams *TeamsService
}

func NewSongsService(db *gorm.DB, auth *AuthService, teams *TeamsService) *SongsService {
	return &SongsService{db, auth, teams}
}

func FilterByUserTeams(user *models.User) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if user == nil {
			return db.Where("team_id IS NULL")
		} else {
			return db.Joins("LEFT JOIN user_teams ON user_teams.team_id = songs.team_id").
				Where("songs.team_id IS NULL OR user_teams.user_id = ?", user.ID)
		}
	}
}

func (s SongsService) GetSong(uuidString string, user *models.User) (*models.Song, error) {
	var song models.Song

	uuid, err := uuid.Parse(uuidString)
	if err != nil {
		return nil, common.NewAPIError(400, "invalid id", err)
	}

	err = s.db.Preload("Team").Where("uuid", uuid).Take(&song).Error
	if err != nil {
		return nil, common.NewAPIError(404, "song not found", err)
	}

	if !s.auth.Can(user, "read", &song) {
		return nil, common.NewAPIError(403, "forbidden", nil)
	}

	return &song, nil
}

func (s SongsService) FilterSongs(query string, user *models.User) []models.Song {
	querySlug := common.Slugify(query)

	var songs []models.Song
	s.db.Scopes(FilterByUserTeams(user)).
		Preload("Team").
		Omit("lyrics").
		Where("slug LIKE ?", "%"+querySlug+"%").
		Order("title ASC, subtitle ASC").
		Find(&songs)

	return songs
}

func (s SongsService) CreateSong(input dtos.SongRequest, user *models.User) (*models.Song, error) {
	team, err := s.teams.GetTeam(input.Team)
	if err != nil {
		return nil, common.NewAPIError(403, "team not found", err)
	}

	song := &models.Song{
		Title:    input.Title,
		Subtitle: sql.NullString{String: input.Subtitle, Valid: input.Subtitle != ""},
		Lyrics:   strings.Join(input.Lyrics, "\n\n"),
	}

	if team != nil {
		song.Team = team
		song.TeamID = &team.ID
	}

	if !s.auth.Can(user, "create", song) {
		return nil, common.NewAPIError(403, "forbidden", nil)
	}

	err = s.db.Create(song).Error
	if err != nil {
		return nil, common.NewAPIError(500, "failed to create a song", err)
	}

	return song, nil
}

func (s SongsService) UpdateSong(id string, input dtos.SongRequest, user *models.User) (*models.Song, error) {
	song, err := s.GetSong(id, user)
	if err != nil {
		fmt.Printf("did not get song %+v\n", err)
		return nil, err
	}

	if !s.auth.Can(user, "update", song) {
		fmt.Printf("no permission to update song %+v\n", err)
		return nil, common.NewAPIError(403, "forbidden", nil)
	}

	song.Title = input.Title
	song.Subtitle = sql.NullString{String: input.Subtitle, Valid: input.Subtitle != ""}
	song.Lyrics = strings.Join(input.Lyrics, "\n\n")

	err = s.db.Save(&song).Error
	if err != nil {
		fmt.Printf("failed to save song %+v\n", err)
		return nil, common.NewAPIError(500, "failed to save", err)
	}

	return song, nil
}

func (s SongsService) DeleteSong(id string, user *models.User) error {
	song, err := s.GetSong(id, user)
	if err != nil {
		return err
	}

	if !s.auth.Can(user, "delete", song) {
		return common.NewAPIError(403, "forbidden", nil)
	}

	err = s.db.Delete(song).Error
	if err != nil {
		return common.NewAPIError(500, "failed to delete", err)
	}

	return nil
}
