package config

import (
	"encoding/json"
	"fmt"

	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/config"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/internal/pkg/cli"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/pkg/log"

	"github.com/spf13/cobra"
)

func AddConfigCommand(a *cli.App) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "config",
		Short: "config outputs your instances computed configuration",
		Annotations: map[string]string{
			"CheckMigration":  "false",
			"ShouldBootstrap": "false",
		},
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.Get()
			if err != nil {
				log.Fatalf("Error getting config: %v\n", err)
			}

			data, err := json.MarshalIndent(cfg, "", "    ")
			if err != nil {
				log.Fatalf("Error printing config: %v\n", err)
			}

			fmt.Println(string(data))
		},
	}

	return cmd
}
