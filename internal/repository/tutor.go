package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"my-go-app/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TutorRepository struct {
	db *gorm.DB
}

func NewTutorRepository(db *gorm.DB) *TutorRepository {
	return &TutorRepository{db: db}
}

func (r *TutorRepository) CreateSession(ctx context.Context, session *model.TutorSession) error {
	if session.Status == "" {
		session.Status = model.TutorSessionStatusActive
	}
	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		return fmt.Errorf("create tutor session: %w", err)
	}
	return nil
}

func (r *TutorRepository) FindSession(ctx context.Context, sessionID, userID uuid.UUID) (*model.TutorSession, error) {
	var session model.TutorSession
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", sessionID, userID).
		First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find tutor session: %w", err)
	}
	return &session, nil
}

// FindLatestActiveSession returns the newest active session for a user+subject (for resume).
func (r *TutorRepository) FindLatestActiveSession(ctx context.Context, userID uuid.UUID, subject string) (*model.TutorSession, error) {
	var session model.TutorSession
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND subject = ? AND status = ?", userID, subject, model.TutorSessionStatusActive).
		Order("updated_at DESC").
		First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find active tutor session: %w", err)
	}
	return &session, nil
}

func (r *TutorRepository) TouchSession(ctx context.Context, sessionID uuid.UUID) error {
	now := time.Now().UTC()
	err := r.db.WithContext(ctx).Model(&model.TutorSession{}).
		Where("id = ?", sessionID).
		Updates(map[string]any{
			"updated_at":      now,
			"last_message_at": now,
			"message_count":   gorm.Expr("message_count + 1"),
		}).Error
	if err != nil {
		return fmt.Errorf("touch tutor session: %w", err)
	}
	return nil
}

func (r *TutorRepository) CloseSession(ctx context.Context, sessionID, userID uuid.UUID) error {
	res := r.db.WithContext(ctx).Model(&model.TutorSession{}).
		Where("id = ? AND user_id = ?", sessionID, userID).
		Update("status", model.TutorSessionStatusClosed)
	if res.Error != nil {
		return fmt.Errorf("close tutor session: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *TutorRepository) NextSequence(ctx context.Context, sessionID uuid.UUID) (int64, error) {
	var maxSeq *int64
	err := r.db.WithContext(ctx).Model(&model.TutorMessage{}).
		Select("MAX(sequence)").
		Where("session_id = ?", sessionID).
		Scan(&maxSeq).Error
	if err != nil {
		return 0, fmt.Errorf("next message sequence: %w", err)
	}
	if maxSeq == nil {
		return 1, nil
	}
	return *maxSeq + 1, nil
}

func (r *TutorRepository) AddMessage(ctx context.Context, msg *model.TutorMessage) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if msg.Sequence == 0 {
			var maxSeq *int64
			if err := tx.Model(&model.TutorMessage{}).
				Select("MAX(sequence)").
				Where("session_id = ?", msg.SessionID).
				Scan(&maxSeq).Error; err != nil {
				return fmt.Errorf("message sequence: %w", err)
			}
			if maxSeq == nil {
				msg.Sequence = 1
			} else {
				msg.Sequence = *maxSeq + 1
			}
		}
		if msg.Channel == "" {
			msg.Channel = model.TutorMessageChannelVoice
		}
		if err := tx.Create(msg).Error; err != nil {
			return fmt.Errorf("create tutor message: %w", err)
		}

		now := time.Now().UTC()
		if err := tx.Model(&model.TutorSession{}).
			Where("id = ?", msg.SessionID).
			Updates(map[string]any{
				"updated_at":      now,
				"last_message_at": now,
				"message_count":   gorm.Expr("message_count + 1"),
			}).Error; err != nil {
			return fmt.Errorf("update session counters: %w", err)
		}
		return nil
	})
}

func (r *TutorRepository) ListRecentMessages(ctx context.Context, sessionID uuid.UUID, limit int) ([]model.TutorMessage, error) {
	if limit <= 0 {
		limit = 12
	}
	var messages []model.TutorMessage
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("sequence DESC").
		Limit(limit).
		Find(&messages).Error
	if err != nil {
		return nil, fmt.Errorf("list tutor messages: %w", err)
	}
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}

func (r *TutorRepository) GetSubjectMemory(ctx context.Context, userID uuid.UUID, subject string) (*model.TutorSubjectMemory, error) {
	var memory model.TutorSubjectMemory
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND subject = ?", userID, subject).
		First(&memory).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get subject memory: %w", err)
	}
	return &memory, nil
}

func (r *TutorRepository) UpsertSubjectMemory(ctx context.Context, memory *model.TutorSubjectMemory) error {
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "subject"}},
		DoUpdates: clause.AssignmentColumns([]string{"summary", "updated_at"}),
	}).Create(memory).Error
	if err != nil {
		return fmt.Errorf("upsert subject memory: %w", err)
	}
	return nil
}
