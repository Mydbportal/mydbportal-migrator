package postgres

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"mydbportal.com/dbmigrate/internal/config"
	"mydbportal.com/dbmigrate/internal/engine"
	"mydbportal.com/dbmigrate/internal/util"
)

func init() {
	engine.Register("postgres", func() engine.Engine {
		return &PostgresEngine{}
	})
}

type PostgresEngine struct{}

func (e *PostgresEngine) ID() string {
	return "postgres"
}

func (e *PostgresEngine) getEnv(creds config.ServerConfig) []string {
	env := os.Environ()
	env = append(env, fmt.Sprintf("PGPASSWORD=%s", creds.Password))
	return env
}

func (e *PostgresEngine) ListDatabases(creds config.ServerConfig) ([]string, error) {
	// psql -h host -p port -U user -d postgres -t -c "SELECT datname FROM pg_database WHERE datistemplate = false;"
	// Note: -d postgres is usually required to connect to *something* to list DBs.
	args := []string{
		"-h", creds.Host,
		"-p", fmt.Sprintf("%d", creds.Port),
		"-U", creds.User,
		"-d", "postgres",
		"-t",
		"-c", "SELECT datname FROM pg_database WHERE datistemplate = false;",
	}

	cmd := exec.Command("psql", args...)
	cmd.Env = e.getEnv(creds)

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
		dbs = append(dbs, db)
	}
	return dbs, nil
}

func (e *PostgresEngine) BackupDatabase(creds config.ServerConfig, dbName string, destPath string) error {
	// pg_dump -C -F p ...
	// -C: Include commands to create the database
	// -F p: Output plain-text SQL script
	
	// Retries for transient failures
	maxRetries := 3
	var lastErr error
	
	for i := 0; i < maxRetries; i++ {
		args := []string{
			"-h", creds.Host,
			"-p", fmt.Sprintf("%d", creds.Port),
			"-U", creds.User,
			"-F", "p",
			"-C",
			dbName,
		}

		cmd := exec.Command("pg_dump", args...)
		cmd.Env = e.getEnv(creds)

		lastErr = util.RunDumpToFile(cmd, destPath)
		if lastErr == nil {
			return nil
		}
		
		// If error, wait and retry
		time.Sleep(time.Second * time.Duration(i+1))
	}
	
	return lastErr
}

func (e *PostgresEngine) BackupAll(creds config.ServerConfig, destDir string) ([]engine.BackupResult, error) {
	dbs, err := e.ListDatabases(creds)
	if err != nil {
		return nil, err
	}

	var results []engine.BackupResult
	timestamp := time.Now().Format("2006-01-02T15:04:05Z")

	for _, db := range dbs {
		filename := fmt.Sprintf("%s_%s.sql.gz", db, timestamp)
		destPath := filepath.Join(destDir, filename)
		
		err := e.BackupDatabase(creds, db, destPath)
		
		res := engine.BackupResult{
			Database: db,
			Filename: filename,
			Error:    err,
		}
		results = append(results, res)
	}
	return results, nil
}

func (e *PostgresEngine) RestoreBackup(creds config.ServerConfig, filePath string, dbName string) error {
	// psql -h target -U user -d postgres (since the file contains CREATE DATABASE, we connect to postgres)
	args := []string{
		"-h", creds.Host,
		"-p", fmt.Sprintf("%d", creds.Port),
		"-U", creds.User,
		"-d", "postgres",
	}
	
	cmd := exec.Command("psql", args...)
	cmd.Env = e.getEnv(creds)

	return util.RestoreFromFile(cmd, filePath)
}