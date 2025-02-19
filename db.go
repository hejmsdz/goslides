package main

import (
	"database/sql"
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DbSong struct {
	gorm.Model
	UUID     uuid.UUID
	Title    string
	Subtitle sql.NullString
	Number   string
	Tags     string
	Slug     string
	Lyrics   string
}

type SQLSongsDB struct {
	db *gorm.DB
}

func (sdb *SQLSongsDB) Initialize() {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
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
	querySlug := Slugify(query)

	var dbSongs []DbSong
	sdb.db.Select("uuid", "title", "subtitle", "number", "tags", "slug").
		Where("slug LIKE ?", "%"+querySlug+"%").
		Find(&dbSongs)

	songs := make([]Song, len(dbSongs))

	for i, s := range dbSongs {
		songs[i].Id = s.UUID.String()
		songs[i].Title = s.Title
		if s.Subtitle.Valid {
			songs[i].Subtitle = s.Subtitle.String
		}
		songs[i].Number = s.Number
		songs[i].Slug = s.Slug
		songs[i].Tags = s.Tags
		songs[i].IsOrdinary = strings.Contains(s.Tags, ordinaryTag)
	}

	return songs
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
