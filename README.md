<div align="center">

<pre>
          .--.
         /.-. '----------.
         \'-' .--"--""-"-'
          '--'
</pre>

# agent-secrets

**A CLI-first secrets manager built for AI agents**

Secrets are stored locally in `~/.agent-secrets/` and agents find them by meaning using full-text search — no hardcoded variable names, no brittle lookups.

```bash
agent-secrets query "OpenAI API key for GPT-4"
# → sk-abc123...
```

</div>

---

## How it works

You define secrets in two files in your home directory:

| File | Purpose |
|---|---|
| `~/.agent-secrets/secrets.def` | Variable names + human descriptions (committed to dotfiles) |
| `~/.agent-secrets/.secrets` | Variable names + actual values (never committed) |

Both files use the same dotenv format — `VARIABLE_NAME="value"` — the only difference is what the value means.

When you run any command, the CLI automatically syncs both files into a local SQLite database with a full-text search index on the descriptions. Agents query that index by meaning and get back the secret value.

```
secrets.def          .secrets
    │                    │
    └────────┬───────────┘
             ▼
        auto-sync
             ▼
     SQLite + FTS5 DB
             ▼
    agent-secrets query
             ▼
       secret value
```

---

## Installation

```bash
curl -sSL https://raw.githubusercontent.com/omriashke/agent-secrets-cli/main/install.sh | sh
```

Installs the latest release for your OS and architecture (macOS and Linux, amd64 and arm64).
Writes to `/usr/local/bin` if writable, otherwise falls back to `~/.local/bin` — no `sudo` required.

On first run, `~/.agent-secrets/` is scaffolded automatically with template files.

---

## Quick start

**1. Edit your definitions file:**

```bash
# ~/.agent-secrets/secrets.def
OPENAI_API_KEY="OpenAI API key for GPT-4 calls"
DATABASE_PASSWORD="Postgres password for the production database"
STRIPE_SECRET="Stripe secret key for payment processing"
```

**2. Add your actual values:**

```bash
# ~/.agent-secrets/.secrets  (never commit this)
OPENAI_API_KEY=sk-abc123...
DATABASE_PASSWORD=hunter2
STRIPE_SECRET=sk_live_xyz789
```

Or open both files at once in your `$EDITOR`:

```bash
agent-secrets edit
```

**3. Use from an agent:**

```bash
# See what secrets are available
agent-secrets list

# Find a secret by meaning
agent-secrets query "OpenAI API key"
# name:        OPENAI_API_KEY
# description: OpenAI API key for GPT-4 calls
# value:       sk-abc123...

# Just the value (for scripts and pipes)
agent-secrets query "OpenAI API key" --value-only
# → sk-abc123...
```

---

## Commands

### `list`

List all secret names and their descriptions.

```bash
agent-secrets list
```

```
DATABASE_PASSWORD              Postgres password for the production database
OPENAI_API_KEY                 OpenAI API key for GPT-4 calls
STRIPE_SECRET                  Stripe secret key for payment processing
```

Add `--json` for machine-readable output:

```bash
agent-secrets list --json
# [{"name":"DATABASE_PASSWORD","description":"Postgres password..."},...]
```

### `query <description>`

Find a secret by meaning using full-text search. Prints name, description, and value by default.

```bash
agent-secrets query "payment processing"
# name:        STRIPE_SECRET
# description: Stripe secret key for payment processing
# value:       sk_live_xyz789
```

| Flag | Description |
|---|---|
| `--value-only` | Print only the raw secret value |
| `--json` | Output as JSON |

The search runs against descriptions, not variable names — so agents don't need to know the exact variable name, just what the secret is for.

### `edit`

Open `secrets.def` and `.secrets` side-by-side in your `$EDITOR` (falls back to `vi`).

```bash
agent-secrets edit
```

### `push [user@host]`

Push your secrets to a remote server over SSH. If `agent-secrets` is not installed on the remote, you will be prompted to install it automatically.

```bash
agent-secrets push deploy@myserver.com
agent-secrets push deploy@myserver.com -i ~/.ssh/my_key
```

Auth is attempted in order: SSH agent → key file → password prompt.

### `pull [user@host]`

Pull secrets from a remote server back to your local `~/.agent-secrets/`.

```bash
agent-secrets pull deploy@myserver.com
```

### `skill`

Print the agent skill file to stdout. Pipe it to wherever your agent framework reads instructions.

```bash
agent-secrets skill > AGENTS.md                                    # universal
agent-secrets skill > CLAUDE.md                                    # Claude Code
agent-secrets skill > .windsurfrules                               # Windsurf
agent-secrets skill > ~/.cursor/skills/agent-secrets/AGENT-SECRETS.md     # Cursor
agent-secrets skill > .github/copilot-instructions.md             # GitHub Copilot
```

### `version`

Print the installed version.

```bash
agent-secrets version
# 1.2.0
```

---

## Config file

Set defaults in `~/.agent-secrets/config` so you don't need to type `user@host` every time:

```env
REMOTE_HOST=myserver.com
REMOTE_USER=deploy
IDENTITY_FILE=~/.ssh/my_key   # or REMOTE_PASSWORD=mypassword
```

---

## Flags

| Flag | Short | Description |
|---|---|---|
| `--identity` | `-i` | Path to SSH private key (default: `~/.ssh/id_ed25519`) |
| `--instructions` | | Print agent instructions and exit |

---

## For AI agents

### Install a skill file (recommended)

Run this once to teach your agent how to use `agent-secrets` automatically:

```bash
# Universal (works with Codex, Copilot, Jules, Windsurf, and more)
agent-secrets skill > AGENTS.md

# Claude Code
agent-secrets skill > CLAUDE.md

# Cursor
mkdir -p ~/.cursor/skills/agent-secrets
agent-secrets skill > ~/.cursor/skills/agent-secrets/AGENT-SECRETS.md
```

Once the skill file is in place, your agent will automatically use `agent-secrets` whenever it needs a secret — no manual wiring required.

### Other agents / system prompt

Add this snippet to your agent's system prompt:

```
To access secrets, use the agent-secrets CLI:
  agent-secrets list                    — discover available secrets
  agent-secrets query "<description>"  — retrieve a secret by meaning
```

If the CLI is not installed or you are unsure how to use it, run:

```bash
agent-secrets --instructions
```

---

## Design

- **Declarative** — secrets are defined in plain text files, not through a UI or API
- **Agent-first** — descriptions are written for agents to find, not for humans to remember variable names
- **Local-first** — everything lives in `~/.agent-secrets/`, no cloud dependency
- **Auto-sync** — the database is rebuilt automatically when source files change, no manual compile step
- **Portable** — a single static Go binary, no runtime dependencies

---

## Project structure

```
agent-secrets-cli/
├── cmd/                    # Cobra commands (list, query, push, pull, edit, version, skill)
├── internal/
│   ├── config/             # ~/.agent-secrets/ path resolution + scaffolding
│   ├── parser/             # dotenv parser + merge/validation
│   ├── db/                 # SQLite schema, auto-sync, FTS5 query
│   ├── ssh/                # SSH dial, SFTP push/pull, remote install
│   ├── skill/              # go:embed SKILL.md
│   └── instructions/       # go:embed INSTRUCTIONS.md
├── .cursor/
│   └── skills/
│       └── agent-secrets/
│           └── AGENT-SECRETS.md  # Cursor agent skill — auto-loaded in this workspace
├── install.sh              # One-liner install script
└── .goreleaser.yaml        # Cross-compile + GitHub Releases config
```

---

## License

MIT
