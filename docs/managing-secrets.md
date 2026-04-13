# Managing secrets

Secrets are stored in two dotenv files under `~/.agent-secrets/`:

| File | Purpose |
|---|---|
| `secrets.def` | Variable names + human-readable descriptions |
| `.secrets` | Variable names + actual secret values |

Both use the same format: `VARIABLE_NAME="value"`. The CLI keeps them in sync with a local SQLite database that powers full-text search.

---

## Adding a secret

```bash
agent-secrets add <NAME> --description "<description>" --value "<value>"
```

Both `--description` and `--value` are required. The name must not already exist.

### Examples

```bash
agent-secrets add DATABASE_URL \
  --description "Postgres connection string for production" \
  --value "postgres://user:pass@host:5432/mydb"

agent-secrets add STRIPE_SECRET \
  --description "Stripe secret key for payment processing" \
  --value "sk_live_xyz789"
```

### What happens

1. The description is appended to `~/.agent-secrets/secrets.def`
2. The value is appended to `~/.agent-secrets/.secrets`
3. The next command that reads the database (e.g. `list`, `query`) auto-syncs

### Errors

- If the name already exists: `secret "NAME" already exists — use 'agent-secrets edit' to update it`

---

## Editing a secret

### Interactive (open in editor)

```bash
agent-secrets edit
```

Opens both `secrets.def` and `.secrets` in your `$EDITOR` (falls back to `vi`). Edit the files directly, save, and quit. The database re-syncs on the next command.

### Inline (from the command line)

```bash
agent-secrets edit <NAME> [--description "<new description>"] [--value "<new value>"]
```

At least one of `--description` or `--value` is required.

### Examples

```bash
# Update just the value
agent-secrets edit DATABASE_URL --value "postgres://newuser:newpass@host:5432/mydb"

# Update just the description
agent-secrets edit DATABASE_URL --description "Postgres connection string for staging"

# Update both at once
agent-secrets edit DATABASE_URL \
  --description "Postgres connection string for staging" \
  --value "postgres://staging:pass@host:5432/staging_db"
```

### Errors

- If the name does not exist: `secret "NAME" not found — use 'agent-secrets add' to create it`
- If neither flag is provided: `provide --description and/or --value to update NAME`

---

## Deleting a secret

```bash
agent-secrets delete <NAME> [--yes]
```

By default the command shows the secret's description and asks for confirmation:

```
Delete DATABASE_URL (Postgres connection string for production)? [y/N]
```

Pass `--yes` (or `-y`) to skip the prompt — useful for scripts and agents.

### Aliases

`delete`, `rm`, and `remove` all do the same thing:

```bash
agent-secrets rm DATABASE_URL --yes
agent-secrets remove DATABASE_URL --yes
```

### What happens

1. The entry is removed from `secrets.def`
2. The entry is removed from `.secrets`
3. The database prunes the entry on the next auto-sync

### Errors

- If the name does not exist: `secret "NAME" not found`

---

## Listing secrets

```bash
agent-secrets list [--json]
```

Prints all secret names and descriptions:

```
DATABASE_URL                   Postgres connection string for production
OPENAI_API_KEY                 OpenAI API key for GPT-4 calls
STRIPE_SECRET                  Stripe secret key for payment processing
```

With `--json`:

```json
[
  {"name":"DATABASE_URL","description":"Postgres connection string for production"},
  {"name":"OPENAI_API_KEY","description":"OpenAI API key for GPT-4 calls"},
  {"name":"STRIPE_SECRET","description":"Stripe secret key for payment processing"}
]
```

---

## Querying by meaning

```bash
agent-secrets query "<description>" [--value-only] [--json] [--top N]
```

The search runs against descriptions using full-text search — agents don't need to know variable names, just what the secret is for.

By default, the query automatically returns all closely-ranked matches. A precise query returns one result; a vague query returns several.

```bash
# Precise query — one result
agent-secrets query "Stripe payment key"
# name:        STRIPE_SECRET
# description: Stripe secret key for payment processing
# value:       sk_live_xyz789

# Vague query — multiple close matches returned automatically
agent-secrets query "API key"
# [1]
# name:        OPENAI_API_KEY
# description: OpenAI API key for GPT-4 calls
# value:       sk-abc123...
#
# [2]
# name:        STRIPE_SECRET
# description: Stripe secret key for payment processing
# value:       sk_live_xyz789

agent-secrets query "payment processing" --value-only
# sk_live_xyz789
```

### Forcing exact result count

Use `--top N` to override auto-detection and force exactly N results:

```bash
agent-secrets query "API key" --top 1
# name:        OPENAI_API_KEY
# description: OpenAI API key for GPT-4 calls
# value:       sk-abc123...

agent-secrets query "API key" --top 3
# [1]
# name:        OPENAI_API_KEY
# ...
# [2]
# name:        STRIPE_SECRET
# ...
# [3]
# name:        TRELLO_API_KEY
# ...
```

Works with `--json` (returns an array) and `--value-only` (prints one value per line).

---

## File format reference

Both files use standard dotenv format. Comments (lines starting with `#`) are preserved when the CLI writes to the files.

**secrets.def:**
```bash
# API keys
OPENAI_API_KEY="OpenAI API key for GPT-4 calls"
STRIPE_SECRET="Stripe secret key for payment processing"

# Database
DATABASE_URL="Postgres connection string for production"
```

**.secrets:**
```bash
# API keys
OPENAI_API_KEY="sk-abc123..."
STRIPE_SECRET="sk_live_xyz789"

# Database
DATABASE_URL="postgres://user:pass@host:5432/mydb"
```
