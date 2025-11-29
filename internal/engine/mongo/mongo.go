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
	// mongosh --host host --port port --username user --password --authenticationDatabase admin --eval "db.adminCommand('listDatabases').databases.forEach(d => print(d.name))" --quiet
	// Requires piping password to stdin.
	
	args := []string{
		"--host", creds.Host,
		"--port", fmt.Sprintf("%d", creds.Port),
		"--username", creds.User,
		"--password",
		"--authenticationDatabase", "admin", // Default assumption
		"--eval", "db.adminCommand('listDatabases').databases.forEach(d => print(d.name))",
		"--quiet",
	}

	cmd := exec.Command("mongosh", args...)
	cmd.Stdin = strings.NewReader(creds.Password + "\n")

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try 'mongo' legacy client if 'mongosh' fails?
		// For now, error out.
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
	// mongodump ... --archive --db ...
	args := []string{
		"--host", creds.Host,
		"--port", fmt.Sprintf("%d", creds.Port),
		"--username", creds.User,
		"--password",
		"--authenticationDatabase", "admin",
		"--archive",
		"--db", dbName,
	}

	cmd := exec.Command("mongodump", args...)
	
	// We need to pipe password to Stdin AND pipe stdout to file (via util).
	// BUT util.RunDumpToFile sets cmd.Stdout.
	// We need to set Stdin.
	// Wait, if we set cmd.Stdin, we can't just use simple string reader if `RunDumpToFile` doesn't touch Stdin.
	// `RunDumpToFile` does NOT touch Stdin. So we can set it.
	cmd.Stdin = strings.NewReader(creds.Password + "\n")

	return util.RunDumpToFile(cmd, destPath)
}

func (e *MongoEngine) BackupAll(creds config.ServerConfig, destDir string) (map[string]string, error) {
	// mongodump --archive ... (dumps all)
	// We only produce one file for Mongo "All" backup usually.
	
	timestamp := time.Now().Format("2006-01-02T15:04:05Z")
	filename := fmt.Sprintf("all-databases_%s.archive.gz", timestamp)
	destPath := filepath.Join(destDir, filename)
	
	args := []string{
		"--host", creds.Host,
		"--port", fmt.Sprintf("%d", creds.Port),
		"--username", creds.User,
		"--password",
		"--authenticationDatabase", "admin",
		"--archive",
	}
	
	cmd := exec.Command("mongodump", args...)
	cmd.Stdin = strings.NewReader(creds.Password + "\n")
	
	if err := util.RunDumpToFile(cmd, destPath); err != nil {
		return nil, err
	}
	
	return map[string]string{"all": filename}, nil
}

func (e *MongoEngine) RestoreBackup(creds config.ServerConfig, filePath string, dbName string) error {
	// mongorestore --archive ...
	args := []string{
		"--host", creds.Host,
		"--port", fmt.Sprintf("%d", creds.Port),
		"--username", creds.User,
		"--password",
		"--authenticationDatabase", "admin",
		"--archive",
		"--nsInclude=*", // Restore everything in archive
	}
	
	cmd := exec.Command("mongorestore", args...)
	
	// mongorestore needs password on Stdin AND archive on Stdin?
	// NO. If reading from Stdin, `--archive` without value reads from Stdin.
	// BUT how to provide password?
	// If we provide password via Stdin, we can't provide archive via Stdin easily (unless mixed stream? No).
	// We MUST use URI or config file or different auth mechanism if we pipe archive to Stdin.
	// Or use `--password=PASS` arg (visible in process list).
	// Or `mongorestore` supports `--archive=FILE`. We have the file on disk (gzipped).
	// `mongorestore` supports gzipped archives directly?
	// "mongorestore --archive=test.20150715.gz --gzip"
	// Our file is manually gzipped.
	// So we can just pass the file path to `--archive` and use `--gzip`?
	// Wait, we gzipped the output of `mongodump --archive`.
	// `mongorestore --gzip --archive=path` should work.
	// Then we can use Stdin for password.
	
	// Let's check util.RestoreFromFile. It opens file, gzips reader, pipes to Stdin.
	// So `cmd` receives unzipped stream on Stdin.
	// So we run `mongorestore --archive`. (reads from stdin).
	// Problem: Password input.
	
	// Solution: Use URI for Restore to avoid Stdin conflict.
	// OR, since `mongorestore` is local execution, maybe env var?
	// Mongo tools don't support env var password.
	// URI is the viable option if Stdin is used for data.
	// `mongodb://user:pass@host:port/?authSource=admin`
	
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d/?authSource=admin", 
		creds.User, creds.Password, creds.Host, creds.Port)
		
	args = []string{
		"--uri", uri,
		"--archive",
		"--nsInclude=*",
	}
	
	cmd = exec.Command("mongorestore", args...)
	// Stdin will be set by util.RestoreFromFile
	
	return util.RestoreFromFile(cmd, filePath)
}
