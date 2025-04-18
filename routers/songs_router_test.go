package routers_test

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hejmsdz/goslides/dtos"
	"github.com/hejmsdz/goslides/models"
	"github.com/hejmsdz/goslides/tests"
	"github.com/stretchr/testify/assert"
)

func getSongs(t *testing.T, router *gin.Engine, token string, queryParams ...string) (*httptest.ResponseRecorder, *[]dtos.SongSummaryResponse, *dtos.ErrorResponse) {
	query := ""
	if len(queryParams) > 0 {
		query = "?" + strings.Join(queryParams, "&")
	}

	return tests.Request[[]dtos.SongSummaryResponse](t, router, tests.RequestOptions{
		Method: "GET",
		Path:   "/v2/songs" + query,
		Token:  token,
	})
}

func getSong(t *testing.T, router *gin.Engine, songUUID string, token string) (*httptest.ResponseRecorder, *dtos.SongDetailResponse, *dtos.ErrorResponse) {
	return tests.Request[dtos.SongDetailResponse](t, router, tests.RequestOptions{
		Method: "GET",
		Path:   fmt.Sprintf("/v2/songs/%s", songUUID),
		Token:  token,
	})
}

func postSong(t *testing.T, router *gin.Engine, body *gin.H, token string) (*httptest.ResponseRecorder, *dtos.SongDetailResponse, *dtos.ErrorResponse) {
	return tests.Request[dtos.SongDetailResponse](t, router, tests.RequestOptions{
		Method: "POST",
		Path:   "/v2/songs",
		Body:   body,
		Token:  token,
	})
}

func patchSong(t *testing.T, router *gin.Engine, songUUID string, body *gin.H, token string) (*httptest.ResponseRecorder, *dtos.SongDetailResponse, *dtos.ErrorResponse) {
	return tests.Request[dtos.SongDetailResponse](t, router, tests.RequestOptions{
		Method: "PATCH",
		Path:   fmt.Sprintf("/v2/songs/%s", songUUID),
		Body:   body,
		Token:  token,
	})
}

type TestData struct {
	teams map[string]*models.Team
	users map[string]*models.User
	songs map[string]*models.Song
}

func createTestData(t *testing.T, tce *tests.TestCaseEnvironment) TestData {
	db := tce.DB

	zebrani := &models.Team{Name: "Zebrani w dnia połowie"}
	roch := &models.Team{Name: "Schola DA św. Rocha"}
	db.Create(zebrani)
	db.Create(roch)

	root := &models.User{Email: "root@admin.com", Teams: []*models.Team{}, IsAdmin: true}
	user1 := &models.User{Email: "user1@ofm.pl", Teams: []*models.Team{zebrani}}
	user2 := &models.User{Email: "user2@swro.ch", Teams: []*models.Team{roch}}
	user3 := &models.User{Email: "user3@archpoznan.pl", Teams: []*models.Team{roch, zebrani}}
	db.Create(root)
	db.Create(user1)
	db.Create(user2)
	db.Create(user3)

	ubiCaritas, err := tce.Container.Songs.CreateSong(dtos.SongRequest{
		Title:  "Ubi caritas",
		Lyrics: []string{"Ubi caritas et amor,\nubi caritas\nDeus ibi est."},
	}, root)
	assert.Nil(t, err)

	panBliskoJest, err := tce.Container.Songs.CreateSong(dtos.SongRequest{
		Title:  "Pan blisko jest",
		Lyrics: []string{"Pan blisko jest, oczekuj Go.\nPan blisko jest, w Nim serca moc."},
		TeamID: roch.UUID.String(),
	}, user3)
	assert.Nil(t, err)

	modlitwaJezusowa, err := tce.Container.Songs.CreateSong(dtos.SongRequest{
		Title:  "Modlitwa Jezusowa",
		Lyrics: []string{"Jezu Chryste, Synu Boga żywego!"},
		TeamID: zebrani.UUID.String(),
	}, user3)
	assert.Nil(t, err)

	pozostanZNami, err := tce.Container.Songs.CreateSong(dtos.SongRequest{
		Title:  "Pozostań z nami, Panie",
		Lyrics: []string{"Pozostań z nami, Panie,\nbo dzień już się nachylił.\nPozostań z nami, Panie,\nbo zmrok ziemię okrywa."},
		TeamID: zebrani.UUID.String(),
	}, user3)
	assert.Nil(t, err)

	return TestData{
		teams: map[string]*models.Team{
			"zebrani": zebrani,
			"roch":    roch,
		},
		users: map[string]*models.User{
			"root":  root,
			"user1": user1,
			"user2": user2,
			"user3": user3,
		},
		songs: map[string]*models.Song{
			"ubiCaritas":       ubiCaritas,
			"panBliskoJest":    panBliskoJest,
			"modlitwaJezusowa": modlitwaJezusowa,
			"pozostanZNami":    pozostanZNami,
		},
	}
}

