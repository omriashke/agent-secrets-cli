package db

import (
	"database/sql"
	"fmt"
	"math"
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
	results, err := QueryTop(db, description, 1)
	if err != nil {
		return nil, err
	}
	return &results[0], nil
}

// relevanceThreshold controls how close a result's rank must be to the best
// match to be included in auto-detection mode. A result is included if its
// rank is within this factor of the best rank. FTS5 ranks are negative
// (closer to 0 = better), so we compare absolute values.
const relevanceThreshold = 0.5

// QueryAuto performs a full-text search and returns all results whose rank is
// within the relevance threshold of the best match. This automatically returns
// 1 result for precise queries and multiple for ambiguous ones.
func QueryAuto(db *sql.DB, description string) ([]Secret, error) {
	rows, err := db.Query(`
		SELECT s.uuid, s.name, s.description, s.secret_value, secrets_fts.rank
		FROM secrets_fts
		JOIN secrets s ON secrets_fts.rowid = s.rowid
		WHERE secrets_fts MATCH ?
		ORDER BY rank
	`, description)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	type ranked struct {
		secret Secret
		rank   float64
	}
	var all []ranked
	for rows.Next() {
		var r ranked
		if err := rows.Scan(&r.secret.UUID, &r.secret.Name, &r.secret.Description, &r.secret.Value, &r.rank); err != nil {
			return nil, fmt.Errorf("cannot scan secret row: %w", err)
		}
		all = append(all, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	if len(all) == 0 {
		return nil, fmt.Errorf("no secret found matching %q — run 'agent-secrets list' to see available secrets", description)
	}

	bestRank := math.Abs(all[0].rank)
	cutoff := bestRank * (1 + relevanceThreshold)

	var results []Secret
	for _, r := range all {
		if math.Abs(r.rank) > cutoff {
			break
		}
		results = append(results, r.secret)
	}

	return results, nil
}

// QueryTop performs a full-text search and returns up to n results ranked by relevance.
func QueryTop(db *sql.DB, description string, n int) ([]Secret, error) {
	if n < 1 {
		n = 1
	}
	rows, err := db.Query(`
		SELECT s.uuid, s.name, s.description, s.secret_value
		FROM secrets_fts
		JOIN secrets s ON secrets_fts.rowid = s.rowid
		WHERE secrets_fts MATCH ?
		ORDER BY rank
		LIMIT ?
	`, description, n)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
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
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no secret found matching %q — run 'agent-secrets list' to see available secrets", description)
	}
	return results, nil
}
