package ai

import (
	"context"

	"my-go-app/internal/gemini"
)

type geminiProvider struct {
	client *gemini.Client
}

func NewGemini(apiKey, model string) Provider {
	return &geminiProvider{client: gemini.NewClient(apiKey, model)}
}

func (p *geminiProvider) Name() string { return "gemini" }

func (p *geminiProvider) Enabled() bool { return p.client.Enabled() }

func (p *geminiProvider) MissingConfig() string {
	return "GEMINI_API_KEY is not configured"
}

func (p *geminiProvider) Chat(ctx context.Context, messages []ChatMessage) (string, error) {
	return p.client.Chat(ctx, toGeminiMessages(messages))
}

func (p *geminiProvider) VoiceTurn(ctx context.Context, system string, history []ChatMessage, audio []byte, filename string) (string, string, string, error) {
	student, reply, err := p.client.VoiceTurn(ctx, system, toGeminiMessages(history), audio, filename)
	return student, reply, "", err
}

func (p *geminiProvider) Speak(context.Context, string) ([]byte, string, error) {
	// Gemini path is text-only for now (no TTS).
	return nil, "", nil
}

func toGeminiMessages(msgs []ChatMessage) []gemini.ChatMessage {
	out := make([]gemini.ChatMessage, len(msgs))
	for i, m := range msgs {
		out[i] = gemini.ChatMessage{Role: m.Role, Content: m.Content}
	}
	return out
}
