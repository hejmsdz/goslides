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

func (s SongsService) getSongsQuery(query string, teamID uint, includeUnofficial bool) (*gorm.DB, error) {
	querySlug := common.Slugify(query)

	db := s.db.Debug().Model(&models.Song{})
	if teamID == 0 {
		db = db.Where("team_id IS NULL")
	} else {
		db = db.Joins("LEFT JOIN songs AS overrides ON overrides.overridden_song_id = songs.id AND overrides.team_id = ? AND overrides.deleted_at IS NULL", teamID).
			Where("songs.team_id = ? OR (songs.team_id IS NULL AND overrides.id IS NULL)", teamID)
	}

	if query != "" {
		db = db.Where("songs.slug LIKE ?", "%"+querySlug+"%")
	}

	if !includeUnofficial {
		db = db.Where("songs.is_unofficial = false")
	}

	db = db.Preload("Team").
		Omit("lyrics").
		Order("title ASC, subtitle ASC")

	return db, nil
}

func (s SongsService) FilterSongsPaginated(query string, user *models.User, teamUUID string, limit int, offset int) ([]models.Song, int64, error) {
	var songs []models.Song
	var teamID uint = 0
	includeUnofficial := false

	if teamUUID != "" {
		team, err := s.teams.GetUserTeam(user, teamUUID)
		if err != nil {
			return songs, 0, nil
		} else {
			teamID = team.ID
			includeUnofficial = team.CanAccessUnofficialSongs
		}
	}

	db, err := s.getSongsQuery(query, teamID, includeUnofficial)
	if err != nil {
		return nil, 0, err
	}

	if limit > 0 && offset >= 0 {
		err = db.Offset(offset).Limit(limit).Find(&songs).Error
	} else {
		err = db.Find(&songs).Error
	}

	if err != nil {
		return songs, 0, err
	}

	var total int64
	db, err = s.getSongsQuery(query, teamID, includeUnofficial)
	if err != nil {
		return nil, 0, err
	}

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	return songs, total, nil
}

func (s SongsService) FilterSongs(query string, user *models.User, teamUUID string) ([]models.Song, error) {
	songs, _, err := s.FilterSongsPaginated(query, user, teamUUID, -1, -1)
	return songs, err
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

	if user.IsAdmin {
		song.IsUnofficial = input.IsUnofficial
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

	if user.IsAdmin {
		song.IsUnofficial = input.IsUnofficial
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
		Where("overridden_song_id = ?", song.ID).
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
