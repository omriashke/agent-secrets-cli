package cmd

import (
	"fmt"
	"os"

	"github.com/omriashke/agent-secrets-cli/internal/config"
	internalssh "github.com/omriashke/agent-secrets-cli/internal/ssh"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade [user@host]",
	Short: "Upgrade agent-secrets on a remote server",
	Long: `Upgrade the agent-secrets CLI on a remote server to the latest version
(or a specific version with --version).

Examples:
  agent-secrets upgrade deploy@myserver.com
  agent-secrets upgrade deploy@myserver.com --version 1.3.0
  agent-secrets upgrade                      # uses config defaults`,
	Args: cobra.MaximumNArgs(1),
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

		targetVersion, _ := cmd.Flags().GetString("version")

		fmt.Printf("Connecting to %s@%s...\n", user, host)
		client, err := internalssh.Dial(user, host, identity)
		if err != nil {
			return err
		}
		defer client.Close()

		currentVersion, err := internalssh.RemoteVersion(client)
		if err != nil {
			fmt.Println("agent-secrets is not installed on the remote — installing...")
		} else {
			fmt.Printf("Current remote version: %s\n", currentVersion)
		}

		if err := internalssh.Upgrade(client, targetVersion, os.Stdout); err != nil {
			return err
		}

		newVersion, err := internalssh.RemoteVersion(client)
		if err != nil {
			return fmt.Errorf("upgrade succeeded but cannot verify version: %w", err)
		}

		fmt.Printf("Remote is now running agent-secrets %s\n", newVersion)
		return nil
	},
}

func init() {
	upgradeCmd.Flags().String("version", "", "Specific version to install (e.g. 1.3.0) — defaults to latest")
}
