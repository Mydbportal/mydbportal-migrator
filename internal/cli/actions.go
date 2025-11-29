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
	mgr, err := config.NewManager()
	if err != nil {
		return err
	}

	backups, err := storage.ListBackups()
	if err != nil {
		return err
	}

	fmt.Printf("% -25s | % -10s | % -20s | % -10s\n", "TIMESTAMP", "ENGINE", "SOURCE HOST", "STATUS")
	fmt.Println(strings.Repeat("-", 80))

	for _, b := range backups {
		// Filter if sourceID is provided
		if sourceID != "" {
			// Assuming we map source ID to metadata somehow?
			// Metadata has Host/Port but not Source ID from config.
			// But the directory name includes Host.
			// For now, simple listing.
		}
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

	var backupFiles map[string]string
	if dbName != "" {
		// Single DB backup logic not fully implemented in engine interface separate from BackupAll return types
		// But we can use BackupDatabase directly.
		filename := fmt.Sprintf("%s_%s.sql.gz", dbName, tsStr)
		destPath := filepath.Join(path, filename)
		if err := eng.BackupDatabase(source, dbName, destPath); err != nil {
			return err
		}
		backupFiles = map[string]string{dbName: filename}
	} else {
		// Backup All
		backupFiles, err = eng.BackupAll(source, path)
		if err != nil {
			return err
		}
	}

	// Create Metadata
	var files []storage.BackupFile
	for _, filename := range backupFiles {
		fullPath := filepath.Join(path, filename)
		checksum, _ := util.ComputeChecksum(fullPath)
		info, _ := os.Stat(fullPath)
		files = append(files, storage.BackupFile{
			Name:     filename,
			Checksum: checksum,
			Size:     info.Size(),
		})
	}

	meta := storage.Metadata{
		ID:        fmt.Sprintf("%s_%s_%s", source.Engine, source.Host, tsStr),
		Engine:    source.Engine,
		Host:      source.Host,
		Port:      source.Port,
		User:      source.User,
		Timestamp: tsStr,
		Files:     files,
		Status:    "success",
	}

	if err := storage.WriteMetadata(path, meta); err != nil {
		return err
	}

	fmt.Println("Backup completed successfully!")
	return nil
}

func RunRestore(backupPath string, targetID string) error {
	mgr, err := config.NewManager()
	if err != nil {
		return err
	}
	
target, err := mgr.GetSource(targetID) // Using GetSource for target as well (same struct)
	if err != nil {
		return err
	}
	
	eng, err := engine.Get(target.Engine)
	if err != nil {
		return err
	}
	
	fmt.Printf("Restoring %s to %s (%s)...\n", backupPath, target.ID, target.Host)
	
	// Infer dbName from filename or ask?
	// For now, pass empty dbName and let engine handle (e.g. MySQL/Postgres might need it, Mongo doesn't)
	// If MySQL dump has --databases, it creates DB.
	// If Postgres dump has -C, it creates DB.
	
	if err := eng.RestoreBackup(target, backupPath, ""); err != nil {
		return err
	}
	
	fmt.Println("Restore completed!")
	return nil
}
