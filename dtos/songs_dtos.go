package dtos

import "github.com/hejmsdz/goslides/models"

type SongSummaryResponse struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Subtitle *string `json:"subtitle"`
	Slug     string  `json:"slug"`
	TeamID   *string `json:"teamId"`
}

func NewSongSummaryResponse(song *models.Song) SongSummaryResponse {
	subtitle := &song.Subtitle.String
	if !song.Subtitle.Valid {
		subtitle = nil
	}

	resp := SongSummaryResponse{
		ID:       song.UUID.String(),
		Title:    song.Title,
		Subtitle: subtitle,
		Slug:     song.Slug,
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
	Lyrics    []string `json:"lyrics"`
	CanEdit   bool     `json:"canEdit"`
	CanDelete bool     `json:"canDelete"`
}

func NewSongDetailResponse(song *models.Song, canEdit bool, canDelete bool) SongDetailResponse {
	return SongDetailResponse{
		SongSummaryResponse: NewSongSummaryResponse(song),
		Lyrics:              song.FormatLyrics(models.FormatLyricsOptions{Raw: true}),
		CanEdit:             canEdit,
		CanDelete:           canDelete,
	}
}

type SongRequest struct {
	Title    string   `json:"title"`
	Subtitle string   `json:"subtitle"`
	Lyrics   []string `json:"lyrics"`
	Team     string   `json:"team"`
}

func (r SongRequest) Validate() error {
	return nil
}
