package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/omriashke/agent-secrets-cli/internal/fileutil"
)

// backupAndSetSecrets replaces ~/.agent-secrets/secrets.def and .secrets
// with the given content, returning a cleanup that restores the originals.
func backupAndSetSecrets(t *testing.T, defs, secrets string) func() {
	t.Helper()
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(home, ".agent-secrets")
	os.MkdirAll(dir, 0700)

	defPath := filepath.Join(dir, "secrets.def")
	secretsPath := filepath.Join(dir, ".secrets")

	origDef, defErr := os.ReadFile(defPath)
	origSecrets, secErr := os.ReadFile(secretsPath)

	os.WriteFile(defPath, []byte(defs), 0600)
	os.WriteFile(secretsPath, []byte(secrets), 0600)

	return func() {
		if defErr == nil {
			os.WriteFile(defPath, origDef, 0600)
		}
		if secErr == nil {
			os.WriteFile(secretsPath, origSecrets, 0600)
		}
	}
}

func readSecretFiles(t *testing.T) (defs, secrets map[string]string) {
	t.Helper()
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".agent-secrets")

	defs, err := fileutil.ReadDotenv(filepath.Join(dir, "secrets.def"))
	if err != nil {
		t.Fatal(err)
	}
	secrets, err = fileutil.ReadDotenv(filepath.Join(dir, ".secrets"))
	if err != nil {
		t.Fatal(err)
	}
	return defs, secrets
}

// resetFlags clears persistent flag state that leaks between tests due to
// cobra using global command singletons.
func resetFlags() {
	addCmd.Flags().Set("description", "")
	addCmd.Flags().Set("value", "")
	editCmd.Flags().Set("description", "")
	editCmd.Flags().Set("value", "")
	deleteCmd.Flags().Set("yes", "false")
}

// executeRoot runs the root command with the given args and returns stdout and error.
func executeRoot(args ...string) (string, error) {
	resetFlags()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return buf.String(), err
}

// --- Add command ---

func TestAddCmd_Success(t *testing.T) {
	cleanup := backupAndSetSecrets(t,
		`EXISTING="existing description"`+"\n",
		`EXISTING="existing-value"`+"\n",
	)
	defer cleanup()

	output, err := executeRoot("add", "NEW_SECRET", "--description", "A new test secret", "--value", "new-secret-value")
	if err != nil {
		t.Fatalf("add command failed: %v", err)
	}

	if !strings.Contains(output, "NEW_SECRET added") {
		t.Errorf("expected success message, got: %q", output)
	}

	defs, secrets := readSecretFiles(t)
	if defs["NEW_SECRET"] != "A new test secret" {
		t.Errorf("def = %q, want %q", defs["NEW_SECRET"], "A new test secret")
	}
	if secrets["NEW_SECRET"] != "new-secret-value" {
		t.Errorf("secret = %q, want %q", secrets["NEW_SECRET"], "new-secret-value")
	}
	if defs["EXISTING"] != "existing description" {
		t.Errorf("EXISTING def was clobbered: %q", defs["EXISTING"])
	}
}

func TestAddCmd_DuplicateErrors(t *testing.T) {
	cleanup := backupAndSetSecrets(t,
		`EXISTING="existing description"`+"\n",
		`EXISTING="existing-value"`+"\n",
	)
	defer cleanup()

	_, err := executeRoot("add", "EXISTING", "--description", "dup", "--value", "dup")
	if err == nil {
		t.Fatal("expected error for duplicate, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should mention 'already exists', got: %v", err)
	}
}

