package parser_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/omriashke/agent-secrets-cli/internal/parser"
)

func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("writeTempFile: %v", err)
	}
	return path
}

func TestParseDef(t *testing.T) {
	dir := t.TempDir()
	path := writeTempFile(t, dir, "secrets.def", `
API_KEY="OpenAI API key for GPT-4"
DB_PASS="Postgres password"
`)
	m, err := parser.ParseDef(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m["API_KEY"] != "OpenAI API key for GPT-4" {
		t.Errorf("API_KEY = %q, want %q", m["API_KEY"], "OpenAI API key for GPT-4")
	}
	if m["DB_PASS"] != "Postgres password" {
		t.Errorf("DB_PASS = %q, want %q", m["DB_PASS"], "Postgres password")
	}
}

func TestParseSecrets(t *testing.T) {
	dir := t.TempDir()
	path := writeTempFile(t, dir, ".secrets", `
API_KEY=sk-abc123
DB_PASS=hunter2
`)
	m, err := parser.ParseSecrets(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m["API_KEY"] != "sk-abc123" {
		t.Errorf("API_KEY = %q, want %q", m["API_KEY"], "sk-abc123")
	}
}

func TestMerge_Valid(t *testing.T) {
	defs := map[string]string{
		"API_KEY": "OpenAI API key",
		"DB_PASS": "Postgres password",
	}
	secrets := map[string]string{
		"API_KEY": "sk-abc123",
		"DB_PASS": "hunter2",
	}

	entries, warn, err := parser.Merge(defs, secrets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if warn != nil {
		t.Errorf("unexpected warning: %v", warn)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	// Entries are sorted by name — API_KEY < DB_PASS
	if entries[0].Name != "API_KEY" {
		t.Errorf("entries[0].Name = %q, want %q", entries[0].Name, "API_KEY")
	}
	if entries[0].Value != "sk-abc123" {
		t.Errorf("entries[0].Value = %q, want %q", entries[0].Value, "sk-abc123")
	}
	if entries[0].Description != "OpenAI API key" {
		t.Errorf("entries[0].Description = %q, want %q", entries[0].Description, "OpenAI API key")
	}
}

func TestMerge_MissingInSecrets(t *testing.T) {
	defs := map[string]string{
		"API_KEY": "OpenAI API key",
		"DB_PASS": "Postgres password",
	}
	secrets := map[string]string{
		"API_KEY": "sk-abc123",
		// DB_PASS missing — hard error
	}

	_, _, err := parser.Merge(defs, secrets)
	if err == nil {
		t.Fatal("expected error for missing key in .secrets, got nil")
	}
}

func TestMerge_ExtraInSecrets_IsWarning(t *testing.T) {
	defs := map[string]string{
		"API_KEY": "OpenAI API key",
	}
	secrets := map[string]string{
		"API_KEY": "sk-abc123",
		"DB_PASS": "hunter2", // extra key — should warn, not error
	}

	entries, warn, err := parser.Merge(defs, secrets)
	if err != nil {
		t.Fatalf("unexpected hard error: %v", err)
	}
	if warn == nil {
		t.Fatal("expected a warning for extra key in .secrets, got nil")
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry (only from defs), got %d", len(entries))
	}
	if entries[0].Name != "API_KEY" {
		t.Errorf("entries[0].Name = %q, want API_KEY", entries[0].Name)
	}
}

func TestMerge_Empty(t *testing.T) {
	entries, warn, err := parser.Merge(map[string]string{}, map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if warn != nil {
		t.Errorf("unexpected warning: %v", warn)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}
