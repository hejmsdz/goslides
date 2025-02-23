package main

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SQLSongsDB struct {
	db *gorm.DB
}

type DbSong struct {
	gorm.Model
	UUID     uuid.UUID `gorm:"uniqueIndex"`
	Title    string
	Subtitle sql.NullString
	Number   string
	Tags     string
	Slug     string
	Lyrics   string
}

func (s *DbSong) BeforeSave(tx *gorm.DB) (err error) {
	if s.UUID == uuid.Nil {
		s.UUID = uuid.New()
	}

	s.Slug = fmt.Sprintf("%s|%s|%s|%s", Slugify(s.Title), Slugify(s.Subtitle.String), s.Number, Slugify(s.Tags))

	return nil
}

func (sdb *SQLSongsDB) Initialize(dbPath string) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		panic("failed to connect")
	}

	db.AutoMigrate(&DbSong{})

	sdb.db = db
}

func (sdb SQLSongsDB) Reload() error {
	return nil
}

func (sdb SQLSongsDB) LoadMissingVerses(songIDs []string) error {
	return nil
}

func MapDbSong(s DbSong) Song {
	return Song{
		Id:         s.UUID.String(),
		Title:      s.Title,
		Subtitle:   s.Subtitle.String,
		Number:     s.Number,
		Slug:       s.Slug,
		Tags:       s.Tags,
		IsOrdinary: strings.Contains(s.Tags, ordinaryTag),
	}

}

func (sdb SQLSongsDB) GetSong(songID string) (SongWithLyrics, bool) {
	var song DbSong
	result := sdb.db.Where("uuid", songID).Take(&song)

	if result.Error != nil {
		return SongWithLyrics{}, false
	}

	lyrics := strings.Split(song.Lyrics, "\n\n")

	formattedLyrics := FormatLyrics(lyrics, song.Title, song.Number, GetLyricsOptions{Raw: true})

	return SongWithLyrics{
		Song:   MapDbSong(song),
		Lyrics: formattedLyrics,
	}, true
}

func (sdb SQLSongsDB) GetLyrics(songID string, options GetLyricsOptions) ([]string, bool) {
	var song DbSong
	result := sdb.db.Where("uuid", songID).Take(&song)

	if result.Error != nil {
		return nil, false
	}

	lyrics := strings.Split(song.Lyrics, "\n\n")

	return FormatLyrics(lyrics, song.Title, song.Number, options), true
}

func (sdb SQLSongsDB) FilterSongs(query string) []Song {
	if len(query) < 3 {
		return []Song{}
	}

	querySlug := Slugify(query)

	var dbSongs []DbSong
	sdb.db.Select("uuid", "title", "subtitle", "number", "tags", "slug").
		Where("slug LIKE ?", "%"+querySlug+"%").
		Order("title ASC, subtitle ASC").
		Find(&dbSongs)

	songs := make([]Song, len(dbSongs))

	for i, s := range dbSongs {
		songs[i] = MapDbSong(s)
	}

	return songs
}

func (sdb SQLSongsDB) CreateSong(input SongInput) (string, error) {
	dbSong := DbSong{
		Title:    input.Title,
		Subtitle: sql.NullString{String: input.Subtitle, Valid: input.Subtitle != ""},
		Lyrics:   strings.Join(input.Lyrics, "\n\n"),
	}

	res := sdb.db.Create(&dbSong)
	if res.Error != nil {
		return "", errors.New("failed to create")
	}

	return dbSong.UUID.String(), nil
}

func (sdb SQLSongsDB) UpdateSong(id string, input SongInput) error {
	var dbSong DbSong

	res := sdb.db.Where("uuid = ?", id).Take(&dbSong)
	if res.Error != nil {
		return errors.New("not found")
	}

	dbSong.Title = input.Title
	dbSong.Subtitle = sql.NullString{String: input.Subtitle, Valid: input.Subtitle != ""}
	dbSong.Lyrics = strings.Join(input.Lyrics, "\n\n")

	res = sdb.db.Save(&dbSong)
	if res.Error != nil {
		return errors.New("failed to save")
	}

	return nil
}

func (sdb SQLSongsDB) DeleteSong(id string) error {
	var dbSong DbSong

	res := sdb.db.Where("uuid = ?", id).Take(&dbSong)
	if res.Error != nil {
		return errors.New("not found")
	}

	res = sdb.db.Delete(&dbSong)
	if res.Error != nil {
		return errors.New("failed to delete")
	}

	return nil
}

func (sdb SQLSongsDB) Import(n NotionSongsDB) {
	allSongs := n.FilterSongs("")
	fmt.Printf("found %d songs\n", len(allSongs))
	ids := make([]string, 0)

	for _, ns := range allSongs {
		var dbSong DbSong
		result := sdb.db.Where("uuid", ns.Id).Take(&dbSong)
		if result.Error != nil {
			user := DbSong{
				UUID:  uuid.MustParse(ns.Id),
				Title: ns.Title,
				Subtitle: sql.NullString{
					Valid:  ns.Subtitle != "",
					String: ns.Subtitle,
				},
				Number: ns.Number,
				Tags:   ns.Tags,
				Slug:   ns.Slug,
			}

			sdb.db.Create(&user)
			ids = append(ids, ns.Id)
			fmt.Printf("%s %s created\n", ns.Id, ns.Title)
		} else if ns.updatedAt.After(dbSong.UpdatedAt) {
			dbSong.Title = ns.Title
			dbSong.Subtitle = sql.NullString{
				Valid:  ns.Subtitle != "",
				String: ns.Subtitle,
			}
			dbSong.Number = ns.Number
			dbSong.Tags = ns.Tags
			dbSong.Slug = ns.Slug

			sdb.db.Save(&dbSong)
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
			sdb.db.Model(&DbSong{}).Where("uuid = ?", id).Update("lyrics", lyrics)
		}
	}

	fmt.Println("done!")
}
