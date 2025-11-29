package mysql

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
	engine.Register("mysql", func() engine.Engine {
		return &MySQLEngine{}
	})
}

type MySQLEngine struct{}

func (e *MySQLEngine) ID() string {
	return "mysql"
}

func (e *MySQLEngine) getEnv(creds config.ServerConfig) []string {
	env := os.Environ()
	env = append(env, fmt.Sprintf("MYSQL_PWD=%s", creds.Password))
	return env
}

func (e *MySQLEngine) ListDatabases(creds config.ServerConfig) ([]string, error) {
	// mysql -h host -P port -u user -e "SHOW DATABASES;" --skip-column-names
	args := []string{
		"-h", creds.Host,
		"-P", fmt.Sprintf("%d", creds.Port),
		"-u", creds.User,
		"-e", "SHOW DATABASES;",
		"--skip-column-names",
	}

	cmd := exec.Command("mysql", args...)
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
		// Filter system databases
		switch db {
		case "information_schema", "mysql", "performance_schema", "sys":
			continue
		}
		dbs = append(dbs, db)
	}
	return dbs, nil
}

func (e *MySQLEngine) BackupDatabase(creds config.ServerConfig, dbName string, destPath string) error {
	// mysqldump ...
	args := []string{
		"-h", creds.Host,
		"-P", fmt.Sprintf("%d", creds.Port),
		"-u", creds.User,
		"--single-transaction",
		"--routines",
		"--triggers",
		"--databases", dbName,
	}

	cmd := exec.Command("mysqldump", args...)
	cmd.Env = e.getEnv(creds)

	return util.RunDumpToFile(cmd, destPath)
}

func (e *MySQLEngine) BackupAll(creds config.ServerConfig, destDir string) ([]engine.BackupResult, error) {
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

func (e *MySQLEngine) RestoreBackup(creds config.ServerConfig, filePath string, dbName string) error {
	// mysql ...
	args := []string{
		"-h", creds.Host,
		"-P", fmt.Sprintf("%d", creds.Port),
		"-u", creds.User,
	}
	// If dbName is provided, select it? Usually dump includes CREATE DATABASE/USE if --databases was used.
	// If we want to force a specific DB, we might need to create it first if not exists?
	// But mysqldump with --databases includes CREATE DATABASE.
	// If user wants to restore to a *different* name, it's harder with mysqldump output.
	// For now, assume we restore what's in the file.
    // The caller might pass dbName as context, but often with mysqldump it's embedded.
	
	cmd := exec.Command("mysql", args...)
	cmd.Env = e.getEnv(creds)

	return util.RestoreFromFile(cmd, filePath)
}