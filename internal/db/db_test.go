package db_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/omriashke/agent-secrets-cli/internal/db"
	"github.com/omriashke/agent-secrets-cli/internal/parser"
)

func openTestDB(t *testing.T) (*os.File, interface{ Close() error }) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	database, err := db.Open(path)
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	return nil, database
}

func setupDB(t *testing.T) (database interface{ Close() error }, defPath, secretsPath string) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	d, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}

	defPath = filepath.Join(dir, "secrets.def")
	secretsPath = filepath.Join(dir, ".secrets")

	if err := os.WriteFile(defPath, []byte(`
API_KEY="OpenAI API key for GPT-4 calls"
DB_PASS="Postgres password for production"
STRIPE="Stripe secret key for payments"
`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(secretsPath, []byte(`
API_KEY=sk-abc123
DB_PASS=hunter2
STRIPE=sk_live_xyz
`), 0600); err != nil {
		t.Fatal(err)
	}

	if err := db.AutoSync(d, defPath, secretsPath); err != nil {
		t.Fatalf("AutoSync: %v", err)
	}

	return d, defPath, secretsPath
}

func TestAutoSync_PopulatesDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	defPath := filepath.Join(dir, "secrets.def")
	secretsPath := filepath.Join(dir, ".secrets")
	os.WriteFile(defPath, []byte(`API_KEY="OpenAI key"`+"\n"), 0600)
	os.WriteFile(secretsPath, []byte("API_KEY=sk-test\n"), 0600)

	if err := db.AutoSync(database, defPath, secretsPath); err != nil {
		t.Fatalf("AutoSync: %v", err)
	}

	secrets, err := db.List(database)
	if err != nil {
		t.Fatal(err)
	}
	if len(secrets) != 1 {
		t.Fatalf("expected 1 secret, got %d", len(secrets))
	}
	if secrets[0].Name != "API_KEY" {
		t.Errorf("Name = %q, want API_KEY", secrets[0].Name)
	}
	if secrets[0].Value != "sk-test" {
		t.Errorf("Value = %q, want sk-test", secrets[0].Value)
	}
}

func TestAutoSync_NoOpWhenUnchanged(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	defPath := filepath.Join(dir, "secrets.def")
	secretsPath := filepath.Join(dir, ".secrets")
	os.WriteFile(defPath, []byte(`API_KEY="OpenAI key"`+"\n"), 0600)
	os.WriteFile(secretsPath, []byte("API_KEY=sk-test\n"), 0600)

	// First sync
	if err := db.AutoSync(database, defPath, secretsPath); err != nil {
		t.Fatal(err)
	}

	// Mutate the in-memory content but don't touch files — second sync should be no-op
	// We verify by checking the value is still from the first sync
	secrets, _ := db.List(database)
	firstUUID := secrets[0].UUID

	if err := db.AutoSync(database, defPath, secretsPath); err != nil {
		t.Fatal(err)
	}

	secrets, _ = db.List(database)
	if secrets[0].UUID != firstUUID {
		t.Error("UUID changed on second sync — expected no-op")
	}
}

func TestAutoSync_PreservesUUIDs(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	defPath := filepath.Join(dir, "secrets.def")
	secretsPath := filepath.Join(dir, ".secrets")
	os.WriteFile(defPath, []byte(`API_KEY="OpenAI key"`+"\n"), 0600)
	os.WriteFile(secretsPath, []byte("API_KEY=sk-test\n"), 0600)

	db.AutoSync(database, defPath, secretsPath)
	secrets, _ := db.List(database)
	firstUUID := secrets[0].UUID

	// Update file content and mtime to trigger re-sync
	time.Sleep(10 * time.Millisecond)
	os.WriteFile(defPath, []byte(`API_KEY="Updated description"`+"\n"), 0600)

	db.AutoSync(database, defPath, secretsPath)
	secrets, _ = db.List(database)

	if secrets[0].UUID != firstUUID {
		t.Errorf("UUID changed after re-sync: got %q, want %q", secrets[0].UUID, firstUUID)
	}
	if secrets[0].Description != "Updated description" {
		t.Errorf("Description not updated: got %q", secrets[0].Description)
	}
}

