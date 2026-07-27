package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	TutorSessionStatusActive = "active"
	TutorSessionStatusClosed = "closed"

	TutorMessageRoleUser      = "user"
	TutorMessageRoleAssistant = "assistant"

	TutorMessageChannelVoice = "voice"
)

// TutorSession is one voice learning discussion for a student + subject.
type TutorSession struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null;index:idx_tutor_sessions_user_subject_status,priority:1;index:idx_tutor_sessions_user_updated,priority:1" json:"user_id"`
	Subject       string     `gorm:"size:32;not null;index:idx_tutor_sessions_user_subject_status,priority:2" json:"subject"`
	Language      string     `gorm:"size:64;not null" json:"language"`
	Status        string     `gorm:"size:16;not null;default:active;index:idx_tutor_sessions_user_subject_status,priority:3" json:"status"`
	MessageCount  int        `gorm:"not null;default:0" json:"message_count"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `gorm:"index:idx_tutor_sessions_user_updated,priority:2" json:"updated_at"`

	User     User           `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	Messages []TutorMessage `gorm:"foreignKey:SessionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"messages,omitempty"`
}

func (TutorSession) TableName() string { return "tutor_sessions" }

func (s *TutorSession) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.Status == "" {
		s.Status = TutorSessionStatusActive
	}
	return nil
}

// TutorMessage stores one spoken turn for conversation history / GPT context.
type TutorMessage struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SessionID uuid.UUID `gorm:"type:uuid;not null;index:idx_tutor_messages_session_seq,priority:1;index:idx_tutor_messages_session_created,priority:1" json:"session_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_tutor_messages_user_created,priority:1" json:"user_id"`
	Role      string    `gorm:"size:16;not null" json:"role"`       // user | assistant
	Channel   string    `gorm:"size:16;not null;default:voice" json:"channel"` // voice
	Content   string    `gorm:"type:text;not null" json:"content"`
	Sequence  int64     `gorm:"not null;index:idx_tutor_messages_session_seq,priority:2" json:"sequence"`
	CreatedAt time.Time `gorm:"index:idx_tutor_messages_session_created,priority:2;index:idx_tutor_messages_user_created,priority:2" json:"created_at"`

	Session TutorSession `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	User    User         `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func (TutorMessage) TableName() string { return "tutor_messages" }

func (m *TutorMessage) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	if m.Channel == "" {
		m.Channel = TutorMessageChannelVoice
	}
	return nil
}

// TutorSubjectMemory is long-term learning memory per student + subject.
// Used when the student returns later so GPT can continue effectively.
type TutorSubjectMemory struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:ux_tutor_subject_memory_user_subject,priority:1" json:"user_id"`
	Subject   string    `gorm:"size:32;not null;uniqueIndex:ux_tutor_subject_memory_user_subject,priority:2" json:"subject"`
	Summary   string    `gorm:"type:text;not null;default:''" json:"summary"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func (TutorSubjectMemory) TableName() string { return "tutor_subject_memories" }

func (m *TutorSubjectMemory) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
