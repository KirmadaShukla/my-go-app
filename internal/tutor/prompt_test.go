package tutor_test

import (
	"testing"

	"my-go-app/internal/tutor"
)

func TestNormalizeClass(t *testing.T) {
	cases := map[string]string{
		"3":       "3",
		"Class 5": "5",
		"10th":    "10",
		"1":       "1",
	}
	for in, want := range cases {
		got, err := tutor.NormalizeClass(in)
		if err != nil {
			t.Fatalf("NormalizeClass(%q): %v", in, err)
		}
		if got != want {
			t.Fatalf("NormalizeClass(%q)=%q, want %q", in, got, want)
		}
	}
	if _, err := tutor.NormalizeClass("11"); err == nil {
		t.Fatal("expected error for class 11")
	}
}

func TestIsValidSubject(t *testing.T) {
	if !tutor.IsValidSubject("Maths") {
		t.Fatal("maths should be valid")
	}
	if tutor.IsValidSubject("history") {
		t.Fatal("history should be invalid")
	}
}
