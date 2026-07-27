package prompt

// Mode selects which interaction-style instructions to append.
type Mode string

const (
	ModeGreeting Mode = "greeting"
	ModeChat     Mode = "chat"
	ModeVoice    Mode = "voice"
)

// Input is everything needed to assemble a clear JSON system prompt.
type Input struct {
	StudentName string
	ChildAge    int
	ChildClass  string // normalized "1".."10"
	Subject     string // maths | science | english | activities
	Language    string
	Mode        Mode
	MemorySummary string // long-term notes for this student+subject
}
