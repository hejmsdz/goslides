package services

import (
	"errors"

	"github.com/hejmsdz/goslides/models"
	"gorm.io/gorm"
)

type TeamsService struct {
	db *gorm.DB
}

func NewTeamsService(db *gorm.DB) *TeamsService {
	return &TeamsService{db}
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
