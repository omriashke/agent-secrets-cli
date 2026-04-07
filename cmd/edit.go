package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/omriashke/agent-secrets-cli/internal/config"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open secrets.def and .secrets in $EDITOR",
	Long: `Opens both ~/.agent-secrets/secrets.def and ~/.agent-secrets/.secrets
in your $EDITOR (falls back to vi if unset).

Edit secrets.def to add or update descriptions, and .secrets to set values.
The DB is re-synced automatically the next time you run any command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		defPath, err := config.DefPath()
		if err != nil {
			return err
		}
		secretsPath, err := config.SecretsPath()
		if err != nil {
			return err
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}

		which := exec.Command("which", editor)
		if err := which.Run(); err != nil {
			return fmt.Errorf("editor %q not found — set $EDITOR to a valid editor", editor)
		}

		c := exec.Command(editor, defPath, secretsPath)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}
