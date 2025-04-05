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

func (s *TeamsService) GetTeam(uuid string) (*models.Team, error) {
	if uuid == "" {
		return nil, nil
	}

	var team models.Team
	err := s.db.Where("uuid = ?", uuid).First(&team).Error
	if err != nil {
		return nil, err
	}

	return &team, nil
}
