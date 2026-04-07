package db

import (
	"database/sql"
	"fmt"
)

// Secret is a row from the secrets table.
type Secret struct {
	UUID        string
	Name        string
	Description string
	Value       string
}

// List returns all secrets ordered by name.
func List(db *sql.DB) ([]Secret, error) {
	rows, err := db.Query(`SELECT uuid, name, description, secret_value FROM secrets ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("cannot list secrets: %w", err)
	}
	defer rows.Close()

	var results []Secret
	for rows.Next() {
		var s Secret
		if err := rows.Scan(&s.UUID, &s.Name, &s.Description, &s.Value); err != nil {
			return nil, fmt.Errorf("cannot scan secret row: %w", err)
		}
		results = append(results, s)
	}
	return results, rows.Err()
}

// Query performs a full-text search on descriptions and returns the best match.
func Query(db *sql.DB, description string) (*Secret, error) {
	row := db.QueryRow(`
		SELECT s.uuid, s.name, s.description, s.secret_value
		FROM secrets_fts
		JOIN secrets s ON secrets_fts.rowid = s.rowid
		WHERE secrets_fts MATCH ?
		ORDER BY rank
		LIMIT 1
	`, description)

	var s Secret
	err := row.Scan(&s.UUID, &s.Name, &s.Description, &s.Value)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no secret found matching %q — run 'agent-secrets list' to see available secrets", description)
	}
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	return &s, nil
}
