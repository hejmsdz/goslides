package dtos

import "github.com/hejmsdz/goslides/models"

type SongSummaryResponse struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Subtitle *string `json:"subtitle"`
	Slug     string  `json:"slug"`
}

func NewSongSummaryResponse(song *models.Song) SongSummaryResponse {
	subtitle := &song.Subtitle.String
	if !song.Subtitle.Valid {
		subtitle = nil
	}

	return SongSummaryResponse{
		ID:       song.UUID.String(),
		Title:    song.Title,
		Subtitle: subtitle,
		Slug:     song.Slug,
	}
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
	Lyrics []string `json:"lyrics"`
}

func NewSongDetailResponse(song *models.Song) SongDetailResponse {
	return SongDetailResponse{
		SongSummaryResponse: NewSongSummaryResponse(song),
		Lyrics:              song.FormatLyrics(models.FormatLyricsOptions{Raw: true}),
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