func TestAddCmd_MissingDescription(t *testing.T) {
	cleanup := backupAndSetSecrets(t, "", "")
	defer cleanup()

	_, err := executeRoot("add", "KEY", "--value", "val")
	if err == nil {
		t.Fatal("expected error for missing --description, got nil")
	}
	if !strings.Contains(err.Error(), "--description is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAddCmd_MissingValue(t *testing.T) {
	cleanup := backupAndSetSecrets(t, "", "")
	defer cleanup()

	_, err := executeRoot("add", "KEY", "--description", "desc")
	if err == nil {
		t.Fatal("expected error for missing --value, got nil")
	}
	if !strings.Contains(err.Error(), "--value is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

// --- Delete command ---

func TestDeleteCmd_WithYes(t *testing.T) {
	cleanup := backupAndSetSecrets(t,
		"KEY_A=\"desc A\"\nKEY_B=\"desc B\"\n",
		"KEY_A=\"val-a\"\nKEY_B=\"val-b\"\n",
	)
	defer cleanup()

	output, err := executeRoot("delete", "KEY_A", "--yes")
	if err != nil {
		t.Fatalf("delete command failed: %v", err)
	}

	if !strings.Contains(output, "KEY_A deleted") {
		t.Errorf("expected success message, got: %q", output)
	}

	defs, secrets := readSecretFiles(t)
	if _, ok := defs["KEY_A"]; ok {
		t.Error("KEY_A should have been deleted from defs")
	}
	if _, ok := secrets["KEY_A"]; ok {
		t.Error("KEY_A should have been deleted from secrets")
	}
	if defs["KEY_B"] != "desc B" {
		t.Errorf("KEY_B def was clobbered: %q", defs["KEY_B"])
	}
}

func TestDeleteCmd_NotFound(t *testing.T) {
	cleanup := backupAndSetSecrets(t,
		`KEY_A="desc A"`+"\n",
		`KEY_A="val-a"`+"\n",
	)
	defer cleanup()

	_, err := executeRoot("delete", "NONEXISTENT", "--yes")
	if err == nil {
		t.Fatal("expected error for missing key, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestDeleteCmd_Aliases(t *testing.T) {
	cleanup := backupAndSetSecrets(t,
		"TO_RM=\"desc\"\n",
		"TO_RM=\"val\"\n",
	)
	defer cleanup()

	output, err := executeRoot("rm", "TO_RM", "--yes")
	if err != nil {
		t.Fatalf("rm alias failed: %v", err)
	}
	if !strings.Contains(output, "TO_RM deleted") {
		t.Errorf("expected success message, got: %q", output)
	}
}

// --- Edit command (inline mode) ---

func TestEditCmd_UpdateValue(t *testing.T) {
	cleanup := backupAndSetSecrets(t,
		`MY_KEY="original description"`+"\n",
		`MY_KEY="original-value"`+"\n",
	)
	defer cleanup()

	output, err := executeRoot("edit", "MY_KEY", "--value", "updated-value")
	if err != nil {
		t.Fatalf("edit command failed: %v", err)
	}

	if !strings.Contains(output, "MY_KEY updated") {
		t.Errorf("expected success message, got: %q", output)
	}

	defs, secrets := readSecretFiles(t)
	if defs["MY_KEY"] != "original description" {
		t.Errorf("description should be unchanged, got: %q", defs["MY_KEY"])
	}
	if secrets["MY_KEY"] != "updated-value" {
		t.Errorf("value = %q, want %q", secrets["MY_KEY"], "updated-value")
	}
}

func TestEditCmd_UpdateDescription(t *testing.T) {
	cleanup := backupAndSetSecrets(t,
		`MY_KEY="old desc"`+"\n",
		`MY_KEY="the-value"`+"\n",
	)
	defer cleanup()

	output, err := executeRoot("edit", "MY_KEY", "--description", "new desc")
	if err != nil {
		t.Fatalf("edit command failed: %v", err)
	}

	if !strings.Contains(output, "MY_KEY updated") {
		t.Errorf("expected success message, got: %q", output)
	}

	defs, secrets := readSecretFiles(t)
	if defs["MY_KEY"] != "new desc" {
		t.Errorf("description = %q, want %q", defs["MY_KEY"], "new desc")
	}
	if secrets["MY_KEY"] != "the-value" {
		t.Errorf("value should be unchanged, got: %q", secrets["MY_KEY"])
	}
}

func TestEditCmd_UpdateBoth(t *testing.T) {
	cleanup := backupAndSetSecrets(t,
		`MY_KEY="old desc"`+"\n",
		`MY_KEY="old-value"`+"\n",
	)
	defer cleanup()

	output, err := executeRoot("edit", "MY_KEY", "--description", "new desc", "--value", "new-value")
	if err != nil {
		t.Fatalf("edit command failed: %v", err)
	}

	if !strings.Contains(output, "MY_KEY updated") {
		t.Errorf("expected success message, got: %q", output)
	}

	defs, secrets := readSecretFiles(t)
	if defs["MY_KEY"] != "new desc" {
		t.Errorf("description = %q, want %q", defs["MY_KEY"], "new desc")
	}
	if secrets["MY_KEY"] != "new-value" {
		t.Errorf("value = %q, want %q", secrets["MY_KEY"], "new-value")
	}
}

func TestEditCmd_NotFound(t *testing.T) {
	cleanup := backupAndSetSecrets(t,
		`OTHER="desc"`+"\n",
		`OTHER="val"`+"\n",
	)
	defer cleanup()

	_, err := executeRoot("edit", "MISSING", "--value", "x")
	if err == nil {
		t.Fatal("expected error for missing key, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestEditCmd_NameWithoutFlags(t *testing.T) {
	cleanup := backupAndSetSecrets(t,
		`MY_KEY="desc"`+"\n",
		`MY_KEY="val"`+"\n",
	)
	defer cleanup()

	_, err := executeRoot("edit", "MY_KEY")
	if err == nil {
		t.Fatal("expected error when NAME given without flags, got nil")
	}
	if !strings.Contains(err.Error(), "--description and/or --value") {
		t.Errorf("unexpected error: %v", err)
	}
}

// --- Full lifecycle: add → edit → delete ---

func TestAddEditDelete_Lifecycle(t *testing.T) {
	cleanup := backupAndSetSecrets(t, "", "")
	defer cleanup()

	// Add
	output, err := executeRoot("add", "LIFECYCLE_KEY", "--description", "lifecycle test", "--value", "v1")
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if !strings.Contains(output, "LIFECYCLE_KEY added") {
		t.Errorf("add output: %q", output)
	}

	defs, secrets := readSecretFiles(t)
	if defs["LIFECYCLE_KEY"] != "lifecycle test" {
		t.Fatalf("add failed: def = %q", defs["LIFECYCLE_KEY"])
	}
	if secrets["LIFECYCLE_KEY"] != "v1" {
		t.Fatalf("add failed: secret = %q", secrets["LIFECYCLE_KEY"])
	}

	// Edit
	output, err = executeRoot("edit", "LIFECYCLE_KEY", "--value", "v2", "--description", "updated lifecycle")
	if err != nil {
		t.Fatalf("edit: %v", err)
	}
	if !strings.Contains(output, "LIFECYCLE_KEY updated") {
		t.Errorf("edit output: %q", output)
	}

	defs, secrets = readSecretFiles(t)
	if defs["LIFECYCLE_KEY"] != "updated lifecycle" {
		t.Fatalf("edit failed: def = %q", defs["LIFECYCLE_KEY"])
	}
	if secrets["LIFECYCLE_KEY"] != "v2" {
		t.Fatalf("edit failed: secret = %q", secrets["LIFECYCLE_KEY"])
	}

	// Delete
	output, err = executeRoot("delete", "LIFECYCLE_KEY", "--yes")
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if !strings.Contains(output, "LIFECYCLE_KEY deleted") {
		t.Errorf("delete output: %q", output)
	}

	defs, secrets = readSecretFiles(t)
	if _, ok := defs["LIFECYCLE_KEY"]; ok {
		t.Error("delete failed: key still in defs")
	}
	if _, ok := secrets["LIFECYCLE_KEY"]; ok {
		t.Error("delete failed: key still in secrets")
	}
}
