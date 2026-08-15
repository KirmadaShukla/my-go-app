package gemini

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

const baseURL = "https://generativelanguage.googleapis.com/v1beta"

// Client talks to the free Google AI Studio Gemini API.
type Client struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewClient(apiKey, model string) *Client {
	if strings.TrimSpace(model) == "" {
		model = "gemini-2.5-flash-lite"
	}
	return &Client{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
	}
}

func (c *Client) Enabled() bool {
	return strings.TrimSpace(c.apiKey) != ""
}

// ChatMessage mirrors OpenAI-style roles for easy swapping.
type ChatMessage struct {
	Role    string `json:"role"` // system | user | assistant
	Content string `json:"content"`
}

type generateRequest struct {
	SystemInstruction *content          `json:"systemInstruction,omitempty"`
	Contents          []content         `json:"contents"`
	GenerationConfig  *generationConfig `json:"generationConfig,omitempty"`
}

type generationConfig struct {
	ResponseMIMEType string `json:"responseMimeType,omitempty"`
}

type content struct {
	Role  string `json:"role,omitempty"`
	Parts []part `json:"parts"`
}

type part struct {
	Text       string      `json:"text,omitempty"`
	InlineData *inlineData `json:"inlineData,omitempty"`
}

type inlineData struct {
	MIMEType string `json:"mimeType"`
	Data     string `json:"data"`
}

type generateResponse struct {
	Candidates []struct {
		Content content `json:"content"`
	} `json:"candidates"`
	Error *apiError `json:"error,omitempty"`
}

type apiError struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// Chat sends a text conversation to Gemini and returns the assistant reply.
func (c *Client) Chat(ctx context.Context, messages []ChatMessage) (string, error) {
	if !c.Enabled() {
		return "", fmt.Errorf("GEMINI_API_KEY is not set")
	}

	reqBody, err := buildChatRequest(messages, nil)
	if err != nil {
		return "", err
	}

	raw, err := c.doGenerate(ctx, reqBody)
	if err != nil {
		return "", err
	}
	return extractText(raw)
}

// VoiceTurn sends student audio + history to Gemini and returns heard text + reply.
// Free-tier friendly: no OpenAI Whisper/TTS required.
func (c *Client) VoiceTurn(ctx context.Context, system string, history []ChatMessage, audio []byte, filename string) (studentText, replyText string, err error) {
	if !c.Enabled() {
		return "", "", fmt.Errorf("GEMINI_API_KEY is not set")
	}
	if len(audio) == 0 {
		return "", "", fmt.Errorf("empty audio")
	}

	msgs := make([]ChatMessage, 0, len(history)+2)
	if strings.TrimSpace(system) != "" {
		msgs = append(msgs, ChatMessage{Role: "system", Content: system})
	}
	msgs = append(msgs, history...)
	msgs = append(msgs, ChatMessage{
		Role: "user",
		Content: `Listen to the attached student audio.
Return ONLY valid JSON with keys:
- student_text: what the student said (best-effort transcript)
- reply_text: your spoken-style tutor reply following the system instructions
Do not wrap JSON in markdown.`,
	})

	reqBody, err := buildChatRequest(msgs, &inlineData{
		MIMEType: audioMIME(filename),
		Data:     base64.StdEncoding.EncodeToString(audio),
	})
	if err != nil {
		return "", "", err
	}
	reqBody.GenerationConfig = &generationConfig{ResponseMIMEType: "application/json"}

	raw, err := c.doGenerate(ctx, reqBody)
	if err != nil {
		return "", "", err
	}
	text, err := extractText(raw)
	if err != nil {
		return "", "", err
	}

	var parsed struct {
		StudentText string `json:"student_text"`
		ReplyText   string `json:"reply_text"`
	}
	if err := json.Unmarshal([]byte(stripCodeFence(text)), &parsed); err != nil {
		// Fallback: treat whole response as reply if JSON parse fails.
		return "(audio)", strings.TrimSpace(text), nil
	}
	if strings.TrimSpace(parsed.ReplyText) == "" {
		return "", "", fmt.Errorf("gemini returned empty reply_text")
	}
	if strings.TrimSpace(parsed.StudentText) == "" {
		parsed.StudentText = "(audio)"
	}
	return strings.TrimSpace(parsed.StudentText), strings.TrimSpace(parsed.ReplyText), nil
}

func buildChatRequest(messages []ChatMessage, audio *inlineData) (*generateRequest, error) {
	var systemText string
	contents := make([]content, 0, len(messages))

	for i, m := range messages {
		role := strings.ToLower(strings.TrimSpace(m.Role))
		switch role {
		case "system":
			if systemText != "" {
				systemText += "\n\n"
			}
			systemText += m.Content
		case "assistant", "model":
			contents = append(contents, content{
				Role:  "model",
				Parts: []part{{Text: m.Content}},
			})
		default: // user
			p := []part{{Text: m.Content}}
			// Attach audio to the last user turn when provided.
			if audio != nil && i == len(messages)-1 {
				p = append([]part{{InlineData: audio}}, p...)
			}
			contents = append(contents, content{
				Role:  "user",
				Parts: p,
			})
		}
	}

	if len(contents) == 0 {
		return nil, fmt.Errorf("no user/model messages to send")
	}

	req := &generateRequest{Contents: contents}
	if systemText != "" {
		req.SystemInstruction = &content{Parts: []part{{Text: systemText}}}
	}
	return req, nil
}

func (c *Client) doGenerate(ctx context.Context, body *generateRequest) ([]byte, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", baseURL, c.model, c.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini generate: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var parsed generateResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("decode gemini response: %w", err)
	}
	if parsed.Error != nil {
		return nil, fmt.Errorf("gemini: %s", parsed.Error.Message)
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("gemini: status %d: %s", resp.StatusCode, string(raw))
	}
	return raw, nil
}

func extractText(raw []byte) (string, error) {
	var parsed generateResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini: empty candidates")
	}
	var b strings.Builder
	for _, p := range parsed.Candidates[0].Content.Parts {
		b.WriteString(p.Text)
	}
	return strings.TrimSpace(b.String()), nil
}

func audioMIME(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".wav":
		return "audio/wav"
	case ".mp3":
		return "audio/mp3"
	case ".m4a", ".mp4", ".aac":
		return "audio/mp4"
	case ".ogg", ".oga":
		return "audio/ogg"
	case ".flac":
		return "audio/flac"
	case ".webm":
		return "audio/webm"
	default:
		return "audio/webm"
	}
}

func stripCodeFence(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```JSON")
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	}
	return s
}
