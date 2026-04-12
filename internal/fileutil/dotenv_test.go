package fileutil_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/omriashke/agent-secrets-cli/internal/fileutil"
)

func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("writeTempFile: %v", err)
	}
	return path
}

// --- ReadDotenv ---

func TestReadDotenv_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := writeTempFile(t, dir, "test.env", `API_KEY="my-key"
DB_PASS="hunter2"
`)
	m, err := fileutil.ReadDotenv(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m["API_KEY"] != "my-key" {
		t.Errorf("API_KEY = %q, want %q", m["API_KEY"], "my-key")
	}
	if m["DB_PASS"] != "hunter2" {
		t.Errorf("DB_PASS = %q, want %q", m["DB_PASS"], "hunter2")
	}
}

func TestReadDotenv_NonExistentFile(t *testing.T) {
	m, err := fileutil.ReadDotenv("/tmp/does-not-exist-agent-secrets-test")
	if err != nil {
		t.Fatalf("expected empty map for missing file, got error: %v", err)
	}
	if len(m) != 0 {
		t.Errorf("expected empty map, got %v", m)
	}
}

func TestReadDotenv_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := writeTempFile(t, dir, "empty.env", "")
	m, err := fileutil.ReadDotenv(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m) != 0 {
		t.Errorf("expected empty map for empty file, got %v", m)
	}
}

func TestReadDotenv_AllComments(t *testing.T) {
	dir := t.TempDir()
	path := writeTempFile(t, dir, "comments.env", `# This is a comment
# Another comment
`)
	m, err := fileutil.ReadDotenv(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m) != 0 {
		t.Errorf("expected empty map for all-comments file, got %v", m)
	}
}

func TestReadDotenv_MixedCommentsAndValues(t *testing.T) {
	dir := t.TempDir()
	path := writeTempFile(t, dir, "mixed.env", `# Header comment
API_KEY="value1"
# Middle comment
DB_PASS="value2"
`)
	m, err := fileutil.ReadDotenv(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m) != 2 {
		t.Errorf("expected 2 entries, got %d", len(m))
	}
	if m["API_KEY"] != "value1" {
		t.Errorf("API_KEY = %q, want %q", m["API_KEY"], "value1")
	}
}

// --- WriteDotenv ---

func TestWriteDotenv_WritesEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.env")

	entries := map[string]string{
		"ZEBRA": "z-value",
		"APPLE": "a-value",
	}
	if err := fileutil.WriteDotenv(path, entries); err != nil {
		t.Fatalf("WriteDotenv: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if !strings.Contains(content, `APPLE="a-value"`) {
		t.Errorf("missing APPLE entry in output:\n%s", content)
	}
	if !strings.Contains(content, `ZEBRA="z-value"`) {
		t.Errorf("missing ZEBRA entry in output:\n%s", content)
	}

	// Verify sorted order: APPLE before ZEBRA
	appleIdx := strings.Index(content, "APPLE")
	zebraIdx := strings.Index(content, "ZEBRA")
	if appleIdx > zebraIdx {
		t.Error("entries not sorted: APPLE should come before ZEBRA")
	}
}

