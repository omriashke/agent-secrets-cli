package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/omriashke/agent-secrets-cli/internal/config"
	"github.com/omriashke/agent-secrets-cli/internal/db"
	internalssh "github.com/omriashke/agent-secrets-cli/internal/ssh"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push [user@host]",
	Short: "Push secrets to a remote server",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		remote, err := config.LoadRemote()
		if err != nil {
			return err
		}

		user, host, err := resolveUserHost(args, remote)
		if err != nil {
			return err
		}

		identity, _ := cmd.Flags().GetString("identity")
		if identity == "" {
			identity = remote.IdentityFile
		}

		defPath, err := config.DefPath()
		if err != nil {
			return err
		}
		secretsPath, err := config.SecretsPath()
		if err != nil {
			return err
		}
		dbPath, err := config.DBPath()
		if err != nil {
			return err
		}

		database, err := db.Open(dbPath)
		if err != nil {
			return err
		}
		defer database.Close()

		if err := db.AutoSync(database, defPath, secretsPath); err != nil {
			return err
		}
		database.Close()

		fmt.Printf("Connecting to %s@%s...\n", user, host)
		client, err := internalssh.Dial(user, host, identity)
		if err != nil {
			return err
		}
		defer client.Close()

		if err := internalssh.Push(client, defPath, secretsPath, os.Stdin, os.Stdout); err != nil {
			return err
		}

		fmt.Printf("Secrets pushed to %s@%s successfully.\n", user, host)
		return nil
	},
}

var pullCmd = &cobra.Command{
	Use:   "pull [user@host]",
	Short: "Pull secrets from a remote server",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		remote, err := config.LoadRemote()
		if err != nil {
			return err
		}

		user, host, err := resolveUserHost(args, remote)
		if err != nil {
			return err
		}

		identity, _ := cmd.Flags().GetString("identity")
		if identity == "" {
			identity = remote.IdentityFile
		}

		defPath, err := config.DefPath()
		if err != nil {
			return err
		}
		secretsPath, err := config.SecretsPath()
		if err != nil {
			return err
		}

		fmt.Printf("Connecting to %s@%s...\n", user, host)
		client, err := internalssh.Dial(user, host, identity)
		if err != nil {
			return err
		}
		defer client.Close()

		if err := internalssh.Pull(client, defPath, secretsPath); err != nil {
			return err
		}

		fmt.Printf("Secrets pulled from %s@%s successfully.\n", user, host)
		return nil
	},
}

// resolveUserHost picks user@host from CLI args, falling back to config.
func resolveUserHost(args []string, remote *config.Remote) (user, host string, err error) {
	if len(args) == 1 {
		return parseUserHost(args[0])
	}
	if remote.Host != "" && remote.User != "" {
		return remote.User, remote.Host, nil
	}
	if remote.Host != "" {
		return "", "", fmt.Errorf("REMOTE_HOST is set in config but REMOTE_USER is missing")
	}
	return "", "", fmt.Errorf("no remote specified — provide user@host or set REMOTE_HOST and REMOTE_USER in ~/.agent-secrets/config")
}

func parseUserHost(arg string) (user, host string, err error) {
	parts := strings.SplitN(arg, "@", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid format %q — expected user@host", arg)
	}
	return parts[0], parts[1], nil
}
