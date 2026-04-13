# Getting started

## Install

```bash
curl -sSL https://raw.githubusercontent.com/omriashke/agent-secrets-cli/main/install.sh | sh
```

This downloads the latest release for your OS and architecture (macOS / Linux, amd64 / arm64) and places the binary in `/usr/local/bin` (or `~/.local/bin` if `/usr/local/bin` is not writable).

On first run the directory `~/.agent-secrets/` is created automatically with template files.

## Add your first secret

```bash
agent-secrets add OPENAI_API_KEY \
  --description "OpenAI API key for GPT-4 calls" \
  --value "sk-abc123..."
```

This writes to two files:

| File | What was written |
|---|---|
| `~/.agent-secrets/secrets.def` | `OPENAI_API_KEY="OpenAI API key for GPT-4 calls"` |
| `~/.agent-secrets/.secrets` | `OPENAI_API_KEY="sk-abc123..."` |

## Verify

```bash
agent-secrets list
# OPENAI_API_KEY                 OpenAI API key for GPT-4 calls

agent-secrets query "OpenAI"
# name:        OPENAI_API_KEY
# description: OpenAI API key for GPT-4 calls
# value:       sk-abc123...
```

## Teach your agent

Run once to install the skill file for your agent framework:

```bash
# Universal (Codex, Copilot, Jules, Windsurf, and more)
agent-secrets skill > AGENTS.md

# Claude Code
agent-secrets skill > CLAUDE.md

# Cursor
mkdir -p ~/.cursor/skills/agent-secrets
agent-secrets skill > ~/.cursor/skills/agent-secrets/AGENT-SECRETS.md
```

After this your agent will automatically use `agent-secrets query` whenever it needs a secret.

## Next steps

- [Managing secrets](managing-secrets.md) — add, edit, delete in detail
- [Remote sync](remote-sync.md) — push and pull secrets to servers
- [Configuration](configuration.md) — set defaults for remote connections
