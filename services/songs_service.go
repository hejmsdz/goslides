package services

import (
	"database/sql"
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

func (s SongsService) GetSong(uuidString string, user *models.User) (*models.Song, error) {
	var song models.Song

	uuid, err := uuid.Parse(uuidString)
	if err != nil {
		return nil, common.NewAPIError(400, "invalid id", err)
	}

	err = s.db.Preload("Team").Preload("OverriddenSong").Where("uuid", uuid).Take(&song).Error
	if err != nil {
		return nil, common.NewAPIError(404, "song not found", err)
	}

	if !s.auth.Can(user, "read", &song) {
		return nil, common.NewAPIError(403, "forbidden", nil)
	}

	return &song, nil
}

func (s SongsService) FilterSongs(query string, user *models.User, teamUUID string) []models.Song {
	querySlug := common.Slugify(query)

	var songs []models.Song

	db := s.db
	if teamUUID == "" {
		db = s.db.Where("team_id IS NULL")
	} else {
		team, err := s.teams.GetUserTeam(user, teamUUID)
		if err != nil {
			return songs
		}

		db = db.Joins("LEFT JOIN songs AS overrides ON overrides.overridden_song_id = songs.id AND overrides.team_id = ? AND overrides.deleted_at IS NULL", team.ID).
			Where("songs.team_id = ? OR (songs.team_id IS NULL AND overrides.id IS NULL)", team.ID)
	}

	db.Preload("Team").
		Omit("lyrics").
		Where("songs.slug LIKE ?", "%"+querySlug+"%").
		Order("title ASC, subtitle ASC").
		Find(&songs)

	return songs
}

func (s SongsService) CreateSong(input dtos.SongRequest, user *models.User) (*models.Song, error) {
	team, err := s.teams.GetUserTeamAllowingEmptyForAdmin(user, input.TeamID)
	if err != nil {
		return nil, common.NewAPIError(404, "team not found", err)
	}

	song := &models.Song{
		Title:       input.Title,
		Subtitle:    sql.NullString{String: input.Subtitle, Valid: input.Subtitle != ""},
		Lyrics:      strings.Join(input.Lyrics, "\n\n"),
		CreatedByID: user.ID,
		UpdatedByID: user.ID,
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
		return nil, err
	}

	if !s.auth.Can(user, "update", song) {
		return nil, common.NewAPIError(403, "forbidden", nil)
	}

	newTeam, err := s.teams.GetUserTeamAllowingEmptyForAdmin(user, input.TeamID)
	if err != nil {
		return nil, common.NewAPIError(404, "team not found", err)
	}

	song.Title = input.Title
	song.Subtitle = sql.NullString{String: input.Subtitle, Valid: input.Subtitle != ""}
	song.Lyrics = strings.Join(input.Lyrics, "\n\n")
	song.UpdatedByID = user.ID

	if newTeam == nil {
		song.Team = nil
		song.TeamID = nil
	} else {
		song.Team = newTeam
		song.TeamID = &(newTeam.ID)
	}

	err = s.db.Save(&song).Error
	if err != nil {
		return nil, common.NewAPIError(500, "failed to save", err)
	}

	return song, nil
}

func (s SongsService) OverrideSong(id string, input dtos.SongRequest, user *models.User) (*models.Song, error) {
	song, err := s.GetSong(id, user)
	if err != nil {
		return nil, err
	}

	if song.TeamID != nil {
		return nil, common.NewAPIError(409, "song already overridden", nil)
	}

	team, err := s.teams.GetUserTeam(user, input.TeamID)
	if err != nil || team == nil || !s.auth.UserBelongsToTeam(user, team.ID) {
		return nil, common.NewAPIError(404, "team not found", err)
	}

	var count int64
	s.db.Model(&models.Song{}).
		Where("overriden_song_id = ?", song.ID).
		Where("team_id = ?", team.ID).
		Count(&count)

	if count > 0 {
		return nil, common.NewAPIError(409, "song already overridden", nil)
	}

	newSong, err := s.CreateSong(input, user)
	if err != nil {
		return nil, err
	}

	newSong.OverriddenSong = song
	newSong.OverriddenSongID = &song.ID

	err = s.db.Save(newSong).Error
	if err != nil {
		return nil, err
	}

	return newSong, nil
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
