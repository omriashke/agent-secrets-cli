package cmd

import (
	"fmt"

	"github.com/omriashke/agent-secrets-cli/internal/skill"
	"github.com/spf13/cobra"
)

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Print the agent skill file — pipe it to the location your agent expects",
	Long: `Prints the agent-secrets skill file to stdout.

Pipe the output to the location your agent framework expects:

  Cursor:
    agent-secrets skill > ~/.cursor/skills/agent-secrets/AGENT-SECRETS.md

  AGENTS.md (universal — works with Codex, Copilot, Jules, Windsurf, and more):
    agent-secrets skill > AGENTS.md

  Claude Code:
    agent-secrets skill > CLAUDE.md

  Windsurf:
    agent-secrets skill > .windsurfrules

  GitHub Copilot:
    agent-secrets skill > .github/copilot-instructions.md`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print(skill.Content)
		return nil
	},
}
