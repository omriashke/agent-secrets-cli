package ssh

import (
	"fmt"
	"io"
	"os"

	"github.com/pkg/sftp"
	gossh "golang.org/x/crypto/ssh"
)

// Pull downloads secrets.def and .secrets from the remote and overwrites local copies.
// The local CLI will auto-sync its own DB on next use.
func Pull(client *gossh.Client, localDefPath, localSecretsPath string) error {
	remoteDir, err := remoteAgentSecretsDir(client)
	if err != nil {
		return err
	}

	sc, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("cannot open SFTP session: %w", err)
	}
	defer sc.Close()

	if err := downloadFile(sc, remoteDir+"/secrets.def", localDefPath); err != nil {
		return err
	}
	return downloadFile(sc, remoteDir+"/.secrets", localSecretsPath)
}

// DownloadToTemp downloads the remote secrets.def and .secrets to temporary
// files and returns their paths. The caller is responsible for cleanup.
func DownloadToTemp(client *gossh.Client) (tmpDefPath, tmpSecretsPath string, err error) {
	remoteDir, err := remoteAgentSecretsDir(client)
	if err != nil {
		return "", "", err
	}

	sc, err := sftp.NewClient(client)
	if err != nil {
		return "", "", fmt.Errorf("cannot open SFTP session: %w", err)
	}
	defer sc.Close()

	tmpDef, err := os.CreateTemp("", "agent-secrets-remote-def-*")
	if err != nil {
		return "", "", fmt.Errorf("cannot create temp file: %w", err)
	}
	tmpDef.Close()

	tmpSecrets, err := os.CreateTemp("", "agent-secrets-remote-secrets-*")
	if err != nil {
		os.Remove(tmpDef.Name())
		return "", "", fmt.Errorf("cannot create temp file: %w", err)
	}
	tmpSecrets.Close()

	if err := downloadFile(sc, remoteDir+"/secrets.def", tmpDef.Name()); err != nil {
		os.Remove(tmpDef.Name())
		os.Remove(tmpSecrets.Name())
		return "", "", err
	}
	if err := downloadFile(sc, remoteDir+"/.secrets", tmpSecrets.Name()); err != nil {
		os.Remove(tmpDef.Name())
		os.Remove(tmpSecrets.Name())
		return "", "", err
	}

	return tmpDef.Name(), tmpSecrets.Name(), nil
}

func downloadFile(sc *sftp.Client, remotePath, localPath string) error {
	src, err := sc.Open(remotePath)
	if err != nil {
		return fmt.Errorf("cannot open remote file %s: %w", remotePath, err)
	}
	defer src.Close()

	dst, err := os.OpenFile(localPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("cannot open local file %s for writing: %w", localPath, err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("download of %s failed: %w", remotePath, err)
	}
	return nil
}
