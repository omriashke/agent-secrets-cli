package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/omriashke/agent-secrets-cli/internal/config"
	"github.com/omriashke/agent-secrets-cli/internal/db"
	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:   "query <description>",
	Short: "Find a secret by description using full-text search",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		defPath, err := config.DefPath()
		if err != nil {
			return err
		}
		secretsPath, err := config.SecretsPath()
		if err != nil {
			return err
		}
		dbPath, err := config.DBPath()
		if err != nil {
			return err
		}

		database, err := db.Open(dbPath)
		if err != nil {
			return err
		}
		defer database.Close()

		if err := db.AutoSync(database, defPath, secretsPath); err != nil {
			return err
		}

		secret, err := db.Query(database, args[0])
		if err != nil {
			return err
		}

		valueOnly, _ := cmd.Flags().GetBool("value-only")
		asJSON, _ := cmd.Flags().GetBool("json")

		switch {
		case asJSON:
			return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]string{
				"name":        secret.Name,
				"description": secret.Description,
				"value":       secret.Value,
			})
		case valueOnly:
			fmt.Fprintln(cmd.OutOrStdout(), secret.Value)
		default:
			fmt.Fprintf(cmd.OutOrStdout(), "name:        %s\ndescription: %s\nvalue:       %s\n",
				secret.Name, secret.Description, secret.Value)
		}
		return nil
	},
}

func init() {
	queryCmd.Flags().Bool("value-only", false, "Print only the secret value (default behaviour for scripts)")
	queryCmd.Flags().Bool("json", false, "Output as JSON")
}
