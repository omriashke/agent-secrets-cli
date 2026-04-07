package ssh

import (
	"fmt"
	"io"
	"os"

	"github.com/pkg/sftp"
	gossh "golang.org/x/crypto/ssh"
)

// Pull downloads the remote DB file and overwrites the local one.
func Pull(client *gossh.Client, localDBPath string) error {
	remoteDir, err := remoteAgentSecretsDir(client)
	if err != nil {
		return err
	}

	sc, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("cannot open SFTP session: %w", err)
	}
	defer sc.Close()

	src, err := sc.Open(remoteDir + "/db")
	if err != nil {
		return fmt.Errorf("cannot open remote db: %w", err)
	}
	defer src.Close()

	dst, err := os.OpenFile(localDBPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("cannot open local db for writing: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	return nil
}
