package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"

	"my-go-app/internal/model"
	"my-go-app/internal/openai"
	"my-go-app/internal/repository"
	"my-go-app/internal/tutor"
	"my-go-app/internal/tutor/prompt"

	"github.com/google/uuid"
)

var (
	ErrTutorUnavailable = errors.New("tutor unavailable")
	ErrInvalidSubject   = errors.New("invalid subject")
	ErrSessionNotFound  = errors.New("tutor session not found")
)

type TutorService struct {
	users  *repository.UserRepository
	tutors *repository.TutorRepository
	ai     *openai.Client
}

func NewTutorService(users *repository.UserRepository, tutors *repository.TutorRepository, ai *openai.Client) *TutorService {
	return &TutorService{users: users, tutors: tutors, ai: ai}
}

type StartSessionInput struct {
	UserID   uuid.UUID
	Subject  string
	Language string
	ForceNew bool // if true, always create a fresh session
}

type SessionResult struct {
	Session     *model.TutorSession
	Resumed     bool
	Greeting    string
	AudioBase64 string
	AudioMIME   string
}

func (s *TutorService) Subjects() []string {
	return tutor.Subjects
}

func (s *TutorService) StartSession(ctx context.Context, in StartSessionInput) (*SessionResult, error) {
	if !s.ai.Enabled() {
		return nil, fmt.Errorf("%w: OPENAI_API_KEY is not configured", ErrTutorUnavailable)
	}

	subject := strings.ToLower(strings.TrimSpace(in.Subject))
	if !tutor.IsValidSubject(subject) {
		return nil, fmt.Errorf("%w: use maths, science, english, or activities", ErrInvalidSubject)
	}

	user, err := s.users.FindByID(ctx, in.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	class, err := tutor.NormalizeClass(user.ChildClass)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrValidation, err.Error())
	}

	lang := strings.TrimSpace(in.Language)
	if lang == "" {
		lang = "the student's preferred language"
	}

	memorySummary := s.memorySummary(ctx, user.ID, subject)

	// Resume active session for same subject so GPT keeps short-term history.
	if !in.ForceNew {
		existing, err := s.tutors.FindLatestActiveSession(ctx, user.ID, subject)
		if err == nil {
			if lang != "" {
				existing.Language = lang
			}
			welcomeBack, err := s.ai.Chat(ctx, []openai.ChatMessage{
				{Role: "system", Content: prompt.Build(prompt.Input{
					StudentName:   user.Name,
					ChildAge:      user.ChildAge,
					ChildClass:    class,
					Subject:       subject,
					Language:      existing.Language,
					Mode:          prompt.ModeGreeting,
					MemorySummary: memorySummary,
				})},
				{Role: "user", Content: "The student is returning to an existing session. Greet them briefly, mention you remember their recent practice if learning_memory has history, and ask one short question to continue."},
			})
			if err != nil {
				return nil, fmt.Errorf("%w: %v", ErrTutorUnavailable, err)
			}

			audio, err := s.ai.Speak(ctx, welcomeBack)
			if err != nil {
				return nil, fmt.Errorf("%w: %v", ErrTutorUnavailable, err)
			}

			if err := s.tutors.AddMessage(ctx, &model.TutorMessage{
				SessionID: existing.ID,
				UserID:    user.ID,
				Role:      model.TutorMessageRoleAssistant,
				Channel:   model.TutorMessageChannelVoice,
				Content:   welcomeBack,
			}); err != nil {
				return nil, err
			}

			return &SessionResult{
				Session:     existing,
				Resumed:     true,
				Greeting:    welcomeBack,
				AudioBase64: base64.StdEncoding.EncodeToString(audio),
				AudioMIME:   "audio/mpeg",
			}, nil
		}
		if !errors.Is(err, repository.ErrNotFound) {
			return nil, err
		}
	}

	session := &model.TutorSession{
		UserID:   user.ID,
		Subject:  subject,
		Language: lang,
		Status:   model.TutorSessionStatusActive,
	}
	if err := s.tutors.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	greeting, err := s.ai.Chat(ctx, []openai.ChatMessage{
		{Role: "system", Content: prompt.Build(prompt.Input{
			StudentName:   user.Name,
			ChildAge:      user.ChildAge,
			ChildClass:    class,
			Subject:       subject,
			Language:      lang,
			Mode:          prompt.ModeGreeting,
			MemorySummary: memorySummary,
		})},
		{Role: "user", Content: prompt.GreetingUserMessage},
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTutorUnavailable, err)
	}

	audio, err := s.ai.Speak(ctx, greeting)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTutorUnavailable, err)
	}

	if err := s.tutors.AddMessage(ctx, &model.TutorMessage{
		SessionID: session.ID,
		UserID:    user.ID,
		Role:      model.TutorMessageRoleAssistant,
		Channel:   model.TutorMessageChannelVoice,
		Content:   greeting,
	}); err != nil {
		return nil, err
	}

	return &SessionResult{
		Session:     session,
		Resumed:     false,
		Greeting:    greeting,
		AudioBase64: base64.StdEncoding.EncodeToString(audio),
		AudioMIME:   "audio/mpeg",
	}, nil
}

