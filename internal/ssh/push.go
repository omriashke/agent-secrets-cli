package ssh

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pkg/sftp"
	gossh "golang.org/x/crypto/ssh"
)

const remoteInstallScript = "https://raw.githubusercontent.com/omriashke/agent-secrets-cli/main/install.sh"

// remoteAgentSecretsDir resolves the absolute path to ~/.agent-secrets/ on the
// remote by asking the shell — SFTP does not expand tilde.
func remoteAgentSecretsDir(client *gossh.Client) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("cannot open SSH session: %w", err)
	}
	defer session.Close()

	var out bytes.Buffer
	session.Stdout = &out
	if err := session.Run("echo $HOME"); err != nil {
		return "", fmt.Errorf("cannot resolve remote home: %w", err)
	}
	home := strings.TrimSpace(out.String())
	if home == "" {
		return "", fmt.Errorf("remote $HOME is empty")
	}
	return home + "/.agent-secrets", nil
}

// Push uploads the local DB file to the remote server.
// If the CLI is not installed on the remote, it prompts the user for permission
// to install it via the install script.
func Push(client *gossh.Client, localDBPath string, stdin io.Reader, stdout io.Writer) error {
	if err := ensureCLI(client, stdin, stdout); err != nil {
		return err
	}
	remoteDir, err := remoteAgentSecretsDir(client)
	if err != nil {
		return err
	}
	return uploadFile(client, localDBPath, remoteDir+"/db")
}

func ensureCLI(client *gossh.Client, stdin io.Reader, stdout io.Writer) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("cannot open SSH session: %w", err)
	}
	defer session.Close()

	var out bytes.Buffer
	session.Stdout = &out
	_ = session.Run("which agent-secrets")

	if strings.TrimSpace(out.String()) != "" {
		return nil
	}

	fmt.Fprintln(stdout, "agent-secrets is not installed on the remote server.")
	fmt.Fprint(stdout, "Install it now? [y/N] ")

	scanner := bufio.NewScanner(stdin)
	scanner.Scan()
	answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
	if answer != "y" && answer != "yes" {
		return fmt.Errorf("aborted: CLI not installed on remote")
	}

	return installCLI(client, stdout)
}

func installCLI(client *gossh.Client, stdout io.Writer) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("cannot open SSH session: %w", err)
	}
	defer session.Close()

	session.Stdout = stdout
	session.Stderr = stdout

	cmd := fmt.Sprintf("curl -sSL %s | sh", remoteInstallScript)
	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("CLI installation failed: %w", err)
	}
	fmt.Fprintln(stdout, "agent-secrets installed successfully.")
	return nil
}

func uploadFile(client *gossh.Client, localPath, remotePath string) error {
	sc, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("cannot open SFTP session: %w", err)
	}
	defer sc.Close()

	remoteDir := remotePath[:strings.LastIndex(remotePath, "/")]
	if err := sc.MkdirAll(remoteDir); err != nil {
		return fmt.Errorf("cannot create remote directory %s: %w", remoteDir, err)
	}

	src, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("cannot open local db: %w", err)
	}
	defer src.Close()

	dst, err := sc.Create(remotePath)
	if err != nil {
		return fmt.Errorf("cannot create remote file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	return nil
}
