package fileutil

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/joho/godotenv"
)

// ReadDotenv reads a dotenv file and returns a map of key=value pairs.
// Returns an empty map (not an error) if the file is empty or all comments.
func ReadDotenv(path string) (map[string]string, error) {
	m, err := godotenv.Read(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil, fmt.Errorf("cannot read %s: %w", path, err)
		}
		if strings.TrimSpace(string(data)) == "" || allComments(string(data)) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("cannot parse %s: %w", path, err)
	}
	return m, nil
}

func allComments(content string) bool {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		return false
	}
	return true
}

// WriteDotenv writes a map of key=value pairs to a dotenv file, preserving
// comments from the original file and maintaining sorted key order.
func WriteDotenv(path string, entries map[string]string) error {
	existing, _ := os.ReadFile(path)
	comments := extractComments(string(existing))

	var b strings.Builder
	if comments != "" {
		b.WriteString(comments)
		if !strings.HasSuffix(comments, "\n") {
			b.WriteString("\n")
		}
	}

	keys := make([]string, 0, len(entries))
	for k := range entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		b.WriteString(fmt.Sprintf("%s=%q\n", k, entries[k]))
	}

	return os.WriteFile(path, []byte(b.String()), 0600)
}

// SetKey adds or updates a single key in a dotenv file.
func SetKey(path, key, value string) error {
	m, err := ReadDotenv(path)
	if err != nil {
		return err
	}
	m[key] = value
	return WriteDotenv(path, m)
}

// DeleteKey removes a key from a dotenv file.
func DeleteKey(path, key string) error {
	m, err := ReadDotenv(path)
	if err != nil {
		return err
	}
	if _, ok := m[key]; !ok {
		return fmt.Errorf("key %q not found in %s", key, path)
	}
	delete(m, key)
	return WriteDotenv(path, m)
}

func extractComments(content string) string {
	var comments strings.Builder
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			comments.WriteString(line)
			comments.WriteString("\n")
		}
	}
	return comments.String()
}
