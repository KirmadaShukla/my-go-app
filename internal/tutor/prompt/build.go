package prompt

import (
	"encoding/json"
	"strings"
)

const directive = "Follow every field in this JSON configuration exactly. You are the learning buddy described in persona. Obey safety_rules, discussion_style, language, subject, learning_memory, mode, and response_constraints on every reply."

func buildDocument(in Input) Document {
	subject := strings.ToLower(strings.TrimSpace(in.Subject))
	mode := in.Mode
	if mode == "" {
		mode = ModeVoice
	}

	return Document{
		Directive:           directive,
		Persona:             basePersona(),
		SafetyRules:         safetyRules(),
		DiscussionStyle:     discussionStyle(),
		Student:             studentProfile(in),
		Language:            languageConfig(in.Language),
		Subject:             subjectConfig(subject),
		LearningMemory:      learningMemory(in.MemorySummary),
		Mode:                modeConfig(mode),
		ResponseConstraints: sharedResponseConstraints(),
	}
}

// Build assembles a structured JSON system prompt for high instruction accuracy.
func Build(in Input) string {
	raw, err := json.Marshal(buildDocument(in))
	if err != nil {
		return `{"directive":"Be a polite kids learning buddy.","error":"prompt_marshal_failed"}`
	}
	return string(raw)
}

// BuildPretty returns indented JSON (useful for debugging / logging).
func BuildPretty(in Input) string {
	raw, err := json.MarshalIndent(buildDocument(in), "", "  ")
	if err != nil {
		return Build(in)
	}
	return string(raw)
}
