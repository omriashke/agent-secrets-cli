---
name: agent-secrets
description: >-
  Use the agent-secrets CLI to find, add, edit, and delete secrets by meaning.
  Covers listing available secrets, querying by description, managing secrets,
  and understanding the two-file source format. Use when you need an API key,
  password, token, or any secret value, or when the user mentions
  agent-secrets, secrets.def, or ~/.agent-secrets/.
---

# agent-secrets

Secrets are stored in `~/.agent-secrets/` and retrieved by describing what you need in plain language.

## Find a secret

```bash
# See all available secrets and their descriptions
agent-secrets list

# Retrieve a secret by meaning — returns the value on stdout
agent-secrets query "OpenAI API key for GPT-4"
agent-secrets query "Postgres password for production"
agent-secrets query "Stripe secret key"
```

## Add, edit, and delete secrets

```bash
# Add a new secret
agent-secrets add MY_SECRET --description "What this secret is for" --value "the-actual-value"

# Update an existing secret's value or description
agent-secrets edit MY_SECRET --value "new-value"
agent-secrets edit MY_SECRET --description "Updated description" --value "new-value"

# Delete a secret (asks for confirmation)
agent-secrets delete MY_SECRET
agent-secrets delete MY_SECRET --yes   # skip confirmation
```

## Source files

| File | Contains |
|---|---|
| `~/.agent-secrets/secrets.def` | Variable names + human descriptions |
| `~/.agent-secrets/.secrets` | Variable names + actual values (private) |

Both use dotenv format: `VARIABLE_NAME="value"`

## If the CLI is not installed

```bash
curl -sSL https://raw.githubusercontent.com/omriashke/agent-secrets-cli/main/install.sh | sh
```

## Push / pull secrets to a remote server

```bash
agent-secrets push user@host    # upload secrets + install CLI on remote
agent-secrets pull user@host    # download secrets from remote
```

Both commands show a diff of what will change (descriptions and masked values) and ask for confirmation before proceeding. Use `--yes` to skip the prompt.

## Upgrade agent-secrets on a remote server

```bash
agent-secrets upgrade user@host                  # upgrade to latest
agent-secrets upgrade user@host --version 1.3.0  # pin exact version
```

Set defaults in `~/.agent-secrets/config` to avoid typing `user@host` every time:

```env
REMOTE_HOST=myserver.com
REMOTE_USER=deploy
IDENTITY_FILE=~/.ssh/my_key
```

## Install this skill for your agent

```bash
# Universal (Codex, Copilot, Jules, Windsurf, and more)
agent-secrets skill > AGENTS.md

# Claude Code
agent-secrets skill > CLAUDE.md

# Windsurf
agent-secrets skill > .windsurfrules

# Cursor
agent-secrets skill > ~/.cursor/skills/agent-secrets/AGENT-SECRETS.md

# GitHub Copilot
agent-secrets skill > .github/copilot-instructions.md
```

## On any error

Run `agent-secrets --instructions` to print the full usage guide.