func TestAutoSync_RemovesDeletedSecrets(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	defPath := filepath.Join(dir, "secrets.def")
	secretsPath := filepath.Join(dir, ".secrets")
	os.WriteFile(defPath, []byte("API_KEY=\"key\"\nDB_PASS=\"pass\"\n"), 0600)
	os.WriteFile(secretsPath, []byte("API_KEY=sk\nDB_PASS=pw\n"), 0600)
	db.AutoSync(database, defPath, secretsPath)

	secrets, _ := db.List(database)
	if len(secrets) != 2 {
		t.Fatalf("expected 2 secrets after first sync, got %d", len(secrets))
	}

	// Remove DB_PASS from both files
	time.Sleep(10 * time.Millisecond)
	os.WriteFile(defPath, []byte("API_KEY=\"key\"\n"), 0600)
	os.WriteFile(secretsPath, []byte("API_KEY=sk\n"), 0600)
	db.AutoSync(database, defPath, secretsPath)

	secrets, _ = db.List(database)
	if len(secrets) != 1 {
		t.Fatalf("expected 1 secret after removal, got %d", len(secrets))
	}
	if secrets[0].Name != "API_KEY" {
		t.Errorf("expected API_KEY to remain, got %q", secrets[0].Name)
	}
}

func TestList_OrderedByName(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	defPath := filepath.Join(dir, "secrets.def")
	secretsPath := filepath.Join(dir, ".secrets")
	os.WriteFile(defPath, []byte("ZEBRA=\"z\"\nAPPLE=\"a\"\nMIDDLE=\"m\"\n"), 0600)
	os.WriteFile(secretsPath, []byte("ZEBRA=z\nAPPLE=a\nMIDDLE=m\n"), 0600)
	db.AutoSync(database, defPath, secretsPath)

	secrets, err := db.List(database)
	if err != nil {
		t.Fatal(err)
	}
	names := []string{secrets[0].Name, secrets[1].Name, secrets[2].Name}
	want := []string{"APPLE", "MIDDLE", "ZEBRA"}
	for i, n := range names {
		if n != want[i] {
			t.Errorf("secrets[%d].Name = %q, want %q", i, n, want[i])
		}
	}
}

func TestQuery_FindsByDescription(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	defPath := filepath.Join(dir, "secrets.def")
	secretsPath := filepath.Join(dir, ".secrets")
	os.WriteFile(defPath, []byte(`
API_KEY="OpenAI API key for GPT-4 calls"
DB_PASS="Postgres password for production database"
`), 0600)
	os.WriteFile(secretsPath, []byte("API_KEY=sk-abc\nDB_PASS=hunter2\n"), 0600)
	db.AutoSync(database, defPath, secretsPath)

	tests := []struct {
		query     string
		wantName  string
		wantValue string
	}{
		{"OpenAI API key", "API_KEY", "sk-abc"},
		{"Postgres password", "DB_PASS", "hunter2"},
		{"production database", "DB_PASS", "hunter2"},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			s, err := db.Query(database, tt.query)
			if err != nil {
				t.Fatalf("Query(%q): %v", tt.query, err)
			}
			if s.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", s.Name, tt.wantName)
			}
			if s.Value != tt.wantValue {
				t.Errorf("Value = %q, want %q", s.Value, tt.wantValue)
			}
		})
	}
}

func TestQuery_NotFound(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	defPath := filepath.Join(dir, "secrets.def")
	secretsPath := filepath.Join(dir, ".secrets")
	os.WriteFile(defPath, []byte(`API_KEY="OpenAI key"`+"\n"), 0600)
	os.WriteFile(secretsPath, []byte("API_KEY=sk-abc\n"), 0600)
	db.AutoSync(database, defPath, secretsPath)

	_, err = db.Query(database, "something completely unrelated xyz")
	if err == nil {
		t.Fatal("expected error for no match, got nil")
	}
}

// Ensure parser.Merge is exercised via file round-trip
func TestParserMerge_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	defPath := filepath.Join(dir, "secrets.def")
	secretsPath := filepath.Join(dir, ".secrets")

	os.WriteFile(defPath, []byte(`KEY="description"`+"\n"), 0600)
	os.WriteFile(secretsPath, []byte("KEY=value\n"), 0600)

	defs, err := parser.ParseDef(defPath)
	if err != nil {
		t.Fatal(err)
	}
	secrets, err := parser.ParseSecrets(secretsPath)
	if err != nil {
		t.Fatal(err)
	}
	entries, _, err := parser.Merge(defs, secrets)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name != "KEY" || entries[0].Value != "value" {
		t.Errorf("unexpected entries: %+v", entries)
	}
}
