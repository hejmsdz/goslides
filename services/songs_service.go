package services

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/hejmsdz/goslides/common"
	"github.com/hejmsdz/goslides/dtos"
	"github.com/hejmsdz/goslides/models"
	"gorm.io/gorm"
)

type SongsService struct {
	db *gorm.DB
}

func NewSongsService(db *gorm.DB) *SongsService {
	return &SongsService{db}
}

func (s SongsService) GetSong(uuidString string) (models.Song, error) {
	var song models.Song

	uuid, err := uuid.Parse(uuidString)
	if err != nil {
		return song, errors.New("invalid id")
	}

	result := s.db.Where("uuid", uuid).Take(&song)

	if result.Error != nil {
		return song, errors.New("song not found")
	}

	return song, nil
}

func (s SongsService) FilterSongs(query string) []models.Song {
	querySlug := common.Slugify(query)

	var songs []models.Song
	s.db.Select("uuid", "title", "subtitle", "slug").
		Where("slug LIKE ?", "%"+querySlug+"%").
		Order("title ASC, subtitle ASC").
		Find(&songs)

	return songs
}

func (s SongsService) CreateSong(input dtos.SongRequest) (*models.Song, error) {
	song := models.Song{
		Title:    input.Title,
		Subtitle: sql.NullString{String: input.Subtitle, Valid: input.Subtitle != ""},
		Lyrics:   strings.Join(input.Lyrics, "\n\n"),
	}

	res := s.db.Create(&song)
	if res.Error != nil {
		return nil, errors.New("failed to create a song")
	}

	return &song, nil
}

func (s SongsService) UpdateSong(id string, input dtos.SongRequest) (*models.Song, error) {
	var song models.Song

	res := s.db.Where("uuid = ?", id).Take(&song)
	if res.Error != nil {
		return nil, errors.New("not found")
	}

	song.Title = input.Title
	song.Subtitle = sql.NullString{String: input.Subtitle, Valid: input.Subtitle != ""}
	song.Lyrics = strings.Join(input.Lyrics, "\n\n")

	res = s.db.Save(&song)
	if res.Error != nil {
		return nil, errors.New("failed to save")
	}

	return &song, nil
}

func (s SongsService) DeleteSong(id string) error {
	var song models.Song

	res := s.db.Where("uuid = ?", id).Take(&song)
	if res.Error != nil {
		return errors.New("song not found")
	}

	res = s.db.Delete(&song)
	if res.Error != nil {
		return errors.New("failed to delete")
	}

	return nil
}
