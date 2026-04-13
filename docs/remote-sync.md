# Remote sync

Push and pull secrets between your local machine and remote servers over SSH.

---

## Push

Upload your local secrets to a remote server:

```bash
agent-secrets push [user@host] [-i identity_file] [--yes]
```

### What happens

1. Connects to the remote via SSH
2. Checks if `agent-secrets` is installed on the remote — offers to install it if missing
3. Downloads the remote's current `secrets.def` and `.secrets`
4. Computes a diff and displays what will change (added, removed, changed secrets with masked values)
5. Asks for confirmation
6. Uploads local files to the remote

### Example

```bash
$ agent-secrets push deploy@myserver.com

Connecting to deploy@myserver.com...

The following changes will be applied (push):

  + NEW_API_KEY                   New API key for service X
    value: sk-a****789z
  - OLD_TOKEN                     Deprecated auth token
  ~ DATABASE_URL                  Postgres connection string
    value: post****5432 → post****3306

Proceed with push? [y/N] y
Secrets pushed to deploy@myserver.com successfully.
```

### Skip confirmation

```bash
agent-secrets push deploy@myserver.com --yes
```

---

## Pull

Download secrets from a remote server to your local machine:

```bash
agent-secrets pull [user@host] [-i identity_file] [--yes]
```

### What happens

1. Connects to the remote via SSH
2. Downloads the remote's `secrets.def` and `.secrets` to temp files
3. Computes a diff against your local files
4. Displays what will change and asks for confirmation
5. Overwrites local files with the remote versions

### Example

```bash
$ agent-secrets pull deploy@myserver.com

Connecting to deploy@myserver.com...

The following changes will be applied (pull):

  ~ OPENAI_API_KEY                OpenAI API key for GPT-4 calls
    value: sk-a****123z → sk-n****456z

Proceed with pull? [y/N] y
Secrets pulled from deploy@myserver.com successfully.
```

---

## Diff format

The diff shows three types of changes:

| Symbol | Meaning |
|---|---|
| `+` | Secret exists locally but not on remote (will be added) |
| `-` | Secret exists on remote but not locally (will be removed) |
| `~` | Secret exists on both sides but description or value differs |

Values are always masked — only the first 4 and last 4 characters are shown. Short values are fully masked.

---

## SSH authentication

Auth is attempted in this order:

1. **SSH agent** — if `SSH_AUTH_SOCK` is set
2. **Key file** — from `-i` flag, `IDENTITY_FILE` in config, or defaults (`~/.ssh/id_ed25519`, `~/.ssh/id_rsa`)
3. **Password prompt** — interactive fallback

---

## Upgrade remote CLI

Upgrade `agent-secrets` on a remote server to the latest version:

```bash
agent-secrets upgrade [user@host] [--version <ver>]
```

### Examples

```bash
# Upgrade to latest
agent-secrets upgrade deploy@myserver.com

# Pin a specific version
agent-secrets upgrade deploy@myserver.com --version 1.1.0

# Use config defaults
agent-secrets upgrade
```

The command prints the current remote version, performs the upgrade, and verifies the new version.