func TestWriteDotenv_PreservesComments(t *testing.T) {
	dir := t.TempDir()
	path := writeTempFile(t, dir, "existing.env", `# This is a header comment
# Another comment
OLD_KEY="old-value"
`)

	entries := map[string]string{"NEW_KEY": "new-value"}
	if err := fileutil.WriteDotenv(path, entries); err != nil {
		t.Fatalf("WriteDotenv: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, "# This is a header comment") {
		t.Error("comment not preserved")
	}
	if !strings.Contains(content, `NEW_KEY="new-value"`) {
		t.Error("new entry not written")
	}
	if strings.Contains(content, "OLD_KEY") {
		t.Error("old entry should have been replaced (not in new map)")
	}
}

func TestWriteDotenv_EmptyMap(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.env")

	if err := fileutil.WriteDotenv(path, map[string]string{}); err != nil {
		t.Fatalf("WriteDotenv: %v", err)
	}

	data, _ := os.ReadFile(path)
	if strings.TrimSpace(string(data)) != "" {
		t.Errorf("expected empty file for empty map, got: %q", string(data))
	}
}

// --- SetKey ---

func TestSetKey_AddsNewKey(t *testing.T) {
	dir := t.TempDir()
	path := writeTempFile(t, dir, "set.env", `EXISTING="value"
`)

	if err := fileutil.SetKey(path, "NEW_KEY", "new-value"); err != nil {
		t.Fatalf("SetKey: %v", err)
	}

	m, _ := fileutil.ReadDotenv(path)
	if m["EXISTING"] != "value" {
		t.Errorf("EXISTING = %q, want %q", m["EXISTING"], "value")
	}
	if m["NEW_KEY"] != "new-value" {
		t.Errorf("NEW_KEY = %q, want %q", m["NEW_KEY"], "new-value")
	}
}

func TestSetKey_UpdatesExistingKey(t *testing.T) {
	dir := t.TempDir()
	path := writeTempFile(t, dir, "set.env", `MY_KEY="old"
`)

	if err := fileutil.SetKey(path, "MY_KEY", "updated"); err != nil {
		t.Fatalf("SetKey: %v", err)
	}

	m, _ := fileutil.ReadDotenv(path)
	if m["MY_KEY"] != "updated" {
		t.Errorf("MY_KEY = %q, want %q", m["MY_KEY"], "updated")
	}
}

func TestSetKey_CreatesFileIfMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new.env")

	if err := fileutil.SetKey(path, "KEY", "val"); err != nil {
		t.Fatalf("SetKey: %v", err)
	}

	m, err := fileutil.ReadDotenv(path)
	if err != nil {
		t.Fatal(err)
	}
	if m["KEY"] != "val" {
		t.Errorf("KEY = %q, want %q", m["KEY"], "val")
	}
}

// --- DeleteKey ---

func TestDeleteKey_RemovesKey(t *testing.T) {
	dir := t.TempDir()
	path := writeTempFile(t, dir, "del.env", `KEY_A="a"
KEY_B="b"
`)

	if err := fileutil.DeleteKey(path, "KEY_A"); err != nil {
		t.Fatalf("DeleteKey: %v", err)
	}

	m, _ := fileutil.ReadDotenv(path)
	if _, ok := m["KEY_A"]; ok {
		t.Error("KEY_A should have been deleted")
	}
	if m["KEY_B"] != "b" {
		t.Errorf("KEY_B = %q, want %q", m["KEY_B"], "b")
	}
}

func TestDeleteKey_ErrorOnMissingKey(t *testing.T) {
	dir := t.TempDir()
	path := writeTempFile(t, dir, "del.env", `KEY_A="a"
`)

	err := fileutil.DeleteKey(path, "NONEXISTENT")
	if err == nil {
		t.Fatal("expected error for missing key, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestDeleteKey_LastKey(t *testing.T) {
	dir := t.TempDir()
	path := writeTempFile(t, dir, "del.env", `ONLY_KEY="value"
`)

	if err := fileutil.DeleteKey(path, "ONLY_KEY"); err != nil {
		t.Fatalf("DeleteKey: %v", err)
	}

	m, _ := fileutil.ReadDotenv(path)
	if len(m) != 0 {
		t.Errorf("expected empty map after deleting last key, got %v", m)
	}
}

// --- Round-trip ---

func TestRoundTrip_SetThenDeleteThenRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "roundtrip.env")

	fileutil.SetKey(path, "A", "1")
	fileutil.SetKey(path, "B", "2")
	fileutil.SetKey(path, "C", "3")

	m, _ := fileutil.ReadDotenv(path)
	if len(m) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(m))
	}

	fileutil.DeleteKey(path, "B")

	m, _ = fileutil.ReadDotenv(path)
	if len(m) != 2 {
		t.Fatalf("expected 2 keys after delete, got %d", len(m))
	}
	if _, ok := m["B"]; ok {
		t.Error("B should be deleted")
	}
	if m["A"] != "1" || m["C"] != "3" {
		t.Errorf("unexpected values: %v", m)
	}
}
