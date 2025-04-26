package services

import (
	"errors"
	"time"

	"github.com/hejmsdz/goslides/dtos"
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

func (s UsersService) UpdateUser(user *models.User, input dtos.UserUpdateRequest) error {
	user.DisplayName = input.DisplayName

	return s.db.Save(user).Error
}

func (s UsersService) DeleteUser(user *models.User) error {
	user.DisplayName = ""
	user.Email = user.UUID.String() + "@deleted"
	user.Teams = []*models.Team{}
	user.IsAdmin = false
	user.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}

	err := s.db.Save(user).Error
	if err != nil {
		return err
	}

	err = s.db.Where("user_id = ?", user.ID).Delete(&models.RefreshToken{}).Error
	if err != nil {
		return err
	}

	return nil
}
