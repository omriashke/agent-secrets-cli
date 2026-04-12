package diff

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/omriashke/agent-secrets-cli/internal/fileutil"
)

type Change struct {
	Kind        string // "added", "removed", "changed"
	Name        string
	Description string // for display
	OldValue    string // only for "changed"
	NewValue    string // only for "changed" and "added"
}

// MaskValue shows only the first 4 and last 4 characters, masking the rest.
func MaskValue(v string) string {
	if len(v) <= 10 {
		return strings.Repeat("*", len(v))
	}
	return v[:4] + strings.Repeat("*", len(v)-8) + v[len(v)-4:]
}

// ComputeChanges compares local and remote dotenv files and returns the
// changes that would be applied if the "incoming" side is accepted.
// direction is "push" (local→remote) or "pull" (remote→local).
// Returns changes from the perspective of what will happen to the destination.
func ComputeChanges(localDefPath, localSecretsPath, remoteDefPath, remoteSecretsPath string) ([]Change, error) {
	localDefs, err := fileutil.ReadDotenv(localDefPath)
	if err != nil {
		return nil, fmt.Errorf("reading local definitions: %w", err)
	}
	localSecrets, err := fileutil.ReadDotenv(localSecretsPath)
	if err != nil {
		return nil, fmt.Errorf("reading local secrets: %w", err)
	}
	remoteDefs, err := fileutil.ReadDotenv(remoteDefPath)
	if err != nil {
		return nil, fmt.Errorf("reading remote definitions: %w", err)
	}
	remoteSecrets, err := fileutil.ReadDotenv(remoteSecretsPath)
	if err != nil {
		return nil, fmt.Errorf("reading remote secrets: %w", err)
	}

	return computeMapChanges(remoteDefs, remoteSecrets, localDefs, localSecrets), nil
}

// computeMapChanges computes what changes going from "old" to "new".
func computeMapChanges(oldDefs, oldSecrets, newDefs, newSecrets map[string]string) []Change {
	var changes []Change

	allKeys := map[string]bool{}
	for k := range oldDefs {
		allKeys[k] = true
	}
	for k := range newDefs {
		allKeys[k] = true
	}

	sorted := make([]string, 0, len(allKeys))
	for k := range allKeys {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)

	for _, name := range sorted {
		oldDesc, inOld := oldDefs[name]
		newDesc, inNew := newDefs[name]

		if inNew && !inOld {
			changes = append(changes, Change{
				Kind:        "added",
				Name:        name,
				Description: newDesc,
				NewValue:    newSecrets[name],
			})
		} else if inOld && !inNew {
			changes = append(changes, Change{
				Kind:        "removed",
				Name:        name,
				Description: oldDesc,
			})
		} else if inOld && inNew {
			descChanged := oldDesc != newDesc
			valChanged := oldSecrets[name] != newSecrets[name]
			if descChanged || valChanged {
				changes = append(changes, Change{
					Kind:        "changed",
					Name:        name,
					Description: newDesc,
					OldValue:    oldSecrets[name],
					NewValue:    newSecrets[name],
				})
			}
		}
	}

	return changes
}

// PrintChanges writes a human-readable summary of changes to w.
// direction is "push" or "pull" — used for the header.
func PrintChanges(w io.Writer, changes []Change, direction string) {
	if len(changes) == 0 {
		fmt.Fprintf(w, "No differences found — %s is a no-op.\n", direction)
		return
	}

	fmt.Fprintf(w, "\nThe following changes will be applied (%s):\n\n", direction)

	for _, c := range changes {
		switch c.Kind {
		case "added":
			fmt.Fprintf(w, "  + %-28s  %s\n", c.Name, c.Description)
			fmt.Fprintf(w, "    value: %s\n", MaskValue(c.NewValue))
		case "removed":
			fmt.Fprintf(w, "  - %-28s  %s\n", c.Name, c.Description)
		case "changed":
			fmt.Fprintf(w, "  ~ %-28s  %s\n", c.Name, c.Description)
			fmt.Fprintf(w, "    value: %s → %s\n", MaskValue(c.OldValue), MaskValue(c.NewValue))
		}
	}
	fmt.Fprintln(w)
}
