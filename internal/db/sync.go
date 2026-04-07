package db

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/omriashke/agent-secrets-cli/internal/parser"
)

const metaKeyLastSynced = "last_synced"

// AutoSync re-syncs the DB from defPath and secretsPath if either file has
// changed since the last sync. It is a no-op if both files are unchanged.
func AutoSync(db *sql.DB, defPath, secretsPath string) error {
	defMtime, err := mtime(defPath)
	if err != nil {
		return err
	}
	secretsMtime, err := mtime(secretsPath)
	if err != nil {
		return err
	}

	lastSynced, err := getLastSynced(db)
	if err != nil {
		return err
	}

	newest := defMtime
	if secretsMtime.After(defMtime) {
		newest = secretsMtime
	}

	if !newest.After(lastSynced) {
		return nil
	}

	defs, err := parser.ParseDef(defPath)
	if err != nil {
		return err
	}
	secrets, err := parser.ParseSecrets(secretsPath)
	if err != nil {
		return err
	}
	entries, warn, err := parser.Merge(defs, secrets)
	if err != nil {
		return err
	}
	if warn != nil {
		fmt.Fprintf(os.Stderr, "warning: %s\n", warn.Error())
	}

	return syncEntries(db, entries)
}

func syncEntries(db *sql.DB, entries []parser.Entry) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("cannot begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, e := range entries {
		var existingUUID string
		err := tx.QueryRow(`SELECT uuid FROM secrets WHERE name = ?`, e.Name).Scan(&existingUUID)
		if err == sql.ErrNoRows {
			existingUUID = uuid.New().String()
		} else if err != nil {
			return fmt.Errorf("cannot query existing secret: %w", err)
		}

		_, err = tx.Exec(`
			INSERT INTO secrets (uuid, name, description, secret_value)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(name) DO UPDATE SET
				description  = excluded.description,
				secret_value = excluded.secret_value
		`, existingUUID, e.Name, e.Description, e.Value)
		if err != nil {
			return fmt.Errorf("cannot upsert secret %s: %w", e.Name, err)
		}
	}

	// Remove entries no longer in the definition files.
	names := make([]any, len(entries))
	placeholders := ""
	for i, e := range entries {
		names[i] = e.Name
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
	}
	if len(names) > 0 {
		_, err = tx.Exec(`DELETE FROM secrets WHERE name NOT IN (`+placeholders+`)`, names...)
		if err != nil {
			return fmt.Errorf("cannot prune removed secrets: %w", err)
		}
	} else {
		_, err = tx.Exec(`DELETE FROM secrets`)
		if err != nil {
			return fmt.Errorf("cannot clear secrets: %w", err)
		}
	}

	// Rebuild FTS index.
	_, err = tx.Exec(`INSERT INTO secrets_fts(secrets_fts) VALUES('rebuild')`)
	if err != nil {
		return fmt.Errorf("cannot rebuild FTS index: %w", err)
	}

	// Update last_synced timestamp.
	_, err = tx.Exec(`
		INSERT INTO meta (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, metaKeyLastSynced, strconv.FormatInt(time.Now().UnixNano(), 10))
	if err != nil {
		return fmt.Errorf("cannot update last_synced: %w", err)
	}

	return tx.Commit()
}

func getLastSynced(db *sql.DB) (time.Time, error) {
	var raw string
	err := db.QueryRow(`SELECT value FROM meta WHERE key = ?`, metaKeyLastSynced).Scan(&raw)
	if err == sql.ErrNoRows {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("cannot read last_synced: %w", err)
	}
	ns, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("cannot parse last_synced: %w", err)
	}
	return time.Unix(0, ns), nil
}

func mtime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, fmt.Errorf("cannot stat %s: %w", path, err)
	}
	return info.ModTime(), nil
}
