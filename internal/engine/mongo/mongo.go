package mongo

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"mydbportal.com/dbmigrate/internal/config"
	"mydbportal.com/dbmigrate/internal/engine"
	"mydbportal.com/dbmigrate/internal/util"
)

func init() {
	engine.Register("mongo", func() engine.Engine {
		return &MongoEngine{}
	})
}

type MongoEngine struct{}

func (e *MongoEngine) ID() string {
	return "mongo"
}

func (e *MongoEngine) ListDatabases(creds config.ServerConfig) ([]string, error) {
	// Use explicit password flag to avoid parsing issues
	args := []string{
		"--host", creds.Host,
		"--port", fmt.Sprintf("%d", creds.Port),
		"--username", creds.User,
		"--password", creds.Password,
		"--authenticationDatabase", "admin", 
		"--eval", "db.adminCommand('listDatabases').databases.forEach(d => print(d.name))",
		"--quiet",
	}

	cmd := exec.Command("mongosh", args...)
	// Stdin not needed for password anymore

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %s, output: %s", err, string(output))
	}

	var dbs []string
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		db := strings.TrimSpace(scanner.Text())
		if db == "" {
			continue
		}
		// Filter system dbs? local, admin, config
		switch db {
		case "admin", "config", "local":
			continue
		}
		dbs = append(dbs, db)
	}
	return dbs, nil
}

func (e *MongoEngine) BackupDatabase(creds config.ServerConfig, dbName string, destPath string) error {
	// Retry logic
	maxRetries := 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		// Use explicit password flag
		args := []string{
			"--host", creds.Host,
			"--port", fmt.Sprintf("%d", creds.Port),
			"--username", creds.User,
			"--password", creds.Password,
			"--authenticationDatabase", "admin",
			"--archive",
			"--db", dbName,
		}

		cmd := exec.Command("mongodump", args...)
		
		lastErr = util.RunDumpToFile(cmd, destPath)
		if lastErr == nil {
			return nil
		}
		// If error, wait and retry
		time.Sleep(time.Second * time.Duration(i+1))
	}

	return lastErr
}

func (e *MongoEngine) BackupAll(creds config.ServerConfig, destDir string) ([]engine.BackupResult, error) {
	// mongodump --archive ... (dumps all)
	
	timestamp := time.Now().Format("2006-01-02T15:04:05Z")
	filename := fmt.Sprintf("all-databases_%s.archive.gz", timestamp)
	destPath := filepath.Join(destDir, filename)
	
	// Retry logic
	maxRetries := 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		args := []string{
			"--host", creds.Host,
			"--port", fmt.Sprintf("%d", creds.Port),
			"--username", creds.User,
			"--password", creds.Password,
			"--authenticationDatabase", "admin",
			"--archive",
		}
		
		cmd := exec.Command("mongodump", args...)
		// Stdin not needed for password
		
		lastErr = util.RunDumpToFile(cmd, destPath)
		if lastErr == nil {
			break
		}
		// If error, wait and retry
		time.Sleep(time.Second * time.Duration(i+1))
	}
	
	res := engine.BackupResult{
		Database: "all",
		Filename: filename,
		Error:    lastErr,
	}
	
	return []engine.BackupResult{res}, nil
}

func (e *MongoEngine) RestoreBackup(creds config.ServerConfig, filePath string, dbName string) error {
	// Use URI for restore as before (standard for restore + archive piping)
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d/?authSource=admin", 
		creds.User, creds.Password, creds.Host, creds.Port)
		
	args := []string{
		"--uri", uri,
		"--archive",
		"--nsInclude=*",
	}
	
	cmd := exec.Command("mongorestore", args...)
	// Stdin will be set by util.RestoreFromFile
	
	return util.RestoreFromFile(cmd, filePath)
}