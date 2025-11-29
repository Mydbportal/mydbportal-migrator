package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/term"

	"mydbportal.com/dbmigrate/internal/config"
	"mydbportal.com/dbmigrate/internal/engine"
	"mydbportal.com/dbmigrate/internal/storage"
	"mydbportal.com/dbmigrate/internal/util"
)

// Helper to read line from stdin
func readLine(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

// Helper to read password
func readPassword(prompt string) string {
	fmt.Print(prompt)
	bytePassword, _ := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	return string(bytePassword)
}

func RunInit() error {
	fmt.Println("=== Add Source Server ===")
	mgr, err := config.NewManager()
	if err != nil {
		return err
	}

	id := readLine("Source ID (name): ")
	engineType := readLine("Engine (mysql, postgres, mongo): ")
	host := readLine("Host (IP/Domain): ")
	portStr := readLine("Port: ")
	user := readLine("User: ")
	pass := readPassword("Password: ")

	var port int
	fmt.Sscanf(portStr, "%d", &port)

	server := config.ServerConfig{
		ID:       id,
		Engine:   engineType,
		Host:     host,
		Port:     port,
		User:     user,
		Password: pass,
	}

	if err := mgr.AddSource(server); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	fmt.Println("Source added successfully!")
	return nil
}

func RunList(sourceID string) error {
	backups, err := storage.ListBackups()
	if err != nil {
		return err
	}

	fmt.Printf("% -25s | % -10s | % -20s | % -10s\n", "TIMESTAMP", "ENGINE", "SOURCE HOST", "STATUS")
	fmt.Println(strings.Repeat("-", 80))

	for _, b := range backups {
		fmt.Printf("% -25s | % -10s | % -20s | % -10s\n", b.Timestamp, b.Engine, b.Host, b.Status)
	}
	return nil
}

func RunBackup(sourceID string, dbName string) error {
	mgr, err := config.NewManager()
	if err != nil {
		return err
	}

	source, err := mgr.GetSource(sourceID)
	if err != nil {
		return err
	}

	eng, err := engine.Get(source.Engine)
	if err != nil {
		return err
	}

	timestamp := time.Now()
	path, tsStr, err := storage.InitBackupDir(source.Engine, source.Host, timestamp)
	if err != nil {
		return err
	}

	fmt.Printf("Starting backup for %s to %s...\n", source.ID, path)

	var backupResults []engine.BackupResult
	
	if dbName != "" {
		filename := fmt.Sprintf("%s_%s.sql.gz", dbName, tsStr)
		destPath := filepath.Join(path, filename)
		
		// Retry logic is now handled within Postgres engine, but for others or generic single calls:
		err := eng.BackupDatabase(source, dbName, destPath)
		backupResults = append(backupResults, engine.BackupResult{
			Database: dbName,
			Filename: filename,
			Error:    err,
		})
	} else {
		// Backup All
		var err error
		backupResults, err = eng.BackupAll(source, path)
		if err != nil {
			return fmt.Errorf("critical failure listing/backing up databases: %w", err)
		}
	}

	// Process Results
	var files []storage.BackupFile
	successCount := 0
	failCount := 0

	for _, res := range backupResults {
		bf := storage.BackupFile{
			Name: res.Filename,
		}

		if res.Error != nil {
			bf.Status = "failed"
			bf.Error = res.Error.Error()
			failCount++
			fmt.Printf(" [FAILED] %s: %v\n", res.Database, res.Error)
		} else {
			bf.Status = "success"
			successCount++
			fullPath := filepath.Join(path, res.Filename)
			checksum, _ := util.ComputeChecksum(fullPath)
			info, _ := os.Stat(fullPath)
			bf.Checksum = checksum
			if info != nil {
				bf.Size = info.Size()
			}
			fmt.Printf(" [OK] %s\n", res.Database)
		}
		files = append(files, bf)
	}

	status := "success"
	if failCount > 0 {
		if successCount == 0 {
			status = "failed"
		} else {
			status = "partial"
		}
	}

	meta := storage.Metadata{
		ID:        fmt.Sprintf("%s_%s_%s", source.Engine, source.Host, tsStr),
		Engine:    source.Engine,
		Host:      source.Host,
		Port:      source.Port,
		User:      source.User,
		Timestamp: tsStr,
		Files:     files,
		Status:    status,
	}

	if err := storage.WriteMetadata(path, meta); err != nil {
		return err
	}

	fmt.Printf("Backup operation completed with status: %s\n", status)
	return nil
}

func RunRestore(backupPath string, targetID string) error {
	mgr, err := config.NewManager()
	if err != nil {
		return err
	}
	
target, err := mgr.GetSource(targetID)
	if err != nil {
		return err
	}
	
	eng, err := engine.Get(target.Engine)
	if err != nil {
		return err
	}
	
	fmt.Printf("Restoring %s to %s (%s)...\n", backupPath, target.ID, target.Host)
	
	if err := eng.RestoreBackup(target, backupPath, ""); err != nil {
		return err
	}
	
	fmt.Println("Restore completed!")
	return nil
}
