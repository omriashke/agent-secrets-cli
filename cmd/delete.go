package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/omriashke/agent-secrets-cli/internal/config"
	"github.com/omriashke/agent-secrets-cli/internal/fileutil"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "delete <NAME>",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete a secret",
	Long: `Delete a secret by its variable name. Removes the entry from both
secrets.def and .secrets. Asks for confirmation unless --yes is passed.

Example:
  agent-secrets delete OPENAI_API_KEY
  agent-secrets delete OPENAI_API_KEY --yes`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.TrimSpace(args[0])
		if name == "" {
			return fmt.Errorf("secret name cannot be empty")
		}

		defPath, err := config.DefPath()
		if err != nil {
			return err
		}
		secretsPath, err := config.SecretsPath()
		if err != nil {
			return err
		}

		defs, err := fileutil.ReadDotenv(defPath)
		if err != nil {
			return err
		}
		desc, exists := defs[name]
		if !exists {
			return fmt.Errorf("secret %q not found", name)
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Fprintf(cmd.OutOrStdout(), "Delete %s (%s)? [y/N] ", name, desc)
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
			if answer != "y" && answer != "yes" {
				fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
				return nil
			}
		}

		if err := fileutil.DeleteKey(defPath, name); err != nil {
			return fmt.Errorf("cannot remove from secrets.def: %w", err)
		}
		// .secrets may not have the key (orphan def) — ignore not-found.
		_ = fileutil.DeleteKey(secretsPath, name)

		fmt.Fprintf(cmd.OutOrStdout(), "Secret %s deleted.\n", name)
		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
}
