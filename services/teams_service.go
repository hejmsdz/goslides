package services

import (
	"github.com/hejmsdz/goslides/models"
	"gorm.io/gorm"
)

type TeamsService struct {
	db *gorm.DB
}

func NewTeamsService(db *gorm.DB) *TeamsService {
	return &TeamsService{db}
}

func (t *TeamsService) GetTeam(uuid string) (*models.Team, error) {
	if uuid == "" {
		return nil, nil
	}

	var team models.Team
	err := t.db.Where("uuid = ?", uuid).First(&team).Error
	if err != nil {
		return nil, err
	}

	return &team, nil
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
