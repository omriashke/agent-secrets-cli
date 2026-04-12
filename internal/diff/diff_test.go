package diff_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/omriashke/agent-secrets-cli/internal/diff"
)

// --- MaskValue ---

func TestMaskValue_ShortValue(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"abc", "***"},
		{"1234567890", "**********"},
	}
	for _, tt := range tests {
		got := diff.MaskValue(tt.input)
		if got != tt.want {
			t.Errorf("MaskValue(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMaskValue_LongValue(t *testing.T) {
	got := diff.MaskValue("sk-abc123xyz789end")
	if !strings.HasPrefix(got, "sk-a") {
		t.Errorf("expected prefix 'sk-a', got %q", got)
	}
	if !strings.HasSuffix(got, "9end") {
		t.Errorf("expected suffix '9end', got %q", got)
	}
	if !strings.Contains(got, "***") {
		t.Errorf("expected masked middle, got %q", got)
	}
	if len(got) != len("sk-abc123xyz789end") {
		t.Errorf("masked length %d != original length %d", len(got), len("sk-abc123xyz789end"))
	}
}

func TestMaskValue_ExactlyElevenChars(t *testing.T) {
	input := "12345678901"
	got := diff.MaskValue(input)
	if got != "1234***8901" {
		t.Errorf("MaskValue(%q) = %q, want %q", input, got, "1234***8901")
	}
}

// --- ComputeChanges ---

func writeDotenv(t *testing.T, dir, name string, entries map[string]string) string {
	t.Helper()
	var b strings.Builder
	for k, v := range entries {
		b.WriteString(k + `="` + v + `"` + "\n")
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(b.String()), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestComputeChanges_NoChanges(t *testing.T) {
	dir := t.TempDir()
	defs := map[string]string{"KEY": "desc"}
	secrets := map[string]string{"KEY": "val"}

	localDef := writeDotenv(t, dir, "local.def", defs)
	localSec := writeDotenv(t, dir, "local.sec", secrets)
	remoteDef := writeDotenv(t, dir, "remote.def", defs)
	remoteSec := writeDotenv(t, dir, "remote.sec", secrets)

	changes, err := diff.ComputeChanges(localDef, localSec, remoteDef, remoteSec)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 0 {
		t.Errorf("expected no changes, got %d", len(changes))
	}
}

func TestComputeChanges_AddedSecret(t *testing.T) {
	dir := t.TempDir()

	localDef := writeDotenv(t, dir, "local.def", map[string]string{
		"KEY_A": "desc A",
		"KEY_B": "desc B",
	})
	localSec := writeDotenv(t, dir, "local.sec", map[string]string{
		"KEY_A": "val-a",
		"KEY_B": "val-b",
	})
	remoteDef := writeDotenv(t, dir, "remote.def", map[string]string{
		"KEY_A": "desc A",
	})
	remoteSec := writeDotenv(t, dir, "remote.sec", map[string]string{
		"KEY_A": "val-a",
	})

	changes, err := diff.ComputeChanges(localDef, localSec, remoteDef, remoteSec)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Kind != "added" {
		t.Errorf("Kind = %q, want 'added'", changes[0].Kind)
	}
	if changes[0].Name != "KEY_B" {
		t.Errorf("Name = %q, want 'KEY_B'", changes[0].Name)
	}
}

func TestComputeChanges_RemovedSecret(t *testing.T) {
	dir := t.TempDir()

	localDef := writeDotenv(t, dir, "local.def", map[string]string{
		"KEY_A": "desc A",
	})
	localSec := writeDotenv(t, dir, "local.sec", map[string]string{
		"KEY_A": "val-a",
	})
	remoteDef := writeDotenv(t, dir, "remote.def", map[string]string{
		"KEY_A": "desc A",
		"KEY_B": "desc B",
	})
	remoteSec := writeDotenv(t, dir, "remote.sec", map[string]string{
		"KEY_A": "val-a",
		"KEY_B": "val-b",
	})

	changes, err := diff.ComputeChanges(localDef, localSec, remoteDef, remoteSec)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Kind != "removed" {
		t.Errorf("Kind = %q, want 'removed'", changes[0].Kind)
	}
	if changes[0].Name != "KEY_B" {
		t.Errorf("Name = %q, want 'KEY_B'", changes[0].Name)
	}
}

func TestComputeChanges_ChangedValue(t *testing.T) {
	dir := t.TempDir()

	localDef := writeDotenv(t, dir, "local.def", map[string]string{"KEY": "desc"})
	localSec := writeDotenv(t, dir, "local.sec", map[string]string{"KEY": "new-val"})
	remoteDef := writeDotenv(t, dir, "remote.def", map[string]string{"KEY": "desc"})
	remoteSec := writeDotenv(t, dir, "remote.sec", map[string]string{"KEY": "old-val"})

	changes, err := diff.ComputeChanges(localDef, localSec, remoteDef, remoteSec)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Kind != "changed" {
		t.Errorf("Kind = %q, want 'changed'", changes[0].Kind)
	}
	if changes[0].OldValue != "old-val" {
		t.Errorf("OldValue = %q, want 'old-val'", changes[0].OldValue)
	}
	if changes[0].NewValue != "new-val" {
		t.Errorf("NewValue = %q, want 'new-val'", changes[0].NewValue)
	}
}

func TestComputeChanges_ChangedDescription(t *testing.T) {
	dir := t.TempDir()

	localDef := writeDotenv(t, dir, "local.def", map[string]string{"KEY": "new desc"})
	localSec := writeDotenv(t, dir, "local.sec", map[string]string{"KEY": "same-val"})
	remoteDef := writeDotenv(t, dir, "remote.def", map[string]string{"KEY": "old desc"})
	remoteSec := writeDotenv(t, dir, "remote.sec", map[string]string{"KEY": "same-val"})

	changes, err := diff.ComputeChanges(localDef, localSec, remoteDef, remoteSec)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Kind != "changed" {
		t.Errorf("Kind = %q, want 'changed'", changes[0].Kind)
	}
	if changes[0].Description != "new desc" {
		t.Errorf("Description = %q, want 'new desc'", changes[0].Description)
	}
}

func TestComputeChanges_MultipleChanges(t *testing.T) {
	dir := t.TempDir()

	localDef := writeDotenv(t, dir, "local.def", map[string]string{
		"KEPT":    "same",
		"ADDED":   "new secret",
		"CHANGED": "updated desc",
	})
	localSec := writeDotenv(t, dir, "local.sec", map[string]string{
		"KEPT":    "same-val",
		"ADDED":   "added-val",
		"CHANGED": "changed-val",
	})
	remoteDef := writeDotenv(t, dir, "remote.def", map[string]string{
		"KEPT":    "same",
		"REMOVED": "will be removed",
		"CHANGED": "old desc",
	})
	remoteSec := writeDotenv(t, dir, "remote.sec", map[string]string{
		"KEPT":    "same-val",
		"REMOVED": "removed-val",
		"CHANGED": "old-val",
	})

	changes, err := diff.ComputeChanges(localDef, localSec, remoteDef, remoteSec)
	if err != nil {
		t.Fatal(err)
	}

	kinds := map[string]bool{}
	for _, c := range changes {
		kinds[c.Kind] = true
	}
	if !kinds["added"] {
		t.Error("expected an 'added' change")
	}
	if !kinds["removed"] {
		t.Error("expected a 'removed' change")
	}
	if !kinds["changed"] {
		t.Error("expected a 'changed' change")
	}
	if len(changes) != 3 {
		t.Errorf("expected 3 changes (added, removed, changed), got %d", len(changes))
	}
}

func TestComputeChanges_SortedByName(t *testing.T) {
	dir := t.TempDir()

	localDef := writeDotenv(t, dir, "local.def", map[string]string{
		"ZEBRA": "z",
		"APPLE": "a",
		"MANGO": "m",
	})
	localSec := writeDotenv(t, dir, "local.sec", map[string]string{
		"ZEBRA": "z",
		"APPLE": "a",
		"MANGO": "m",
	})
	remoteDef := writeDotenv(t, dir, "remote.def", map[string]string{})
	remoteSec := writeDotenv(t, dir, "remote.sec", map[string]string{})

	changes, err := diff.ComputeChanges(localDef, localSec, remoteDef, remoteSec)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 3 {
		t.Fatalf("expected 3 changes, got %d", len(changes))
	}
	if changes[0].Name != "APPLE" || changes[1].Name != "MANGO" || changes[2].Name != "ZEBRA" {
		t.Errorf("changes not sorted: %s, %s, %s", changes[0].Name, changes[1].Name, changes[2].Name)
	}
}

// --- PrintChanges ---

func TestPrintChanges_NoChanges(t *testing.T) {
	var buf bytes.Buffer
	diff.PrintChanges(&buf, nil, "push")
	if !strings.Contains(buf.String(), "no-op") {
		t.Errorf("expected no-op message, got: %q", buf.String())
	}
}

func TestPrintChanges_Added(t *testing.T) {
	var buf bytes.Buffer
	changes := []diff.Change{
		{Kind: "added", Name: "NEW_KEY", Description: "A new secret", NewValue: "secret-value"},
	}
	diff.PrintChanges(&buf, changes, "push")
	out := buf.String()

	if !strings.Contains(out, "+ NEW_KEY") {
		t.Errorf("expected '+ NEW_KEY' in output:\n%s", out)
	}
	if !strings.Contains(out, "A new secret") {
		t.Errorf("expected description in output:\n%s", out)
	}
	if strings.Contains(out, "secret-value") {
		t.Errorf("raw value should be masked, not shown in output:\n%s", out)
	}
}

func TestPrintChanges_Removed(t *testing.T) {
	var buf bytes.Buffer
	changes := []diff.Change{
		{Kind: "removed", Name: "OLD_KEY", Description: "An old secret"},
	}
	diff.PrintChanges(&buf, changes, "pull")
	out := buf.String()

	if !strings.Contains(out, "- OLD_KEY") {
		t.Errorf("expected '- OLD_KEY' in output:\n%s", out)
	}
	if !strings.Contains(out, "pull") {
		t.Errorf("expected direction 'pull' in output:\n%s", out)
	}
}

func TestPrintChanges_Changed(t *testing.T) {
	var buf bytes.Buffer
	changes := []diff.Change{
		{Kind: "changed", Name: "MY_KEY", Description: "Updated", OldValue: "old", NewValue: "new"},
	}
	diff.PrintChanges(&buf, changes, "push")
	out := buf.String()

	if !strings.Contains(out, "~ MY_KEY") {
		t.Errorf("expected '~ MY_KEY' in output:\n%s", out)
	}
	if !strings.Contains(out, "→") {
		t.Errorf("expected arrow in changed output:\n%s", out)
	}
}

// --- ComputeChanges with empty/missing files ---

func TestComputeChanges_EmptyRemote(t *testing.T) {
	dir := t.TempDir()

	localDef := writeDotenv(t, dir, "local.def", map[string]string{"KEY": "desc"})
	localSec := writeDotenv(t, dir, "local.sec", map[string]string{"KEY": "val"})
	remoteDef := writeDotenv(t, dir, "remote.def", map[string]string{})
	remoteSec := writeDotenv(t, dir, "remote.sec", map[string]string{})

	changes, err := diff.ComputeChanges(localDef, localSec, remoteDef, remoteSec)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 || changes[0].Kind != "added" {
		t.Errorf("expected 1 added change, got %v", changes)
	}
}

func TestComputeChanges_EmptyLocal(t *testing.T) {
	dir := t.TempDir()

	localDef := writeDotenv(t, dir, "local.def", map[string]string{})
	localSec := writeDotenv(t, dir, "local.sec", map[string]string{})
	remoteDef := writeDotenv(t, dir, "remote.def", map[string]string{"KEY": "desc"})
	remoteSec := writeDotenv(t, dir, "remote.sec", map[string]string{"KEY": "val"})

	changes, err := diff.ComputeChanges(localDef, localSec, remoteDef, remoteSec)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 || changes[0].Kind != "removed" {
		t.Errorf("expected 1 removed change, got %v", changes)
	}
}
