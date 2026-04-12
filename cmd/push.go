package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/omriashke/agent-secrets-cli/internal/config"
	"github.com/omriashke/agent-secrets-cli/internal/db"
	"github.com/omriashke/agent-secrets-cli/internal/diff"
	internalssh "github.com/omriashke/agent-secrets-cli/internal/ssh"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push [user@host]",
	Short: "Push secrets to a remote server",
	Long: `Push local secrets to a remote server over SSH. Before overwriting,
shows a diff of descriptions and values that will change and asks for
confirmation. Use --yes to skip the prompt.`,
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

		if err := internalssh.EnsureCLI(client, os.Stdin, os.Stdout); err != nil {
			return err
		}

		tmpDef, tmpSecrets, err := internalssh.DownloadToTemp(client)
		if err != nil {
			fmt.Println("Remote has no existing secrets — this will be a fresh push.")
		} else {
			defer os.Remove(tmpDef)
			defer os.Remove(tmpSecrets)

			changes, err := diff.ComputeChanges(defPath, secretsPath, tmpDef, tmpSecrets)
			if err != nil {
				return fmt.Errorf("cannot compute diff: %w", err)
			}

			if len(changes) == 0 {
				fmt.Println("No differences found — push is a no-op.")
				return nil
			}

			diff.PrintChanges(os.Stdout, changes, "push")

			yes, _ := cmd.Flags().GetBool("yes")
			if !yes {
				if !confirmPrompt("Proceed with push?") {
					fmt.Println("Aborted.")
					return nil
				}
			}
		}

		if err := internalssh.PushFiles(client, defPath, secretsPath); err != nil {
			return err
		}

		fmt.Printf("Secrets pushed to %s@%s successfully.\n", user, host)
		return nil
	},
}

var pullCmd = &cobra.Command{
	Use:   "pull [user@host]",
	Short: "Pull secrets from a remote server",
	Long: `Pull secrets from a remote server over SSH. Before overwriting local
files, shows a diff of descriptions and values that will change and asks
for confirmation. Use --yes to skip the prompt.`,
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

		tmpDef, tmpSecrets, err := internalssh.DownloadToTemp(client)
		if err != nil {
			return fmt.Errorf("cannot download remote secrets: %w", err)
		}
		defer os.Remove(tmpDef)
		defer os.Remove(tmpSecrets)

		// Diff: remote (incoming) vs local (current).
		// ComputeChanges treats first pair as "new" and second as "old" (destination).
		changes, err := diff.ComputeChanges(tmpDef, tmpSecrets, defPath, secretsPath)
		if err != nil {
			return fmt.Errorf("cannot compute diff: %w", err)
		}

		if len(changes) == 0 {
			fmt.Println("No differences found — pull is a no-op.")
			return nil
		}

		diff.PrintChanges(os.Stdout, changes, "pull")

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			if !confirmPrompt("Proceed with pull?") {
				fmt.Println("Aborted.")
				return nil
			}
		}

		if err := internalssh.Pull(client, defPath, secretsPath); err != nil {
			return err
		}

		fmt.Printf("Secrets pulled from %s@%s successfully.\n", user, host)
		return nil
	},
}

func confirmPrompt(question string) bool {
	fmt.Printf("%s [y/N] ", question)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
	return answer == "y" || answer == "yes"
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

func init() {
	pushCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	pullCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
}
