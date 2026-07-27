package prompt

import "strings"

func studentProfile(in Input) StudentProfile {
	name := strings.TrimSpace(in.StudentName)
	if name == "" {
		name = "friend"
	}
	return StudentProfile{
		Name:           name,
		Age:            in.ChildAge,
		Class:          strings.TrimSpace(in.ChildClass),
		CurrentSubject: strings.ToLower(strings.TrimSpace(in.Subject)),
	}
}
