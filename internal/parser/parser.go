package parser

import (
	"fmt"
	"sort"

	"github.com/joho/godotenv"
)

// Entry holds a single secret's name, description, and value after merging.
type Entry struct {
	Name        string
	Description string
	Value       string
}

// MergeWarning is a non-fatal notice returned alongside a successful Merge.
type MergeWarning struct {
	Messages []string
}

func (w *MergeWarning) Error() string {
	msg := "warnings:\n"
	for _, m := range w.Messages {
		msg += "  " + m + "\n"
	}
	return msg
}

// ParseDef parses secrets.def. Values in this file are human-readable descriptions.
func ParseDef(path string) (map[string]string, error) {
	m, err := godotenv.Read(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	return m, nil
}

// ParseSecrets parses .secrets. Values in this file are actual secret values.
func ParseSecrets(path string) (map[string]string, error) {
	m, err := godotenv.Read(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	return m, nil
}

// Merge combines definitions and secrets into a validated slice of Entry.
// Keys present in secrets.def but missing from .secrets are hard errors —
// the agent would get an empty value. Extra keys in .secrets that have no
// description are silently skipped (they may be env vars the user doesn't
// want to expose to agents) but reported as warnings via MergeWarning.
func Merge(defs, secrets map[string]string) ([]Entry, *MergeWarning, error) {
	var hardErrors []string
	var warnings []string

	for name := range defs {
		if _, ok := secrets[name]; !ok {
			hardErrors = append(hardErrors, fmt.Sprintf("%s is in secrets.def but has no value in .secrets", name))
		}
	}

	if len(hardErrors) > 0 {
		sort.Strings(hardErrors)
		msg := "missing secret values — add these to ~/.agent-secrets/.secrets:\n"
		for _, e := range hardErrors {
			msg += "  " + e + "\n"
		}
		return nil, nil, fmt.Errorf("%s", msg)
	}

	for name := range secrets {
		if _, ok := defs[name]; !ok {
			warnings = append(warnings, fmt.Sprintf("%s is in .secrets but has no description in secrets.def — skipping", name))
		}
	}

	entries := make([]Entry, 0, len(defs))
	for name, desc := range defs {
		entries = append(entries, Entry{
			Name:        name,
			Description: desc,
			Value:       secrets[name],
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	var warn *MergeWarning
	if len(warnings) > 0 {
		sort.Strings(warnings)
		warn = &MergeWarning{Messages: warnings}
	}
	return entries, warn, nil
}
