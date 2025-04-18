package services_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hejmsdz/goslides/dtos"
	"github.com/hejmsdz/goslides/models"
	"github.com/hejmsdz/goslides/tests"
	"github.com/stretchr/testify/assert"
)

func TestCreateTeam(t *testing.T) {
	te := tests.NewTestEnvironment(t)

	te.Run("successful team creation", func(t *testing.T, tce *tests.TestCaseEnvironment) {
		user := &models.User{
			Email:       "test@example.com",
			DisplayName: "Test User",
		}
		err := tce.DB.Create(user).Error
		assert.NoError(t, err)

		input := &dtos.TeamRequest{
			Name: "Test Team",
		}

		team, err := tce.Container.Teams.CreateTeam(user, input)
		assert.NoError(t, err)
		assert.NotNil(t, team)
		assert.Equal(t, "Test Team", team.Name)

		userTeams, err := tce.Container.Teams.GetUserTeam(user, team.UUID.String())
		assert.NoError(t, err)
		assert.Equal(t, team.ID, userTeams.ID)
	})

	te.Run("nil user returns error", func(t *testing.T, tce *tests.TestCaseEnvironment) {
		input := &dtos.TeamRequest{
			Name: "Test Team",
		}

		team, err := tce.Container.Teams.CreateTeam(nil, input)
		assert.Error(t, err)
		assert.Nil(t, team)
		assert.Equal(t, "user is nil", err.Error())
	})

	te.Run("user with too many teams returns error", func(t *testing.T, tce *tests.TestCaseEnvironment) {
		// Create a test user
		user := &models.User{
			Email:       "test@example.com",
			DisplayName: "Test User",
		}
		err := tce.DB.Create(user).Error
		assert.NoError(t, err)

		// Create 10 teams for the user
		for i := 0; i < 10; i++ {
			team := &models.Team{
				Name:  fmt.Sprintf("Team %d", i+1),
				Users: []*models.User{user},
			}
			err := tce.DB.Create(team).Error
			assert.NoError(t, err)
		}

		userTeams, err := tce.Container.Teams.GetUserTeams(user)
		assert.NoError(t, err)
		assert.Equal(t, 10, len(userTeams))

		input := &dtos.TeamRequest{
			Name: "Extra Team",
		}

		team, err := tce.Container.Teams.CreateTeam(user, input)
		assert.Error(t, err)
		assert.Nil(t, team)
		assert.Equal(t, "user has too many teams", err.Error())
	})
}

func TestLeaveTeam(t *testing.T) {
	te := tests.NewTestEnvironment(t)

	te.Run("successful team leaving", func(t *testing.T, tce *tests.TestCaseEnvironment) {
		// Create a test user
		user := &models.User{
			Email:       "test@example.com",
			DisplayName: "Test User",
		}
		err := tce.DB.Create(user).Error
		assert.NoError(t, err)

		// Create a team with the user as member
		team := &models.Team{
			Name:  "Test Team",
			Users: []*models.User{user},
		}
		err = tce.DB.Create(team).Error
		assert.NoError(t, err)

		// Verify user is in team
		userTeams, err := tce.Container.Teams.GetUserTeams(user)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(userTeams))

		// Leave the team
		err = tce.Container.Teams.LeaveTeam(user, team.UUID.String())
		assert.NoError(t, err)

		// Verify user is no longer in team
		userTeams, err = tce.Container.Teams.GetUserTeams(user)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(userTeams))
	})

	te.Run("nil user returns error", func(t *testing.T, tce *tests.TestCaseEnvironment) {
		err := tce.Container.Teams.LeaveTeam(nil, "some-uuid")
		assert.Error(t, err)
		assert.Equal(t, "user is nil", err.Error())
	})

	te.Run("non-existent team returns error", func(t *testing.T, tce *tests.TestCaseEnvironment) {
		// Create a test user
		user := &models.User{
			Email:       "test@example.com",
			DisplayName: "Test User",
		}
		err := tce.DB.Create(user).Error
		assert.NoError(t, err)

		// Try to leave a non-existent team
		err = tce.Container.Teams.LeaveTeam(user, "non-existent-uuid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "record not found")
	})

	te.Run("user not in team returns error", func(t *testing.T, tce *tests.TestCaseEnvironment) {
		// Create two test users
		user1 := &models.User{
			Email:       "user1@example.com",
			DisplayName: "User 1",
		}
		user2 := &models.User{
			Email:       "user2@example.com",
			DisplayName: "User 2",
		}
		err := tce.DB.Create(user1).Error
		assert.NoError(t, err)
		err = tce.DB.Create(user2).Error
		assert.NoError(t, err)

		// Create a team with only user1 as member
		team := &models.Team{
			Name:  "Test Team",
			Users: []*models.User{user1},
		}
		err = tce.DB.Create(team).Error
		assert.NoError(t, err)

		// Try to have user2 leave the team
		err = tce.Container.Teams.LeaveTeam(user2, team.UUID.String())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "record not found")
	})
}

