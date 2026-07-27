package prompt

// Document is the structured system prompt sent to the model as JSON.
// Keeping fields explicit improves instruction-following accuracy.
type Document struct {
	Directive            string         `json:"directive"`
	Persona              Persona        `json:"persona"`
	SafetyRules          []string       `json:"safety_rules"`
	DiscussionStyle      []string       `json:"discussion_style"`
	Student              StudentProfile `json:"student"`
	Language             LanguageConfig `json:"language"`
	Subject              SubjectConfig  `json:"subject"`
	LearningMemory       LearningMemory `json:"learning_memory"`
	Mode                 ModeConfig     `json:"mode"`
	ResponseConstraints  []string       `json:"response_constraints"`
}

type Persona struct {
	Role       string `json:"role"`
	Audience   string `json:"audience"`
	Goal       string `json:"goal"`
	Tone       string `json:"tone"`
	Style      string `json:"style"`
	ClassRange string `json:"class_range"`
}

type StudentProfile struct {
	Name           string `json:"name"`
	Age            int    `json:"age"`
	Class          string `json:"class"`
	CurrentSubject string `json:"current_subject"`
}

type LanguageConfig struct {
	ReplyIn string   `json:"reply_in"`
	Rules   []string `json:"rules"`
}

type SubjectConfig struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	Focus      string   `json:"focus"`
	Guidelines []string `json:"guidelines"`
}

type LearningMemory struct {
	HasHistory bool   `json:"has_history"`
	Summary    string `json:"summary"`
	Rules      []string `json:"rules"`
}

type ModeConfig struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Instructions []string `json:"instructions"`
}
