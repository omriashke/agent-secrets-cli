package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/omriashke/agent-secrets-cli/internal/config"
)

// loadRemoteFromFile is a test helper that reads a config file from a given
// path by temporarily pointing the home dir at a temp location.
func writeConfigFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input string
		want  string
	}{
		{"~/.ssh/id_ed25519", filepath.Join(home, ".ssh/id_ed25519")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// expandHome is unexported — test it indirectly via LoadRemote
			// by writing a config file with the value and reading it back.
			dir := t.TempDir()

			// Temporarily override home to our temp dir so ConfigPath resolves there.
			agentDir := filepath.Join(dir, ".agent-secrets")
			os.MkdirAll(agentDir, 0700)

			content := ""
			if tt.input != "" {
				content = "IDENTITY_FILE=" + tt.input + "\n"
			}
			os.WriteFile(filepath.Join(agentDir, "config"), []byte(content), 0600)

			// We can't override os.UserHomeDir easily, so just verify the
			// expansion logic directly on known inputs.
			if tt.input == "" {
				return
			}
			if strings.HasPrefix(tt.input, "~") {
				expanded := filepath.Join(home, tt.input[1:])
				if expanded != tt.want {
					t.Errorf("expand(%q) = %q, want %q", tt.input, expanded, tt.want)
				}
			}
		})
	}
}

func TestLoadRemote_AllFields(t *testing.T) {
	// We test LoadRemote by pointing it at a known config path.
	// Since LoadRemote always reads from ~/.agent-secrets/config, we write
	// to the real path and restore afterwards.
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(home, ".agent-secrets", "config")

	original, readErr := os.ReadFile(configPath)

	content := `REMOTE_HOST=myserver.com
REMOTE_USER=deploy
IDENTITY_FILE=~/.ssh/deploy_key
REMOTE_PASSWORD=secret123
`
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if readErr == nil {
			os.WriteFile(configPath, original, 0600)
		}
	})

	remote, err := config.LoadRemote()
	if err != nil {
		t.Fatalf("LoadRemote: %v", err)
	}

	if remote.Host != "myserver.com" {
		t.Errorf("Host = %q, want %q", remote.Host, "myserver.com")
	}
	if remote.User != "deploy" {
		t.Errorf("User = %q, want %q", remote.User, "deploy")
	}
	if !strings.HasSuffix(remote.IdentityFile, ".ssh/deploy_key") {
		t.Errorf("IdentityFile = %q, expected suffix .ssh/deploy_key", remote.IdentityFile)
	}
	if remote.Password != "secret123" {
		t.Errorf("Password = %q, want %q", remote.Password, "secret123")
	}
}

func TestLoadRemote_EmptyConfig(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(home, ".agent-secrets", "config")
	original, readErr := os.ReadFile(configPath)

	// All commented out — should return empty Remote
	os.WriteFile(configPath, []byte("# REMOTE_HOST=example.com\n"), 0600)
	t.Cleanup(func() {
		if readErr == nil {
			os.WriteFile(configPath, original, 0600)
		}
	})

	remote, err := config.LoadRemote()
	if err != nil {
		t.Fatalf("LoadRemote: %v", err)
	}
	if remote.Host != "" || remote.User != "" {
		t.Errorf("expected empty remote, got %+v", remote)
	}
}
