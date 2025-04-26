package dtos

import (
	"errors"

	"github.com/hejmsdz/goslides/models"
)

type SongSummaryResponse struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	Subtitle   *string `json:"subtitle"`
	Slug       string  `json:"slug"`
	TeamID     *string `json:"teamId"`
	IsOverride bool    `json:"isOverride"`
}

func NewSongSummaryResponse(song *models.Song) SongSummaryResponse {
	subtitle := &song.Subtitle.String
	if !song.Subtitle.Valid {
		subtitle = nil
	}

	resp := SongSummaryResponse{
		ID:         song.UUID.String(),
		Title:      song.Title,
		Subtitle:   subtitle,
		Slug:       song.Slug,
		IsOverride: song.OverriddenSongID != nil,
	}

	if song.Team != nil {
		teamID := song.Team.UUID.String()
		resp.TeamID = &teamID
	}

	return resp
}

func NewSongListResponse(songs []models.Song) []SongSummaryResponse {
	resp := make([]SongSummaryResponse, len(songs))

	for i, song := range songs {
		resp[i] = NewSongSummaryResponse(&song)
	}

	return resp
}

type SongDetailResponse struct {
	SongSummaryResponse
	OverriddenSongID *string  `json:"overriddenSongId"`
	Lyrics           []string `json:"lyrics"`
	CanEdit          bool     `json:"canEdit"`
	CanDelete        bool     `json:"canDelete"`
	CanOverride      bool     `json:"canOverride"`
}

func NewSongDetailResponse(song *models.Song, canEdit bool, canDelete bool, canOverride bool) SongDetailResponse {
	var overriddenSongID *string
	if song.OverriddenSong != nil {
		songID := song.OverriddenSong.UUID.String()
		overriddenSongID = &songID
	}

	return SongDetailResponse{
		SongSummaryResponse: NewSongSummaryResponse(song),
		OverriddenSongID:    overriddenSongID,
		Lyrics:              song.FormatLyrics(models.FormatLyricsOptions{Raw: true}),
		CanEdit:             canEdit,
		CanDelete:           canDelete,
		CanOverride:         canOverride,
	}
}

type SongRequest struct {
	Title      string   `json:"title"`
	Subtitle   string   `json:"subtitle"`
	Lyrics     []string `json:"lyrics"`
	TeamID     string   `json:"teamId"`
	IsOverride bool     `json:"isOverride"`
}

func (r SongRequest) Validate() error {
	if r.IsOverride {
		if r.TeamID == "" {
			return errors.New("teamId is required when overriding a song")
		}
	}
	return nil
}
