package languages

import "strings"

type Language struct {
	Name        string   // canonical name (e.g. "go", "cpp")
	Aliases     []string // accepted input values
	DisplayName string   // optional friendly name
}

// Resolver provides normalization and validation for language codes.
type Resolver interface {
	Normalize(raw string) string
	Supported() []string
}

type simpleResolver struct {
	langs []Language
}

func NewResolver(langs []Language) Resolver { return &simpleResolver{langs: langs} }

func (r *simpleResolver) Normalize(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	for _, l := range r.langs {
		if v == l.Name { return l.Name }
		for _, a := range l.Aliases { if v == a { return l.Name } }
	}
	return ""
}

func (r *simpleResolver) Supported() []string {
	out := make([]string, 0, len(r.langs))
	for _, l := range r.langs { out = append(out, l.Name) }
	return out
}

