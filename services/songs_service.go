package services

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
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

func (s SongsService) GetSong(uuid string) (models.Song, error) {
	var song models.Song
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

// temporary
func (s SongsService) Import(n NotionSongsDB) {
	allSongs := n.FilterSongs("---")
	fmt.Printf("found %d songs\n", len(allSongs))
	ids := make([]string, 0)

	for _, ns := range allSongs {
		var dbSong models.Song
		result := s.db.Where("uuid", ns.Id).Take(&dbSong)
		if result.Error != nil {
			user := models.Song{
				UUID:  uuid.MustParse(ns.Id),
				Title: ns.Title,
				Subtitle: sql.NullString{
					Valid:  ns.Subtitle != "",
					String: ns.Subtitle,
				},
			}

			s.db.Create(&user)
			ids = append(ids, ns.Id)
			fmt.Printf("%s %s created\n", ns.Id, ns.Title)
		} else if ns.updatedAt.After(dbSong.UpdatedAt) {
			dbSong.Title = ns.Title
			dbSong.Subtitle = sql.NullString{
				Valid:  ns.Subtitle != "",
				String: ns.Subtitle,
			}

			s.db.Save(&dbSong)
			ids = append(ids, ns.Id)
			fmt.Printf("%s %s updated\n", ns.Id, ns.Title)
		} else {
			// fmt.Printf("%s %s is fresh\n", ns.Id, ns.Title)
		}
	}

	fmt.Printf("%d songs to import lyrics\n", len(ids))

	perBurst := 5
	bursts := int(math.Ceil(float64(len(ids)) / float64(perBurst)))
	for i := 0; i < bursts; i++ {
		endIdx := (i + 1) * perBurst
		if endIdx > len(ids) {
			endIdx = len(ids)
		}
		subIds := ids[i*perBurst : endIdx]
		n.LoadMissingVerses(subIds)

		fmt.Printf("%.1f%% / %d\n", float32(i)/float32(bursts)*100, len(ids))
		fmt.Printf("loading lyrics for %+v\n", subIds)

		for _, id := range subIds {
			lyrics := strings.Join(n.LyricsBlocks[id], "\n\n")
			s.db.Model(&models.Song{}).Where("uuid = ?", id).Update("lyrics", lyrics)
		}
	}

	fmt.Println("done!")
}