type VoiceInput struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
	Filename  string
	Audio     io.Reader
}

type VoiceResult struct {
	StudentText string `json:"student_text"`
	ReplyText   string `json:"reply_text"`
	Language    string `json:"language"`
	AudioBase64 string `json:"audio_base64"`
	AudioMIME   string `json:"audio_mime"`
}

func (s *TutorService) Voice(ctx context.Context, in VoiceInput) (*VoiceResult, error) {
	if !s.ai.Enabled() {
		return nil, fmt.Errorf("%w: OPENAI_API_KEY is not configured", ErrTutorUnavailable)
	}

	user, session, class, err := s.loadContext(ctx, in.UserID, in.SessionID)
	if err != nil {
		return nil, err
	}

	filename := in.Filename
	if filename == "" {
		filename = "speech.webm"
	}

	studentText, detectedLang, err := s.ai.Transcribe(ctx, filename, in.Audio)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTutorUnavailable, err)
	}
	if studentText == "" {
		return nil, fmt.Errorf("%w: could not understand the audio", ErrValidation)
	}

	replyLanguage := session.Language
	if replyLanguage == "" || replyLanguage == "the student's preferred language" {
		if detectedLang != "" {
			replyLanguage = detectedLang
		} else {
			replyLanguage = "the same language the student just spoke"
		}
	}

	history, err := s.tutors.ListRecentMessages(ctx, session.ID, 20)
	if err != nil {
		return nil, err
	}

	memorySummary := s.memorySummary(ctx, user.ID, session.Subject)

	messages := []openai.ChatMessage{{
		Role: "system",
		Content: prompt.Build(prompt.Input{
			StudentName:   user.Name,
			ChildAge:      user.ChildAge,
			ChildClass:    class,
			Subject:       session.Subject,
			Language:      replyLanguage,
			Mode:          prompt.ModeVoice,
			MemorySummary: memorySummary,
		}),
	}}
	for _, m := range history {
		messages = append(messages, openai.ChatMessage{Role: m.Role, Content: m.Content})
	}
	messages = append(messages, openai.ChatMessage{Role: "user", Content: studentText})

	reply, err := s.ai.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTutorUnavailable, err)
	}

	audio, err := s.ai.Speak(ctx, reply)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTutorUnavailable, err)
	}

	if err := s.tutors.AddMessage(ctx, &model.TutorMessage{
		SessionID: session.ID,
		UserID:    user.ID,
		Role:      model.TutorMessageRoleUser,
		Channel:   model.TutorMessageChannelVoice,
		Content:   studentText,
	}); err != nil {
		return nil, err
	}
	if err := s.tutors.AddMessage(ctx, &model.TutorMessage{
		SessionID: session.ID,
		UserID:    user.ID,
		Role:      model.TutorMessageRoleAssistant,
		Channel:   model.TutorMessageChannelVoice,
		Content:   reply,
	}); err != nil {
		return nil, err
	}

	// Lightweight rolling memory so future sessions keep subject context.
	_ = s.tutors.UpsertSubjectMemory(ctx, &model.TutorSubjectMemory{
		UserID:  user.ID,
		Subject: session.Subject,
		Summary: trimMemory(fmt.Sprintf(
			"Recent practice in %s. Last student question: %s. Last tutor guidance: %s",
			session.Subject,
			truncate(studentText, 180),
			truncate(reply, 220),
		)),
	})

	return &VoiceResult{
		StudentText: studentText,
		ReplyText:   reply,
		Language:    replyLanguage,
		AudioBase64: base64.StdEncoding.EncodeToString(audio),
		AudioMIME:   "audio/mpeg",
	}, nil
}

func (s *TutorService) memorySummary(ctx context.Context, userID uuid.UUID, subject string) string {
	mem, err := s.tutors.GetSubjectMemory(ctx, userID, subject)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(mem.Summary)
}

func (s *TutorService) loadContext(ctx context.Context, userID, sessionID uuid.UUID) (*model.User, *model.TutorSession, string, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil, "", ErrInvalidCredentials
		}
		return nil, nil, "", err
	}

	session, err := s.tutors.FindSession(ctx, sessionID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil, "", ErrSessionNotFound
		}
		return nil, nil, "", err
	}

	class, err := tutor.NormalizeClass(user.ChildClass)
	if err != nil {
		return nil, nil, "", fmt.Errorf("%w: %s", ErrValidation, err.Error())
	}
	return user, session, class, nil
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func trimMemory(s string) string {
	return strings.TrimSpace(s)
}
