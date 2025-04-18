package models

import "gorm.io/gorm"

var AllModels = []interface{}{
	&Song{},
	&User{},
	&RefreshToken{},
	&Team{},
	&Invitation{},
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(AllModels...)
}