func TestSongsRouter(t *testing.T) {
	te := tests.NewTestEnvironment(t)

	te.Run("GET /songs filters songs according to user's team memberships", func(t *testing.T, tce *tests.TestCaseEnvironment) {
		testData := createTestData(t, tce)

		testCases := []struct {
			user        *models.User
			expectedLen int
		}{
			{nil, 1},
			{testData.users["user1"], 3},
			{testData.users["user2"], 2},
			{testData.users["user3"], 4},
		}

		for _, tc := range testCases {
			userName := "unauthenticated"
			token := ""
			if tc.user != nil {
				userName = tc.user.Email
				token, _ = tce.Container.Auth.GenerateAccessToken(tc.user)
			}

			t.Run(fmt.Sprintf("%s should get %d songs", userName, tc.expectedLen), func(t *testing.T) {
				w, resp, errResp := getSongs(t, tce.App, token)

				assert.Equal(t, 200, w.Code)
				assert.Nil(t, errResp)
				assert.Len(t, *resp, tc.expectedLen)
			})
		}
	})

	te.Run("GET /songs allows to filter by team with a query param teamId and verifies user membership in this team", func(t *testing.T, tce *tests.TestCaseEnvironment) {
		testData := createTestData(t, tce)

		roch := testData.teams["roch"].UUID.String()
		zebrani := testData.teams["zebrani"].UUID.String()

		testCases := []struct {
			user        *models.User
			teamId      string
			expectedLen int
		}{
			{nil, "", 1},
			{nil, roch, 0},
			{testData.users["user1"], zebrani, 3},
			{testData.users["user1"], roch, 0},
			{testData.users["user2"], roch, 2},
			{testData.users["user2"], zebrani, 0},
			{testData.users["user3"], roch, 2},
			{testData.users["user3"], zebrani, 3},
		}

		for _, tc := range testCases {
			userName := "unauthenticated"
			token := ""
			if tc.user != nil {
				userName = tc.user.Email
				token, _ = tce.Container.Auth.GenerateAccessToken(tc.user)
			}

			t.Run(fmt.Sprintf("%s filtering by team %s", userName, tc.teamId), func(t *testing.T) {
				_, resp, errResp := getSongs(t, tce.App, token, "teamId="+tc.teamId)

				assert.Nil(t, errResp)
				assert.Len(t, *resp, tc.expectedLen)
			})
		}
	})

	te.Run("GET /songs/:id filters songs according to user's team memberships", func(t *testing.T, tce *tests.TestCaseEnvironment) {
		testData := createTestData(t, tce)

		testCases := []struct {
			song          *models.Song
			user          *models.User
			shouldSucceed bool
		}{
			// public song
			{testData.songs["ubiCaritas"], nil, true},
			{testData.songs["ubiCaritas"], testData.users["user1"], true},
			{testData.songs["ubiCaritas"], testData.users["user2"], true},
			{testData.songs["ubiCaritas"], testData.users["user3"], true},

			// roch song
			{testData.songs["panBliskoJest"], nil, false},
			{testData.songs["panBliskoJest"], testData.users["user1"], false},
			{testData.songs["panBliskoJest"], testData.users["user2"], true},
			{testData.songs["panBliskoJest"], testData.users["user3"], true},

			// zebrani song
			{testData.songs["modlitwaJezusowa"], nil, false},
			{testData.songs["modlitwaJezusowa"], testData.users["user1"], true},
			{testData.songs["modlitwaJezusowa"], testData.users["user2"], false},
			{testData.songs["modlitwaJezusowa"], testData.users["user3"], true},
		}

		for _, tc := range testCases {
			userName := "unauthenticated"
			token := ""
			if tc.user != nil {
				userName = tc.user.Email
				token, _ = tce.Container.Auth.GenerateAccessToken(tc.user)
			}
			not := ""
			if !tc.shouldSucceed {
				not = " not"
			}

			t.Run(fmt.Sprintf("%s should%s see the song %s", userName, not, tc.song.Title), func(t *testing.T) {
				w, _, _ := getSong(t, tce.App, tc.song.UUID.String(), token)

				if tc.shouldSucceed {
					assert.Equal(t, 200, w.Code)
				} else {
					assert.Equal(t, 404, w.Code)
				}
			})
		}
	})

	te.Run("POST /songs allows to create songs only in the user's team", func(t *testing.T, tce *tests.TestCaseEnvironment) {
		testData := createTestData(t, tce)

		testCases := []struct {
			user          *models.User
			team          *models.Team
			shouldSucceed bool
		}{
			{nil, testData.teams["zebrani"], false},
			{nil, testData.teams["roch"], false},
			{testData.users["user1"], testData.teams["zebrani"], true},
			{testData.users["user1"], testData.teams["roch"], false},
			{testData.users["user2"], testData.teams["zebrani"], false},
			{testData.users["user2"], testData.teams["roch"], true},
			{testData.users["user3"], testData.teams["zebrani"], true},
			{testData.users["user3"], testData.teams["roch"], true},
		}

		for _, tc := range testCases {
			userName := "unauthenticated"
			token := ""
			if tc.user != nil {
				userName = tc.user.Email
				token, _ = tce.Container.Auth.GenerateAccessToken(tc.user)
			}
			not := ""
			if !tc.shouldSucceed {
				not = " not"
			}

			t.Run(fmt.Sprintf("%s should%s be able to create a song in %s", userName, not, tc.team.Name), func(t *testing.T) {
				w, _, _ := postSong(t, tce.App, &gin.H{
					"title":  "Dummy song",
					"lyrics": []string{"Lorem ipsum dolor sit amet", "Consectetur adipiscit elit"},
					"teamId": tc.team.UUID,
				}, token)

				if tc.user == nil {
					assert.Equal(t, 401, w.Code)
				} else if tc.shouldSucceed {
					assert.Equal(t, 201, w.Code)
				} else {
					assert.Equal(t, 404, w.Code)
				}
			})
		}
	})

	te.Run("PATCH /songs/:id allows to override songs", func(t *testing.T, tce *tests.TestCaseEnvironment) {
		testData := createTestData(t, tce)

		token, err := tce.Container.Auth.GenerateAccessToken(testData.users["user2"])
		assert.NoError(t, err)

		w, resp, _ := patchSong(t, tce.App, testData.songs["ubiCaritas"].UUID.String(), &gin.H{
			"title":      "Ubi caritas (customized)",
			"lyrics":     []string{"Ubi caritas et amor, Deus ibi est.\nCongregavit nos in unum Christi amor."},
			"teamId":     testData.teams["roch"].UUID,
			"isOverride": true,
		}, token)

		assert.Equal(t, 200, w.Code)
		songs := tce.Container.Songs.FilterSongs("ubi", testData.users["user2"], "")
		assert.Len(t, songs, 1)
		assert.Equal(t, "Ubi caritas (customized)", songs[0].Title)
		assert.Equal(t, "Ubi caritas (customized)", resp.Title)
	})
}