func TestCreateInvitation(t *testing.T) {
	te := tests.NewTestEnvironment(t)

	te.Run("successful invitation creation", func(t *testing.T, tce *tests.TestCaseEnvironment) {
		// Create a test user
		user := &models.User{
			Email:       "test@example.com",
			DisplayName: "Test User",
		}
		err := tce.DB.Create(user).Error
		assert.NoError(t, err)

		// Create a team with the user as member
		team := &models.Team{
			Name:  "Test Team",
			Users: []*models.User{user},
		}
		err = tce.DB.Create(team).Error
		assert.NoError(t, err)

		// Create an invitation
		token, err := tce.Container.Teams.CreateInvitation(user, team.UUID.String())
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Verify invitation was created
		var invitation models.Invitation
		err = tce.DB.Where("token = ?", token).First(&invitation).Error
		assert.NoError(t, err)
		assert.Equal(t, team.ID, invitation.TeamID)
	})

	te.Run("nil user returns error", func(t *testing.T, tce *tests.TestCaseEnvironment) {
		token, err := tce.Container.Teams.CreateInvitation(nil, "some-uuid")
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Equal(t, "user is nil", err.Error())
	})

	te.Run("non-existent team returns error", func(t *testing.T, tce *tests.TestCaseEnvironment) {
		// Create a test user
		user := &models.User{
			Email:       "test@example.com",
			DisplayName: "Test User",
		}
		err := tce.DB.Create(user).Error
		assert.NoError(t, err)

		// Try to create invitation for non-existent team
		token, err := tce.Container.Teams.CreateInvitation(user, "non-existent-uuid")
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Contains(t, err.Error(), "record not found")
	})

	te.Run("user not in team returns error", func(t *testing.T, tce *tests.TestCaseEnvironment) {
		// Create two test users
		user1 := &models.User{
			Email:       "user1@example.com",
			DisplayName: "User 1",
		}
		user2 := &models.User{
			Email:       "user2@example.com",
			DisplayName: "User 2",
		}
		err := tce.DB.Create(user1).Error
		assert.NoError(t, err)
		err = tce.DB.Create(user2).Error
		assert.NoError(t, err)

		// Create a team with only user1 as member
		team := &models.Team{
			Name:  "Test Team",
			Users: []*models.User{user1},
		}
		err = tce.DB.Create(team).Error
		assert.NoError(t, err)

		// Try to have user2 create an invitation
		token, err := tce.Container.Teams.CreateInvitation(user2, team.UUID.String())
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Contains(t, err.Error(), "record not found")
	})
}

