package cli

import (
	"fmt"
	"os"

	"golang.org/x/term"
	"mydbportal.com/dbmigrate/internal/config"
)

func InteractiveMenu() {
	for {
		fmt.Println("\n=== DB Migrate Interactive Mode ===")
		fmt.Println("1) Init/Add source server")
		fmt.Println("2) Backup all databases from source")
		fmt.Println("3) List backups")
		fmt.Println("4) Restore a backup to target")
		fmt.Println("5) Exit")
		fmt.Print("Choose an option: ")

		key, err := readKey()
		if err != nil {
			fmt.Println("Error reading input:", err)
			break
		}
		fmt.Println(key) // Echo choice

		switch key {
		case "1":
			if err := RunInit(); err != nil {
				fmt.Println("Error:", err)
			}
		case "2":
			// Ask for source
			mgr, _ := config.NewManager()
			fmt.Println("Available Sources:")
			for _, s := range mgr.ListSources() {
				fmt.Printf("- %s (%s)\n", s.ID, s.Engine)
			}
			id := readLine("Enter Source ID to backup: ")
			if err := RunBackup(id, ""); err != nil {
				fmt.Println("Error:", err)
			}
		case "3":
			if err := RunList(""); err != nil {
				fmt.Println("Error:", err)
			}
		case "4":
			// Restore
			path := readLine("Enter full path to backup file: ")
			mgr, _ := config.NewManager()
			fmt.Println("Available Targets (Sources):")
			for _, s := range mgr.ListSources() {
				fmt.Printf("- %s\n", s.ID)
			}
			targetID := readLine("Enter Target ID: ")
			if err := RunRestore(path, targetID); err != nil {
				fmt.Println("Error:", err)
			}
		case "5":
			fmt.Println("Bye!")
			return
		default:
			fmt.Println("Invalid option")
		}
	}
}

func readKey() (string, error) {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	b := make([]byte, 1)
	_, err = os.Stdin.Read(b)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
