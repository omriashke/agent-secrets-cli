package ssh

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/term"
)

// Dial opens an SSH connection to user@host:22.
// Auth is attempted in order: SSH agent → key file → password prompt.
func Dial(user, host, identityFile string) (*ssh.Client, error) {
	authMethods := buildAuthMethods(identityFile)

	cfg := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: known_hosts support
	}

	addr := net.JoinHostPort(host, "22")
	client, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		return nil, fmt.Errorf("SSH connection to %s@%s failed: %w", user, host, err)
	}
	return client, nil
}

func buildAuthMethods(identityFile string) []ssh.AuthMethod {
	var methods []ssh.AuthMethod

	// 1. SSH agent
	if sock := os.Getenv("SSH_AUTH_SOCK"); sock != "" {
		if conn, err := net.Dial("unix", sock); err == nil {
			methods = append(methods, ssh.PublicKeysCallback(agent.NewClient(conn).Signers))
		}
	}

	// 2. Key file (explicit or defaults)
	keyPaths := defaultKeyPaths()
	if identityFile != "" {
		keyPaths = []string{identityFile}
	}
	for _, kp := range keyPaths {
		if signer, err := loadKey(kp); err == nil {
			methods = append(methods, ssh.PublicKeys(signer))
		}
	}

	// 3. Password fallback
	methods = append(methods, ssh.PasswordCallback(func() (string, error) {
		fmt.Print("SSH password: ")
		pw, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		return string(pw), err
	}))

	return methods
}

func defaultKeyPaths() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	return []string{
		filepath.Join(home, ".ssh", "id_ed25519"),
		filepath.Join(home, ".ssh", "id_rsa"),
	}
}

func loadKey(path string) (ssh.Signer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ssh.ParsePrivateKey(data)
}
