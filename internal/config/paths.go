package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	DirName     = ".agent-secrets"
	DefFile     = "secrets.def"
	SecretsFile = ".secrets"
	DBFile      = "db"
	ConfigFile  = "config"
)

// Dir returns the path to ~/.agent-secrets/ and ensures it exists.
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot resolve home directory: %w", err)
	}
	dir := filepath.Join(home, DirName)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("cannot create %s: %w", dir, err)
	}
	return dir, nil
}

// DefPath returns the full path to secrets.def, scaffolding it if absent.
func DefPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	p := filepath.Join(dir, DefFile)
	if err := scaffoldIfMissing(p, defTemplate); err != nil {
		return "", err
	}
	return p, nil
}

// SecretsPath returns the full path to .secrets, scaffolding it if absent.
func SecretsPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	p := filepath.Join(dir, SecretsFile)
	if err := scaffoldIfMissing(p, secretsTemplate); err != nil {
		return "", err
	}
	return p, nil
}

// DBPath returns the full path to the SQLite database file.
func DBPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, DBFile), nil
}

// ConfigPath returns the full path to the config file, scaffolding it if absent.
func ConfigPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	p := filepath.Join(dir, ConfigFile)
	if err := scaffoldIfMissing(p, configTemplate); err != nil {
		return "", err
	}
	return p, nil
}

func scaffoldIfMissing(path, content string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.WriteFile(path, []byte(content), 0600)
	}
	return nil
}

const defTemplate = `# secrets.def — define your secrets here.
# Format: VARIABLE_NAME="human-readable description for agents"
# Example:
# OPENAI_API_KEY="OpenAI API key for GPT-4 calls"
# DATABASE_PASSWORD="Postgres password for the production database"
`

const secretsTemplate = `# .secrets — actual secret values. Keep this file private, never commit it.
# Format: VARIABLE_NAME=actual_value
# Example:
# OPENAI_API_KEY=sk-abc123...
# DATABASE_PASSWORD=hunter2
`

const configTemplate = `# agent-secrets config
# Default remote host for push/pull commands.
# When set, running "agent-secrets push" or "agent-secrets pull" without
# arguments will use these values. CLI flags always override.

# REMOTE_HOST=myserver.com
# REMOTE_USER=deploy

# SSH authentication — choose one:
# IDENTITY_FILE=~/.ssh/id_ed25519
# REMOTE_PASSWORD=mypassword
`
