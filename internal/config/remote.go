package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

// Remote holds the persisted remote connection settings from ~/.agent-secrets/config.
type Remote struct {
	Host         string
	User         string
	IdentityFile string
	Password     string
}

// LoadRemote reads ~/.agent-secrets/config and returns the Remote settings.
// Missing or commented-out keys are returned as empty strings.
func LoadRemote() (*Remote, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	m, err := godotenv.Read(path)
	if err != nil {
		// Config file may be all comments — treat as empty.
		if os.IsNotExist(err) {
			return &Remote{}, nil
		}
		return &Remote{}, nil
	}

	return &Remote{
		Host:         m["REMOTE_HOST"],
		User:         m["REMOTE_USER"],
		IdentityFile: expandHome(m["IDENTITY_FILE"]),
		Password:     m["REMOTE_PASSWORD"],
	}, nil
}

// expandHome replaces a leading ~ with the user's home directory.
func expandHome(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[1:])
}
