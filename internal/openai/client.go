package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

const baseURL = "https://api.openai.com/v1"

type Client struct {
	apiKey     string
	chatModel  string
	ttsModel   string
	ttsVoice   string
	httpClient *http.Client
}

func NewClient(apiKey, chatModel, ttsModel, ttsVoice string) *Client {
	return &Client{
		apiKey:    apiKey,
		chatModel: chatModel,
		ttsModel:  ttsModel,
		ttsVoice:  ttsVoice,
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
	}
}

func (c *Client) Enabled() bool {
	return strings.TrimSpace(c.apiKey) != ""
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
	Error *apiError `json:"error,omitempty"`
}

type apiError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

func (c *Client) Chat(ctx context.Context, messages []ChatMessage) (string, error) {
	if !c.Enabled() {
		return "", fmt.Errorf("OPENAI_API_KEY is not set")
	}

	body, err := json.Marshal(chatRequest{
		Model:    c.chatModel,
		Messages: messages,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai chat: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("decode chat response: %w", err)
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("openai chat: %s", parsed.Error.Message)
	}
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("openai chat: status %d: %s", resp.StatusCode, string(raw))
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("openai chat: empty choices")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}

type transcriptionResponse struct {
	Text     string    `json:"text"`
	Language string    `json:"language"`
	Error    *apiError `json:"error,omitempty"`
}

func (c *Client) Transcribe(ctx context.Context, filename string, audio io.Reader) (text, language string, err error) {
	if !c.Enabled() {
		return "", "", fmt.Errorf("OPENAI_API_KEY is not set")
	}

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("model", "whisper-1")
	_ = w.WriteField("response_format", "verbose_json")
	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		return "", "", err
	}
	if _, err := io.Copy(part, audio); err != nil {
		return "", "", err
	}
	if err := w.Close(); err != nil {
		return "", "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/audio/transcriptions", &buf)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("openai transcribe: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	var parsed transcriptionResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", "", fmt.Errorf("decode transcription: %w", err)
	}
	if parsed.Error != nil {
		return "", "", fmt.Errorf("openai transcribe: %s", parsed.Error.Message)
	}
	if resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("openai transcribe: status %d: %s", resp.StatusCode, string(raw))
	}
	return strings.TrimSpace(parsed.Text), strings.TrimSpace(parsed.Language), nil
}

type ttsRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
	Voice string `json:"voice"`
}

func (c *Client) Speak(ctx context.Context, text string) ([]byte, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("OPENAI_API_KEY is not set")
	}

	body, err := json.Marshal(ttsRequest{
		Model: c.ttsModel,
		Input: text,
		Voice: c.ttsVoice,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/audio/speech", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai tts: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("openai tts: status %d: %s", resp.StatusCode, string(raw))
	}
	return raw, nil
}
