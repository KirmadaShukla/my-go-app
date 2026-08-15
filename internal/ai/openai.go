package ai

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"my-go-app/internal/openai"
)

type openaiProvider struct {
	client *openai.Client
}

func NewOpenAI(apiKey, chatModel, ttsModel, ttsVoice string) Provider {
	return &openaiProvider{
		client: openai.NewClient(apiKey, chatModel, ttsModel, ttsVoice),
	}
}

func (p *openaiProvider) Name() string { return "openai" }

func (p *openaiProvider) Enabled() bool { return p.client.Enabled() }

func (p *openaiProvider) MissingConfig() string {
	return "OPENAI_API_KEY is not configured"
}

func (p *openaiProvider) Chat(ctx context.Context, messages []ChatMessage) (string, error) {
	return p.client.Chat(ctx, toOpenAIMessages(messages))
}

func (p *openaiProvider) VoiceTurn(ctx context.Context, system string, history []ChatMessage, audio []byte, filename string) (string, string, string, error) {
	if len(audio) == 0 {
		return "", "", "", fmt.Errorf("empty audio")
	}

	studentText, detectedLang, err := p.client.Transcribe(ctx, filename, bytes.NewReader(audio))
	if err != nil {
		return "", "", "", err
	}
	if strings.TrimSpace(studentText) == "" {
		return "", "", "", fmt.Errorf("could not understand the audio")
	}

	messages := make([]ChatMessage, 0, len(history)+2)
	if strings.TrimSpace(system) != "" {
		messages = append(messages, ChatMessage{Role: "system", Content: system})
	}
	messages = append(messages, history...)
	messages = append(messages, ChatMessage{Role: "user", Content: studentText})

	reply, err := p.Chat(ctx, messages)
	if err != nil {
		return "", "", "", err
	}
	return studentText, reply, detectedLang, nil
}

func (p *openaiProvider) Speak(ctx context.Context, text string) ([]byte, string, error) {
	audio, err := p.client.Speak(ctx, text)
	if err != nil {
		return nil, "", err
	}
	return audio, "audio/mpeg", nil
}

func toOpenAIMessages(msgs []ChatMessage) []openai.ChatMessage {
	out := make([]openai.ChatMessage, len(msgs))
	for i, m := range msgs {
		out[i] = openai.ChatMessage{Role: m.Role, Content: m.Content}
	}
	return out
}
