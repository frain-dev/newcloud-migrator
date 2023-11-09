package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:     "Migrate-to-Newcloud",
		Version: "0.1.0",
	}

	cmd.AddCommand(AddMigrateCommand())

	err := cmd.Execute()
	if err != nil {
		logrus.Fatal(err)
	}
}
