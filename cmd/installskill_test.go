package cmd

import (
	"strings"
	"testing"

	"github.com/omriashke/agent-secrets-cli/internal/skill"
)

func TestSkill_ContentNotEmpty(t *testing.T) {
	if strings.TrimSpace(skill.Content) == "" {
		t.Fatal("embedded SKILL.md content is empty")
	}
}

func TestSkill_HasFrontmatter(t *testing.T) {
	if !strings.HasPrefix(strings.TrimSpace(skill.Content), "---") {
		t.Error("SKILL.md is missing YAML frontmatter")
	}
	if !strings.Contains(skill.Content, "name: agent-secrets") {
		t.Error("SKILL.md frontmatter missing 'name: agent-secrets'")
	}
	if !strings.Contains(skill.Content, "description:") {
		t.Error("SKILL.md frontmatter missing 'description'")
	}
}

func TestSkill_HasRequiredCommands(t *testing.T) {
	for _, cmd := range []string{"query", "list", "skill", "add", "edit", "delete", "upgrade"} {
		if !strings.Contains(skill.Content, cmd) {
			t.Errorf("SKILL.md missing command %q", cmd)
		}
	}
}

func TestSkill_MentionsMultipleAgents(t *testing.T) {
	for _, agent := range []string{"AGENTS.md", "CLAUDE.md", "Windsurf", "Cursor"} {
		if !strings.Contains(skill.Content, agent) {
			t.Errorf("SKILL.md does not mention agent framework %q", agent)
		}
	}
}
