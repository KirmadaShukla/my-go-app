package database

import (
	"fmt"

	"my-go-app/internal/model"

	"gorm.io/gorm"
)

// Models is the ordered list of tables AutoMigrate should manage.
func Models() []any {
	return []any{
		&model.User{},
		&model.TutorSession{},
		&model.TutorMessage{},
		&model.TutorSubjectMemory{},
	}
}

// Migrate applies schema for all application models.
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(Models()...); err != nil {
		return fmt.Errorf("automigrate: %w", err)
	}
	return nil
}