func TestJoinTeam(t *testing.T) {
	te := tests.NewTestEnvironment(t)

	te.Run("successful team joining", func(t *testing.T, tce *tests.TestCaseEnvironment) {
		// Create two test users
		user1 := &models.User{
			Email:       "user1@example.com",
			DisplayName: "User 1",
		}
		user2 := &models.User{
			Email:       "user2@example.com",
			DisplayName: "User 2",
		}
		err := tce.DB.Create(user1).Error
		assert.NoError(t, err)
		err = tce.DB.Create(user2).Error
		assert.NoError(t, err)

		// Create a team with user1 as member
		team := &models.Team{
			Name:  "Test Team",
			Users: []*models.User{user1},
		}
		err = tce.DB.Create(team).Error
		assert.NoError(t, err)

		// Create an invitation
		token, err := tce.Container.Teams.CreateInvitation(user1, team.UUID.String())
		assert.NoError(t, err)

		// Have user2 join the team using the invitation
		joinedTeam, err := tce.Container.Teams.JoinTeam(user2, token)
		assert.NoError(t, err)
		assert.NotNil(t, joinedTeam)
		assert.Equal(t, team.ID, joinedTeam.ID)

		// Verify user2 is now in the team
		userTeams, err := tce.Container.Teams.GetUserTeams(user2)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(userTeams))
		assert.Equal(t, team.ID, userTeams[0].ID)
	})

	te.Run("nil user returns error", func(t *testing.T, tce *tests.TestCaseEnvironment) {
		team, err := tce.Container.Teams.JoinTeam(nil, "some-token")
		assert.Error(t, err)
		assert.Nil(t, team)
		assert.Equal(t, "user is nil", err.Error())
	})

	te.Run("invalid token returns error", func(t *testing.T, tce *tests.TestCaseEnvironment) {
		// Create a test user
		user := &models.User{
			Email:       "test@example.com",
			DisplayName: "Test User",
		}
		err := tce.DB.Create(user).Error
		assert.NoError(t, err)

		// Try to join with invalid token
		team, err := tce.Container.Teams.JoinTeam(user, "invalid-token")
		assert.Error(t, err)
		assert.Nil(t, team)
		assert.Contains(t, err.Error(), "record not found")
	})

	te.Run("expired invitation returns error", func(t *testing.T, tce *tests.TestCaseEnvironment) {
		// Create two test users
		user1 := &models.User{
			Email:       "user1@example.com",
			DisplayName: "User 1",
		}
		user2 := &models.User{
			Email:       "user2@example.com",
			DisplayName: "User 2",
		}
		err := tce.DB.Create(user1).Error
		assert.NoError(t, err)
		err = tce.DB.Create(user2).Error
		assert.NoError(t, err)

		// Create a team with user1 as member
		team := &models.Team{
			Name:  "Test Team",
			Users: []*models.User{user1},
		}
		err = tce.DB.Create(team).Error
		assert.NoError(t, err)

		// Create an invitation and manually set it as expired
		invitation := &models.Invitation{
			Team:      team,
			ExpiresAt: time.Now().Add(-1 * time.Hour), // Set to 1 hour in the past
		}
		err = tce.DB.Create(invitation).Error
		assert.NoError(t, err)

		// Try to join with expired invitation
		joinedTeam, err := tce.Container.Teams.JoinTeam(user2, invitation.Token)
		assert.Error(t, err)
		assert.Nil(t, joinedTeam)
		assert.Contains(t, err.Error(), "record not found")
	})

	te.Run("user already in team returns error", func(t *testing.T, tce *tests.TestCaseEnvironment) {
		// Create a test user
		user := &models.User{
			Email:       "test@example.com",
			DisplayName: "Test User",
		}
		err := tce.DB.Create(user).Error
		assert.NoError(t, err)

		// Create a team with the user as member
		team := &models.Team{
			Name:  "Test Team",
			Users: []*models.User{user},
		}
		err = tce.DB.Create(team).Error
		assert.NoError(t, err)

		// Create an invitation
		token, err := tce.Container.Teams.CreateInvitation(user, team.UUID.String())
		assert.NoError(t, err)

		// Try to join the team again
		joinedTeam, err := tce.Container.Teams.JoinTeam(user, token)
		assert.Error(t, err)
		assert.Nil(t, joinedTeam)
		assert.Equal(t, "user is already in team", err.Error())
	})
}
