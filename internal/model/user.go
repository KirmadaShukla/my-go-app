package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User is the Sequelize-style model for auth.
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name         string    `gorm:"size:255;not null" json:"name"`
	Email        string    `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Gender       string    `gorm:"size:20;not null" json:"gender"`
	MotherName   string    `gorm:"size:255;not null" json:"mother_name"`
	FatherName   string    `gorm:"size:255;not null" json:"father_name"`
	MobileNumber string    `gorm:"size:20;not null" json:"mobile_number"`
	ChildAge     int       `gorm:"not null" json:"child_age"`
	ChildClass   string    `gorm:"size:50;not null" json:"child_class"`
	Password     string    `gorm:"size:255;not null" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	TutorSessions []TutorSession       `gorm:"foreignKey:UserID" json:"-"`
	TutorMemories []TutorSubjectMemory `gorm:"foreignKey:UserID" json:"-"`
}

func (User) TableName() string { return "users" }

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
