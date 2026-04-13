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

		topExplicit := cmd.Flags().Changed("top")
		top, _ := cmd.Flags().GetInt("top")

		var secrets []db.Secret
		if topExplicit {
			secrets, err = db.QueryTop(database, args[0], top)
		} else {
			secrets, err = db.QueryAuto(database, args[0])
		}
		if err != nil {
			return err
		}

		valueOnly, _ := cmd.Flags().GetBool("value-only")
		asJSON, _ := cmd.Flags().GetBool("json")

		if len(secrets) == 1 {
			return printSingleResult(cmd, secrets[0], valueOnly, asJSON)
		}
		return printMultipleResults(cmd, secrets, valueOnly, asJSON)
	},
}

func printSingleResult(cmd *cobra.Command, s db.Secret, valueOnly, asJSON bool) error {
	switch {
	case asJSON:
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]string{
			"name":        s.Name,
			"description": s.Description,
			"value":       s.Value,
		})
	case valueOnly:
		fmt.Fprintln(cmd.OutOrStdout(), s.Value)
	default:
		fmt.Fprintf(cmd.OutOrStdout(), "name:        %s\ndescription: %s\nvalue:       %s\n",
			s.Name, s.Description, s.Value)
	}
	return nil
}

func printMultipleResults(cmd *cobra.Command, secrets []db.Secret, valueOnly, asJSON bool) error {
	switch {
	case asJSON:
		rows := make([]map[string]string, len(secrets))
		for i, s := range secrets {
			rows[i] = map[string]string{
				"name":        s.Name,
				"description": s.Description,
				"value":       s.Value,
			}
		}
		return json.NewEncoder(cmd.OutOrStdout()).Encode(rows)
	case valueOnly:
		for _, s := range secrets {
			fmt.Fprintln(cmd.OutOrStdout(), s.Value)
		}
	default:
		for i, s := range secrets {
			if i > 0 {
				fmt.Fprintln(cmd.OutOrStdout())
			}
			fmt.Fprintf(cmd.OutOrStdout(), "[%d]\nname:        %s\ndescription: %s\nvalue:       %s\n",
				i+1, s.Name, s.Description, s.Value)
		}
	}
	return nil
}

func init() {
	queryCmd.Flags().Bool("value-only", false, "Print only the secret value")
	queryCmd.Flags().Bool("json", false, "Output as JSON")
	queryCmd.Flags().Int("top", 1, "Force exactly N results (default: auto-detect by relevance)")
}
