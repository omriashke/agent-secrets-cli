package ssh

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	gossh "golang.org/x/crypto/ssh"
)

// RemoteVersion runs `agent-secrets version` on the remote and returns the
// version string. Returns an error if the CLI is not installed.
func RemoteVersion(client *gossh.Client) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("cannot open SSH session: %w", err)
	}
	defer session.Close()

	var out bytes.Buffer
	session.Stdout = &out
	if err := session.Run("agent-secrets version"); err != nil {
		return "", fmt.Errorf("agent-secrets not installed on remote: %w", err)
	}

	version := strings.TrimSpace(out.String())
	if version == "" {
		return "", fmt.Errorf("agent-secrets returned empty version")
	}
	return version, nil
}

// Upgrade installs or upgrades agent-secrets on the remote.
// If version is empty, the latest release is installed.
// If version is set (e.g. "1.3.0"), that exact version is installed.
func Upgrade(client *gossh.Client, version string, stdout io.Writer) error {
	if version == "" {
		return upgradeLatest(client, stdout)
	}
	return upgradeExact(client, version, stdout)
}

func upgradeLatest(client *gossh.Client, stdout io.Writer) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("cannot open SSH session: %w", err)
	}
	defer session.Close()

	session.Stdout = stdout
	session.Stderr = stdout

	fmt.Fprintln(stdout, "Upgrading to latest version...")
	cmd := fmt.Sprintf("curl -sSL %s | sh", remoteInstallScript)
	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("upgrade failed: %w", err)
	}
	return nil
}

func upgradeExact(client *gossh.Client, version string, stdout io.Writer) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("cannot open SSH session: %w", err)
	}
	defer session.Close()

	session.Stdout = stdout
	session.Stderr = stdout

	version = strings.TrimPrefix(version, "v")

	fmt.Fprintf(stdout, "Upgrading to version %s...\n", version)

	// Inline script that downloads the exact version binary.
	script := fmt.Sprintf(`
set -e
REPO="omriashke/agent-secrets-cli"
BIN_NAME="agent-secrets"
VERSION="v%s"
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac
FILENAME="${BIN_NAME}_${OS}_${ARCH}"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"
echo "Downloading ${BIN_NAME} ${VERSION} (${OS}/${ARCH})..."
curl -sSL "$URL" -o "/tmp/${BIN_NAME}"
chmod +x "/tmp/${BIN_NAME}"
if [ -w "/usr/local/bin" ]; then
  INSTALL_DIR="/usr/local/bin"
else
  INSTALL_DIR="$HOME/.local/bin"
  mkdir -p "$INSTALL_DIR"
fi
mv "/tmp/${BIN_NAME}" "${INSTALL_DIR}/${BIN_NAME}"
echo "${BIN_NAME} ${VERSION} installed to ${INSTALL_DIR}/${BIN_NAME}"
`, version)

	if err := session.Run(script); err != nil {
		return fmt.Errorf("upgrade to version %s failed: %w", version, err)
	}
	return nil
}
