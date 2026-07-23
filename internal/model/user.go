package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User is the Sequelize-style model for auth.
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Email        string    `gorm:"size:255;uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
