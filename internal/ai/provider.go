package ai

import "context"

// ChatMessage is the shared chat turn shape for every tutor provider.
type ChatMessage struct {
	Role    string // system | user | assistant
	Content string
}

// Provider is the tutor AI backend (OpenAI or Gemini).
type Provider interface {
	Name() string
	Enabled() bool
	MissingConfig() string
	Chat(ctx context.Context, messages []ChatMessage) (string, error)
	VoiceTurn(ctx context.Context, system string, history []ChatMessage, audio []byte, filename string) (studentText, replyText, detectedLang string, err error)
	// Speak returns TTS audio when the provider supports it; otherwise (nil, "", nil).
	Speak(ctx context.Context, text string) (audio []byte, mime string, err error)
}
