package services

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/hejmsdz/goslides/dtos"
	"github.com/hejmsdz/goslides/models"
	"gorm.io/gorm"
)

type TeamsService struct {
	db                 *gorm.DB
	invitationDuration time.Duration
}

func NewTeamsService(db *gorm.DB) *TeamsService {
	invitationDuration, err := time.ParseDuration(os.Getenv("TEAM_INVITATION_DURATION"))
	if err != nil {
		panic(fmt.Sprintf("failed to read TEAM_INVITATION_DURATION: %s", err.Error()))
	}

	return &TeamsService{db, invitationDuration}
}

func (t *TeamsService) GetUserTeam(user *models.User, uuid string) (*models.Team, error) {
	if user == nil {
		return nil, errors.New("user is nil")
	}

	var team *models.Team
	err := t.db.Joins("INNER JOIN user_teams ON user_teams.team_id = teams.id").
		Where("user_teams.user_id = ?", user.ID).
		Where("uuid = ?", uuid).
		First(&team).Error

	if err != nil {
		return nil, err
	}

	return team, nil
}

func (t *TeamsService) GetUserTeamAllowingEmptyForAdmin(user *models.User, uuid string) (*models.Team, error) {
	if user != nil && user.IsAdmin && uuid == "" {
		return nil, nil
	}

	return t.GetUserTeam(user, uuid)
}

func (t *TeamsService) GetUserTeams(user *models.User) ([]*models.Team, error) {
	var teams []*models.Team
	err := t.db.Joins("INNER JOIN user_teams ON user_teams.team_id = teams.id").
		Where("user_teams.user_id = ?", user.ID).
		Find(&teams).Error

	if err != nil {
		return nil, err
	}

	return teams, nil
}

func (t *TeamsService) CreateTeam(user *models.User, input *dtos.TeamRequest) (*models.Team, error) {
	if user == nil {
		return nil, errors.New("user is nil")
	}

	userTeams, err := t.GetUserTeams(user)
	if err != nil {
		return nil, err
	}

	if len(userTeams) >= 10 {
		return nil, errors.New("user has too many teams")
	}

	team := &models.Team{
		Name:  input.Name,
		Users: []*models.User{user},
	}

	err = t.db.Create(team).Error
	if err != nil {
		return nil, err
	}

	return team, nil
}

func (t *TeamsService) CreateInvitation(user *models.User, uuid string) (string, error) {
	if user == nil {
		return "", errors.New("user is nil")
	}

	team, err := t.GetUserTeam(user, uuid)
	if err != nil {
		return "", err
	}

	invitation := &models.Invitation{
		Team:      team,
		ExpiresAt: time.Now().Add(t.invitationDuration),
	}

	err = t.db.Create(invitation).Error
	if err != nil {
		return "", err
	}

	return invitation.Token, nil
}

func (t *TeamsService) JoinTeam(user *models.User, invitationToken string) (*models.Team, error) {
	if user == nil {
		return nil, errors.New("user is nil")
	}

	var invitation *models.Invitation
	err := t.db.Preload("Team").
		Where("token = ?", invitationToken).
		Where("expires_at > ?", time.Now()).
		Take(&invitation).Error
	if err != nil {
		return nil, err
	}

	team := invitation.Team

	if userTeam, err := t.GetUserTeam(user, team.UUID.String()); err == nil && userTeam != nil {
		return nil, errors.New("user is already in team")
	}

	err = t.db.Model(&team).Association("Users").Append(user)
	if err != nil {
		return nil, err
	}

	return team, nil
}

func (t *TeamsService) LeaveTeam(user *models.User, uuid string) error {
	if user == nil {
		return errors.New("user is nil")
	}

	team, err := t.GetUserTeam(user, uuid)
	if err != nil {
		return err
	}

	err = t.db.Model(&team).Association("Users").Delete(&user)
	if err != nil {
		return err
	}

	return nil
}
