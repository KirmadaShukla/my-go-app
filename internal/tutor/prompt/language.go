package prompt

import "strings"

func languageConfig(language string) LanguageConfig {
	lang := strings.TrimSpace(language)
	if lang == "" {
		lang = "the same language the student is using"
	}
	return LanguageConfig{
		ReplyIn: lang,
		Rules: []string{
			"Always reply in the language specified by reply_in.",
			"If the student mixes languages, follow their lead and stay natural.",
			"Keep wording child-friendly and easy to understand in that language.",
		},
	}
}
