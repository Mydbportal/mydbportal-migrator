package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "dbmigrate",
		Short: "DB-Migrate-Go is a cross-database migration tool",
		Long:  `A cross-database migration tool written in Go that can backup all databases from a source server, tag them, store them locally, and push to a target server.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Default to interactive mode if no args provided? 
            // Or show help. The spec says "dbmigrate interactive" launches interactive mode.
            // If just "dbmigrate" is run, we can maybe show help or also launch interactive.
            // For now, let's just print help.
			cmd.Help()
		},
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
