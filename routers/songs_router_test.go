package routers

/*
import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hejmsdz/goslides/dtos"
	"github.com/hejmsdz/goslides/models"
	"github.com/hejmsdz/goslides/tests"
	"github.com/stretchr/testify/assert"
)

func getSongs(t *testing.T, token string) (*httptest.ResponseRecorder, *[]dtos.SongSummaryResponse, *dtos.ErrorResponse) {
	return tests.Request[[]dtos.SongSummaryResponse](t, testRouter, tests.RequestOptions{
		Method: "GET",
		Path:   "/songs",
		Token:  token,
	})
}

func getSong(t *testing.T, songUUID string, token string) (*httptest.ResponseRecorder, *dtos.SongDetailResponse, *dtos.ErrorResponse) {
	return tests.Request[dtos.SongDetailResponse](t, testRouter, tests.RequestOptions{
		Method: "GET",
		Path:   fmt.Sprintf("/songs/%s", songUUID),
		Token:  token,
	})
}

func postSong(t *testing.T, body *gin.H, token string) (*httptest.ResponseRecorder, *dtos.SongDetailResponse, *dtos.ErrorResponse) {
	return tests.Request[dtos.SongDetailResponse](t, testRouter, tests.RequestOptions{
		Method: "POST",
		Path:   "/songs",
		Body:   body,
		Token:  token,
	})
}

func TestSongsRouter(t *testing.T) {
	db := testContainer.DB
	t.Cleanup(func() { tests.ClearDatabase(db) })

	zebrani := &models.Team{Name: "Zebrani w dnia połowie"}
	roch := &models.Team{Name: "Schola DA św. Rocha"}
	db.Create(zebrani)
	db.Create(roch)

	root := &models.User{Email: "root@admin.com", Teams: []*models.Team{}, IsAdmin: true}
	user1 := &models.User{Email: "user1@ofm.pl", Teams: []*models.Team{zebrani}}
	user2 := &models.User{Email: "user2@swro.ch", Teams: []*models.Team{roch}}
	user3 := &models.User{Email: "user3@both.com", Teams: []*models.Team{roch, zebrani}}
	db.Create(user1)
	db.Create(user2)
	db.Create(user3)

	ubiCaritas, err := testContainer.Songs.CreateSong(dtos.SongRequest{
		Title:  "Ubi caritas",
		Lyrics: []string{"Ubi caritas et amor,\nubi caritas\nDeus ibi est."},
	}, root)
	assert.Nil(t, err)

	panBliskoJest, err := testContainer.Songs.CreateSong(dtos.SongRequest{
		Title:  "Pan blisko jest",
		Lyrics: []string{"Pan blisko jest, oczekuj Go.\nPan blisko jest, w Nim serca moc."},
		Team:   roch.UUID.String(),
	}, user3)
	assert.Nil(t, err)

	modlitwaJezusowa, err := testContainer.Songs.CreateSong(dtos.SongRequest{
		Title:  "Modlitwa Jezusowa",
		Lyrics: []string{"Jezu Chryste, Synu Boga żywego!"},
		Team:   zebrani.UUID.String(),
	}, user3)
	assert.Nil(t, err)

	_, err = testContainer.Songs.CreateSong(dtos.SongRequest{
		Title:  "Pozostań z nami, Panie",
		Lyrics: []string{"Pozostań z nami, Panie,\nbo dzień już się nachylił.\nPozostań z nami, Panie,\nbo zmrok ziemię okrywa."},
		Team:   zebrani.UUID.String(),
	}, user3)
	assert.Nil(t, err)

	t.Run("GET /songs filters songs according to user's team memberships", func(t *testing.T) {
		testCases := []struct {
			user        *models.User
			expectedLen int
		}{
			{nil, 1},
			{user1, 3},
			{user2, 2},
			{user3, 4},
		}

		for _, tc := range testCases {
			userName := "unauthenticated"
			token := ""
			if tc.user != nil {
				userName = tc.user.Email
				token, _ = testContainer.Auth.GenerateAccessToken(tc.user)
			}

			t.Run(fmt.Sprintf("%s should get %d songs", userName, tc.expectedLen), func(t *testing.T) {
				w, resp, errResp := getSongs(t, token)

				assert.Equal(t, 200, w.Code)
				assert.Nil(t, errResp)
				assert.Len(t, *resp, tc.expectedLen)
			})
		}
	})

	t.Run("GET /songs/:id filters songs according to user's team memberships", func(t *testing.T) {
		testCases := []struct {
			song          *models.Song
			user          *models.User
			shouldSucceed bool
		}{
			// public song
			{ubiCaritas, nil, true},
			{ubiCaritas, user1, true},
			{ubiCaritas, user2, true},
			{ubiCaritas, user3, true},

			// roch song
			{panBliskoJest, nil, false},
			{panBliskoJest, user1, false},
			{panBliskoJest, user2, true},
			{panBliskoJest, user3, true},

			// zebrani song
			{modlitwaJezusowa, nil, false},
			{modlitwaJezusowa, user1, true},
			{modlitwaJezusowa, user2, false},
			{modlitwaJezusowa, user3, true},
		}

		for _, tc := range testCases {
			userName := "unauthenticated"
			token := ""
			if tc.user != nil {
				userName = tc.user.Email
				token, _ = testContainer.Auth.GenerateAccessToken(tc.user)
			}
			not := ""
			if !tc.shouldSucceed {
				not = " not"
			}

			t.Run(fmt.Sprintf("%s should%s see the song %s", userName, not, tc.song.Title), func(t *testing.T) {
				w, _, _ := getSong(t, tc.song.UUID.String(), token)

				if tc.shouldSucceed {
					assert.Equal(t, 200, w.Code)
				} else {
					assert.Equal(t, 404, w.Code)
				}
			})
		}
	})

	t.Run("POST /songs allows to create songs only in the user's team", func(t *testing.T) {
		testCases := []struct {
			user          *models.User
			team          *models.Team
			shouldSucceed bool
		}{
			{nil, zebrani, false},
			{nil, roch, false},
			{user1, zebrani, true},
			{user1, roch, false},
			{user2, zebrani, false},
			{user2, roch, true},
			{user3, zebrani, true},
			{user3, roch, true},
		}

		for _, tc := range testCases {
			userName := "unauthenticated"
			token := ""
			if tc.user != nil {
				userName = tc.user.Email
				token, _ = testContainer.Auth.GenerateAccessToken(tc.user)
			}
			not := ""
			if !tc.shouldSucceed {
				not = " not"
			}

			t.Run(fmt.Sprintf("%s should%s be able to create a song in %s", userName, not, tc.team.Name), func(t *testing.T) {
				w, _, _ := postSong(t, &gin.H{
					"title":  "Dummy song",
					"lyrics": []string{"Lorem ipsum dolor sit amet", "Consectetur adipiscit elit"},
					"team":   tc.team.UUID,
				}, token)

				if tc.user == nil {
					assert.Equal(t, 401, w.Code)
				} else if tc.shouldSucceed {
					assert.Equal(t, 201, w.Code)
				} else {
					assert.Equal(t, 403, w.Code)
				}
			})
		}
	})
}
*/
