package services

import (
	"errors"

	"github.com/hejmsdz/goslides/models"
	"gorm.io/gorm"
)

type UsersService struct {
	db *gorm.DB
}

func NewUsersService(db *gorm.DB) *UsersService {
	return &UsersService{db}
}

func (s UsersService) GetUser(uuid string) (*models.User, error) {
	var user *models.User
	result := s.db.Where("uuid", uuid).Take(&user)

	if result.Error != nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (s UsersService) GetUserByEmail(email string) (*models.User, error) {
	var user *models.User
	result := s.db.Where("email", email).Take(&user)

	if result.Error != nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}
