package prompt_test

import (
	"encoding/json"
	"testing"

	"my-go-app/internal/tutor/prompt"
)

func TestBuildReturnsValidJSON(t *testing.T) {
	raw := prompt.Build(prompt.Input{
		StudentName: "Riya",
		ChildAge:    8,
		ChildClass:  "3",
		Subject:     "maths",
		Language:    "Hindi",
		Mode:        prompt.ModeVoice,
	})

	var doc prompt.Document
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, raw)
	}

	if doc.Student.Name != "Riya" {
		t.Fatalf("student.name = %q", doc.Student.Name)
	}
	if doc.Language.ReplyIn != "Hindi" {
		t.Fatalf("language.reply_in = %q", doc.Language.ReplyIn)
	}
	if doc.Subject.ID != "maths" {
		t.Fatalf("subject.id = %q", doc.Subject.ID)
	}
	if doc.Mode.ID != "voice" {
		t.Fatalf("mode.id = %q", doc.Mode.ID)
	}
	if len(doc.SafetyRules) == 0 || len(doc.DiscussionStyle) == 0 {
		t.Fatal("expected safety and discussion rules")
	}
}

func TestBuildVoiceScience(t *testing.T) {
	raw := prompt.Build(prompt.Input{
		StudentName: "Aarav",
		ChildAge:    10,
		ChildClass:  "5",
		Subject:     "science",
		Language:    "English",
		Mode:        prompt.ModeVoice,
	})

	var doc prompt.Document
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if doc.Subject.ID != "science" {
		t.Fatalf("subject.id = %q", doc.Subject.ID)
	}
	if doc.Mode.ID != "voice" {
		t.Fatalf("mode.id = %q", doc.Mode.ID)
	}
}

func TestBuildGreetingEnglish(t *testing.T) {
	raw := prompt.Build(prompt.Input{
		StudentName: "Riya",
		ChildAge:    8,
		ChildClass:  "3",
		Subject:     "english",
		Language:    "Hindi",
		Mode:        prompt.ModeGreeting,
	})

	var doc prompt.Document
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if doc.Mode.ID != "greeting" {
		t.Fatalf("mode.id = %q", doc.Mode.ID)
	}
	if doc.Subject.ID != "english" {
		t.Fatalf("subject.id = %q", doc.Subject.ID)
	}
}
