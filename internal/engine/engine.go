package engine

import (
	"fmt"

	"mydbportal.com/dbmigrate/internal/config"
)

// BackupResult holds result for a single database backup
type BackupResult struct {
	Database string
	Filename string
	Error    error
}

type Engine interface {
	// ID returns the engine type identifier (e.g., "mysql")
	ID() string
	// ListDatabases returns a list of database names from the source
	ListDatabases(creds config.ServerConfig) ([]string, error)
	// BackupDatabase backs up a single database to the specified file path
	BackupDatabase(creds config.ServerConfig, dbName string, destPath string) error
	// BackupAll backs up all databases (or the cluster) to the specified directory
	// Returns a slice of BackupResult for each database processed
	BackupAll(creds config.ServerConfig, destDir string) ([]BackupResult, error)
	// RestoreBackup restores a backup file to the target
	RestoreBackup(creds config.ServerConfig, filePath string, dbName string) error
}

// Factory function type
type Factory func() Engine

var engines = make(map[string]Factory)

func Register(name string, factory Factory) {
	engines[name] = factory
}

func Get(name string) (Engine, error) {
	factory, ok := engines[name]
	if !ok {
		return nil, fmt.Errorf("engine not found: %s", name)
	}
	return factory(), nil
}

func ListEngines() []string {
	keys := make([]string, 0, len(engines))
	for k := range engines {
		keys = append(keys, k)
	}
	return keys
}