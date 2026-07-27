package prompt

func modeConfig(mode Mode) ModeConfig {
	switch mode {
	case ModeGreeting:
		return ModeConfig{
			ID:    "greeting",
			Title: "spoken session greeting",
			Instructions: []string{
				"This greeting will be spoken aloud to the student.",
				"Greet the student warmly by name if available.",
				"Briefly say what subject you will learn together.",
				"Invite them with one short, friendly discussion question.",
				"Keep this first reply short (2-4 spoken sentences).",
				"Avoid markdown, lists, or symbols that sound awkward when spoken.",
			},
		}
	case ModeVoice:
		return ModeConfig{
			ID:    "voice",
			Title: "voice conversation",
			Instructions: []string{
				"The student spoke this turn with their voice.",
				"Keep replies short and natural for speaking aloud (about 2-5 short sentences).",
				"Prefer spoken rhythm: simple sentences, clear pauses, no long lists.",
				"Avoid markdown, bullet walls, or code blocks unless the student asks.",
				"End with one gentle question to continue the discussion.",
			},
		}
	case ModeChat:
		fallthrough
	default:
		// Chat is not exposed as an API right now; keep voice-safe defaults.
		return ModeConfig{
			ID:    "voice",
			Title: "voice conversation",
			Instructions: []string{
				"Keep replies short and natural for speaking aloud.",
				"End with one gentle question to continue the discussion.",
			},
		}
	}
}
