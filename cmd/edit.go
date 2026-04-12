package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/omriashke/agent-secrets-cli/internal/config"
	"github.com/omriashke/agent-secrets-cli/internal/fileutil"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit [NAME]",
	Short: "Edit a secret or open secrets in $EDITOR",
	Long: `Without arguments, opens secrets.def and .secrets in $EDITOR (falls back to vi).

With a NAME argument, updates the specified secret inline using --description
and/or --value flags — useful for scripts and agents.

Examples:
  agent-secrets edit
  agent-secrets edit OPENAI_API_KEY --value "sk-new..."
  agent-secrets edit OPENAI_API_KEY --description "Updated description" --value "sk-new..."`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		description, _ := cmd.Flags().GetString("description")
		value, _ := cmd.Flags().GetString("value")

		if len(args) == 0 && description == "" && value == "" {
			return openEditor()
		}

		if len(args) == 0 {
			return fmt.Errorf("provide a secret NAME when using --description or --value")
		}

		name := args[0]
		if description == "" && value == "" {
			return fmt.Errorf("provide --description and/or --value to update %s", name)
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
		if _, exists := defs[name]; !exists {
			return fmt.Errorf("secret %q not found — use 'agent-secrets add' to create it", name)
		}

		if description != "" {
			if err := fileutil.SetKey(defPath, name, description); err != nil {
				return fmt.Errorf("cannot update description: %w", err)
			}
		}
		if value != "" {
			if err := fileutil.SetKey(secretsPath, name, value); err != nil {
				return fmt.Errorf("cannot update value: %w", err)
			}
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Secret %s updated.\n", name)
		return nil
	},
}

func openEditor() error {
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
}

func init() {
	editCmd.Flags().String("description", "", "New description for the secret")
	editCmd.Flags().String("value", "", "New value for the secret")
}
