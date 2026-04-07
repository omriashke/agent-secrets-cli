package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/omriashke/agent-secrets-cli/internal/config"
	"github.com/omriashke/agent-secrets-cli/internal/db"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all secret names and descriptions",
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

		secrets, err := db.List(database)
		if err != nil {
			return err
		}

		if len(secrets) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No secrets defined. Edit ~/.agent-secrets/secrets.def to add some.")
			return nil
		}

		asJSON, _ := cmd.Flags().GetBool("json")
		if asJSON {
			type row struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			}
			rows := make([]row, len(secrets))
			for i, s := range secrets {
				rows[i] = row{Name: s.Name, Description: s.Description}
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(rows)
		}

		for _, s := range secrets {
			fmt.Fprintf(cmd.OutOrStdout(), "%-30s %s\n", s.Name, s.Description)
		}
		return nil
	},
}

func init() {
	listCmd.Flags().Bool("json", false, "Output as JSON")
}
