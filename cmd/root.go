package cmd

import (
	"fmt"
	"os"

	"github.com/omriashke/agent-secrets-cli/internal/instructions"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "agent-secrets",
	Short:         "Declarative secrets manager for AI agents",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		showInstructions, _ := cmd.Flags().GetBool("instructions")
		if showInstructions {
			fmt.Print(instructions.Content)
			return nil
		}
		return cmd.Help()
	},
}

func init() {
	rootCmd.PersistentFlags().StringP("identity", "i", "", "Path to SSH private key")
	rootCmd.Flags().Bool("instructions", false, "Print agent instructions and exit")

	// Print instructions on any usage error.
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return nil
	}

	rootCmd.SetHelpTemplate(instructions.Content + "\n")

	rootCmd.AddCommand(listCmd, queryCmd, pushCmd, pullCmd, skillCmd, versionCmd, editCmd, addCmd, deleteCmd, upgradeCmd)
}

// Execute is the entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, instructions.Content)
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
