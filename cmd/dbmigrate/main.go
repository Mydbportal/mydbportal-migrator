package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"mydbportal.com/dbmigrate/internal/cli"
	
	// Register engines
	_ "mydbportal.com/dbmigrate/internal/engine/mongo"
	_ "mydbportal.com/dbmigrate/internal/engine/mysql"
	_ "mydbportal.com/dbmigrate/internal/engine/postgres"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "dbmigrate",
		Short: "DB-Migrate-Go CLI",
		Run: func(cmd *cobra.Command, args []string) {
			// Default to help
			cmd.Help()
		},
	}

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Add a source server",
		Run: func(cmd *cobra.Command, args []string) {
			if err := cli.RunInit(); err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
		},
	}

	var backupCmd = &cobra.Command{
		Use:   "backup",
		Short: "Backup databases",
		Run: func(cmd *cobra.Command, args []string) {
			source, _ := cmd.Flags().GetString("source")
			db, _ := cmd.Flags().GetString("db")
			
			if source == "" {
				fmt.Println("Error: --source required")
				os.Exit(1)
			}

			if err := cli.RunBackup(source, db); err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
		},
	}
	backupCmd.Flags().String("source", "", "Source ID")
	backupCmd.Flags().String("db", "", "Specific database name (optional)")

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List backups",
		Run: func(cmd *cobra.Command, args []string) {
			source, _ := cmd.Flags().GetString("source")
			if err := cli.RunList(source); err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
		},
	}
	listCmd.Flags().String("source", "", "Filter by Source ID (optional)")

	var restoreCmd = &cobra.Command{
		Use:   "restore",
		Short: "Restore a backup",
		Run: func(cmd *cobra.Command, args []string) {
			backup, _ := cmd.Flags().GetString("backup")
			target, _ := cmd.Flags().GetString("target")

			if backup == "" || target == "" {
				fmt.Println("Error: --backup and --target required")
				os.Exit(1)
			}

			if err := cli.RunRestore(backup, target); err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
		},
	}
	restoreCmd.Flags().String("backup", "", "Path to backup file")
	restoreCmd.Flags().String("target", "", "Target Source ID")

	var interactiveCmd = &cobra.Command{
		Use:   "interactive",
		Short: "Launch interactive menu",
		Run: func(cmd *cobra.Command, args []string) {
			cli.InteractiveMenu()
		},
	}

	rootCmd.AddCommand(initCmd, backupCmd, listCmd, restoreCmd, interactiveCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}