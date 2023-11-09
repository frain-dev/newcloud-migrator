package main

import (
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var client = http.Client{Timeout: 15 * time.Second}

func AddMigrateCommand() *cobra.Command {
	var oldBaseURL string
	var oldPGDSN string
	var newPGDSN string
	var personalAccessKey string
	var migrateEvents bool

	cmd := &cobra.Command{
		Use:   "run",
		Short: "r",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := NewMigrator(oldBaseURL, oldPGDSN, newPGDSN, personalAccessKey, migrateEvents)
			if err != nil {
				return err
			}

			return m.Run()
		},
	}

	cmd.PersistentFlags().StringVar(&oldBaseURL, "old-base-url", "https://api.getconvoy.io", "Base URL of your previous deployment")
	cmd.PersistentFlags().StringVar(&oldPGDSN, "old-pg-dsn", "", "DSN of your previous postgres DB")
	cmd.PersistentFlags().StringVar(&newPGDSN, "new-pg-dsn", "", "DSN of your current postgres DB")
	cmd.PersistentFlags().StringVar(&personalAccessKey, "pat", "", "Your User Personal Access Token(from old deployment)")
	cmd.PersistentFlags().BoolVar(&migrateEvents, "migrate-events", false, "Run events migration")

	return cmd
}
