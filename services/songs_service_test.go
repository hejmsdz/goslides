package services_test

import (
	"database/sql"
	"testing"

	"github.com/hejmsdz/goslides/models"
	"github.com/hejmsdz/goslides/tests"
	"github.com/stretchr/testify/assert"
)

type TestData struct {
	User  *models.User
	Team  *models.Team
	Songs []*models.Song
}

func createTestData(t *testing.T, tce *tests.TestCaseEnvironment, canAccessUnofficial bool) *TestData {
	// Create a user
	user := &models.User{
		Email:       "test@example.com",
		DisplayName: "Test User",
	}
	err := tce.DB.Create(user).Error
	assert.NoError(t, err)

	// Create a team
	team := &models.Team{
		Name:                     "Test Team",
		CreatedByID:              user.ID,
		CanAccessUnofficialSongs: canAccessUnofficial,
		Users:                    []*models.User{user},
	}
	err = tce.DB.Create(team).Error
	assert.NoError(t, err)

	// Create official songs
	officialSong1 := &models.Song{
		Title:        "Official Song 1",
		Subtitle:     sql.NullString{},
		Lyrics:       "Verse 1\n\nVerse 2",
		IsUnofficial: false,
		CreatedByID:  user.ID,
		UpdatedByID:  user.ID,
	}
	err = tce.DB.Create(officialSong1).Error
	assert.NoError(t, err)

	officialSong2 := &models.Song{
		Title:        "Official Song 2",
		Subtitle:     sql.NullString{},
		Lyrics:       "Verse 1\n\nVerse 2",
		IsUnofficial: false,
		CreatedByID:  user.ID,
		UpdatedByID:  user.ID,
	}
	err = tce.DB.Create(officialSong2).Error
	assert.NoError(t, err)

	// Create unofficial songs
	unofficialSong1 := &models.Song{
		Title:        "Unofficial Song 1",
		Subtitle:     sql.NullString{},
		Lyrics:       "Verse 1\n\nVerse 2",
		IsUnofficial: true,
		CreatedByID:  user.ID,
		UpdatedByID:  user.ID,
	}
	err = tce.DB.Create(unofficialSong1).Error
	assert.NoError(t, err)

	unofficialSong2 := &models.Song{
		Title:        "Unofficial Song 2",
		Subtitle:     sql.NullString{},
		Lyrics:       "Verse 1\n\nVerse 2",
		IsUnofficial: true,
		CreatedByID:  user.ID,
		UpdatedByID:  user.ID,
	}
	err = tce.DB.Create(unofficialSong2).Error
	assert.NoError(t, err)

	return &TestData{
		User:  user,
		Team:  team,
		Songs: []*models.Song{officialSong1, officialSong2, unofficialSong1, unofficialSong2},
	}
}

func TestFilterSongsUnofficialFiltering(t *testing.T) {
	te := tests.NewTestEnvironment(t)

	te.Run("filters out unofficial songs when team cannot access them", func(t *testing.T, tce *tests.TestCaseEnvironment) {
		testData := createTestData(t, tce, false)

		// Filter songs - should only return official songs
		songs, err := tce.Container.Songs.FilterSongs("", testData.User, testData.Team.UUID.String())
		assert.NoError(t, err)
		assert.Len(t, songs, 2)

		// Verify only official songs are returned
		songTitles := make([]string, len(songs))
		for i, song := range songs {
			songTitles[i] = song.Title
		}
		assert.Contains(t, songTitles, "Official Song 1")
		assert.Contains(t, songTitles, "Official Song 2")
		assert.NotContains(t, songTitles, "Unofficial Song 1")
		assert.NotContains(t, songTitles, "Unofficial Song 2")
	})

	te.Run("includes unofficial songs when team can access them", func(t *testing.T, tce *tests.TestCaseEnvironment) {
		testData := createTestData(t, tce, true)

		// Filter songs - should return both official and unofficial songs
		songs, err := tce.Container.Songs.FilterSongs("", testData.User, testData.Team.UUID.String())
		assert.NoError(t, err)
		assert.Len(t, songs, 4)

		// Verify both official and unofficial songs are returned
		songTitles := make([]string, len(songs))
		for i, song := range songs {
			songTitles[i] = song.Title
		}
		assert.Contains(t, songTitles, "Official Song 1")
		assert.Contains(t, songTitles, "Official Song 2")
		assert.Contains(t, songTitles, "Unofficial Song 1")
		assert.Contains(t, songTitles, "Unofficial Song 2")
	})

	te.Run("filters out unofficial songs when no team is specified", func(t *testing.T, tce *tests.TestCaseEnvironment) {
		testData := createTestData(t, tce, false)

		// Filter songs without team - should only return official songs
		songs, err := tce.Container.Songs.FilterSongs("", testData.User, "")
		assert.NoError(t, err)
		assert.Len(t, songs, 2)

		// Verify only official songs are returned
		songTitles := make([]string, len(songs))
		for i, song := range songs {
			songTitles[i] = song.Title
		}
		assert.Contains(t, songTitles, "Official Song 1")
		assert.Contains(t, songTitles, "Official Song 2")
		assert.NotContains(t, songTitles, "Unofficial Song 1")
		assert.NotContains(t, songTitles, "Unofficial Song 2")
	})

	te.Run("does not return unofficial songs from GetSong when user cannot access them", func(t *testing.T, tce *tests.TestCaseEnvironment) {
		testData := createTestData(t, tce, false)

		for _, song := range testData.Songs {
			retrievedSong, err := tce.Container.Songs.GetSong(song.UUID.String(), testData.User)
			if song.IsUnofficial {
				assert.Error(t, err)
				assert.Nil(t, retrievedSong)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, song.Title, retrievedSong.Title)
			}
		}
	})

	te.Run("returns unofficial songs from GetSong when user can access them", func(t *testing.T, tce *tests.TestCaseEnvironment) {
		testData := createTestData(t, tce, true)

		for _, song := range testData.Songs {
			retrievedSong, err := tce.Container.Songs.GetSong(song.UUID.String(), testData.User)
			assert.NoError(t, err)
			assert.Equal(t, song.Title, retrievedSong.Title)
		}
	})
}
