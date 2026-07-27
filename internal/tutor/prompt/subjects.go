package prompt

func subjectConfig(subject string) SubjectConfig {
	switch subject {
	case "maths":
		return SubjectConfig{
			ID:    "maths",
			Title: "Maths",
			Focus: "numbers, arithmetic, geometry, measurement, and word problems for the student's class",
			Guidelines: []string{
				"Break problems into small, clear steps.",
				"Check understanding after each step when useful.",
				"Use everyday examples (toys, money, food, school) for younger classes.",
				"Encourage the student to try calculating before you reveal the final answer.",
				"If they make a mistake, correct gently and explain why.",
				"Keep notation simple and age-appropriate.",
			},
		}
	case "science":
		return SubjectConfig{
			ID:    "science",
			Title: "Science",
			Focus: "science ideas with curiosity and everyday examples for the student's class",
			Guidelines: []string{
				"Explain concepts with real-life examples the student can imagine.",
				"Prefer why/how discussion over memorizing definitions.",
				"Use simple experiments or observations they can do safely at home/school when useful.",
				"Correct misconceptions kindly.",
				"Keep facts accurate and age-appropriate; say when something is a simplified explanation.",
			},
		}
	case "english":
		return SubjectConfig{
			ID:    "english",
			Title: "English",
			Focus: "reading, writing, vocabulary, grammar, listening, and speaking for the student's class",
			Guidelines: []string{
				"Model correct language politely without shaming mistakes.",
				"For younger classes: focus on simple words, sentences, and reading confidence.",
				"For older classes: support grammar, comprehension, and clearer writing.",
				"Give short practice prompts (speak / write / choose the better sentence).",
				"Praise effort and clear communication.",
			},
		}
	case "activities":
		return SubjectConfig{
			ID:    "activities",
			Title: "Activities",
			Focus: "fun educational activities, games, and creative practice suitable for the student's class",
			Guidelines: []string{
				"Keep activities safe, screen-friendly or offline-friendly when possible.",
				"Connect each activity to a learning goal (maths, science, english, creativity, or thinking skills).",
				"Give clear short steps the student can follow.",
				"Offer easier and slightly harder versions when useful.",
				"Keep the tone playful and encouraging.",
			},
		}
	default:
		return SubjectConfig{
			ID:    "general",
			Title: "General learning",
			Focus: "help the student learn through friendly discussion for their class level",
			Guidelines: []string{
				"Stay supportive and conversational.",
				"Keep explanations age-appropriate.",
			},
		}
	}
}
