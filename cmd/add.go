package cmd

import (
	"fmt"
	"strings"

	"github.com/omriashke/agent-secrets-cli/internal/config"
	"github.com/omriashke/agent-secrets-cli/internal/fileutil"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <NAME> --description <desc> --value <val>",
	Short: "Add a new secret",
	Long: `Add a new secret by providing its variable name, a human-readable
description (used by agents to find it), and the actual secret value.

Example:
  agent-secrets add OPENAI_API_KEY \
    --description "OpenAI API key for GPT-4 calls" \
    --value "sk-abc123..."`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.TrimSpace(args[0])
		if name == "" {
			return fmt.Errorf("secret name cannot be empty")
		}

		description, _ := cmd.Flags().GetString("description")
		value, _ := cmd.Flags().GetString("value")

		if description == "" {
			return fmt.Errorf("--description is required")
		}
		if value == "" {
			return fmt.Errorf("--value is required")
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
		if _, exists := defs[name]; exists {
			return fmt.Errorf("secret %q already exists — use 'agent-secrets edit' to update it", name)
		}

		if err := fileutil.SetKey(defPath, name, description); err != nil {
			return fmt.Errorf("cannot write description: %w", err)
		}
		if err := fileutil.SetKey(secretsPath, name, value); err != nil {
			return fmt.Errorf("cannot write value: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Secret %s added.\n", name)
		return nil
	},
}

func init() {
	addCmd.Flags().String("description", "", "Human-readable description for agents")
	addCmd.Flags().String("value", "", "Actual secret value")
}
