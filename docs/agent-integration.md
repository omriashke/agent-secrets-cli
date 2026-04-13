# Agent integration

## How agents use secrets

Agents find secrets by describing what they need in plain language. The CLI searches descriptions using full-text search and returns the matching value:

```bash
agent-secrets query "OpenAI API key for GPT-4"
# → sk-abc123...
```

Agents never need to know variable names — just what the secret is for.

---

## Install a skill file

Run once to teach your agent how to use `agent-secrets`:

```bash
# Universal (Codex, Copilot, Jules, Windsurf, and more)
agent-secrets skill > AGENTS.md

# Claude Code
agent-secrets skill > CLAUDE.md

# Windsurf
agent-secrets skill > .windsurfrules

# Cursor
mkdir -p ~/.cursor/skills/agent-secrets
agent-secrets skill > ~/.cursor/skills/agent-secrets/AGENT-SECRETS.md

# GitHub Copilot
agent-secrets skill > .github/copilot-instructions.md
```

Once installed, the agent will automatically use `agent-secrets` whenever it needs a secret.

---

## What the skill teaches agents

The skill file instructs agents to:

1. Run `agent-secrets list` to discover available secrets
2. Run `agent-secrets query "<description>"` to retrieve a secret by meaning
3. Run `agent-secrets add` to create new secrets when asked
4. Run `agent-secrets edit` to update existing secrets
5. Run `agent-secrets delete` to remove secrets
6. Understand the two-file source format (`secrets.def` + `.secrets`)

---

## Agent-friendly commands

All mutation commands work non-interactively:

```bash
# Add — fully non-interactive
agent-secrets add KEY --description "desc" --value "val"

# Edit — fully non-interactive
agent-secrets edit KEY --value "new-val"

# Delete — skip confirmation with --yes
agent-secrets delete KEY --yes

# Query — just the value for piping
agent-secrets query "description" --value-only
```

---

## System prompt snippet

If you can't use a skill file, add this to your agent's system prompt:

```
To access secrets, use the agent-secrets CLI:
  agent-secrets list                                    — discover available secrets
  agent-secrets query "<description>"                  — retrieve a secret by meaning
  agent-secrets query "<description>" --value-only     — just the value (for scripts)
  agent-secrets add <NAME> --description "..." --value "..."  — add a new secret
  agent-secrets edit <NAME> --value "..."              — update a secret
  agent-secrets delete <NAME> --yes                    — remove a secret
```

---

## Troubleshooting

If the CLI is not installed or an agent encounters an error:

```bash
agent-secrets --instructions
```

This prints the full usage guide with all commands and examples.
