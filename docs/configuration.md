# Configuration

## Config file

Set defaults in `~/.agent-secrets/config` so you don't need to type `user@host` on every push, pull, or upgrade:

```env
REMOTE_HOST=myserver.com
REMOTE_USER=deploy
IDENTITY_FILE=~/.ssh/my_key
```

With this config, all remote commands work without arguments:

```bash
agent-secrets push       # pushes to deploy@myserver.com
agent-secrets pull       # pulls from deploy@myserver.com
agent-secrets upgrade    # upgrades deploy@myserver.com
```

CLI arguments always override config values.

---

## Config options

| Variable | Description |
|---|---|
| `REMOTE_HOST` | Hostname or IP of the remote server |
| `REMOTE_USER` | SSH username |
| `IDENTITY_FILE` | Path to SSH private key (supports `~` expansion) |
| `REMOTE_PASSWORD` | SSH password (not recommended — use key-based auth) |

---

## Global flags

These flags are available on all commands:

| Flag | Short | Description |
|---|---|---|
| `--identity` | `-i` | Path to SSH private key (overrides `IDENTITY_FILE` in config) |
| `--instructions` | | Print the full usage guide and exit |

---

## Directory layout

Everything lives under `~/.agent-secrets/`:

```
~/.agent-secrets/
├── secrets.def     # Variable names + descriptions (safe to commit to dotfiles)
├── .secrets        # Variable names + actual values (never commit)
├── config          # Remote connection defaults
└── db              # SQLite database (auto-generated, do not edit)
```

The `db` file is rebuilt automatically whenever `secrets.def` or `.secrets` changes. You never need to touch it directly.

---

## Environment variables

| Variable | Description |
|---|---|
| `EDITOR` | Editor used by `agent-secrets edit` (defaults to `vi`) |
| `SSH_AUTH_SOCK` | SSH agent socket — used for key-based auth when available |
