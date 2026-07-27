package prompt

import "strings"

func learningMemory(summary string) LearningMemory {
	summary = strings.TrimSpace(summary)
	mem := LearningMemory{
		HasHistory: summary != "",
		Summary:    summary,
		Rules: []string{
			"Use learning_memory.summary to continue from what this student already practiced.",
			"Do not repeat full lessons the student already mastered, unless they ask.",
			"If summary is empty, treat this as a fresh start for the subject.",
		},
	}
	return mem
}
