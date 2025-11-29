package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type BackupFile struct {
	Name     string `json:"name"`
	Checksum string `json:"checksum"`
	Size     int64  `json:"size"`
}

type Metadata struct {
	ID        string       `json:"id"`
	Engine    string       `json:"engine"`
	Host      string       `json:"host"`
	Port      int          `json:"port"`
	User      string       `json:"user"`
	Timestamp string       `json:"timestamp"` // ISO8601
	Files     []BackupFile `json:"files"`
	Status    string       `json:"status"` // success, failed
}

// Root directory for backups
var BackupRoot = "backups"

func InitBackupDir(engine, host string, ts time.Time) (string, string, error) {
	// Format: backups/engine/source-host_timestamp/
	tsStr := ts.Format(time.RFC3339)
	dirName := fmt.Sprintf("source-%s_%s", host, tsStr)
	path := filepath.Join(BackupRoot, engine, dirName)
	
	if err := os.MkdirAll(path, 0755); err != nil {
		return "", "", err
	}
	return path, tsStr, nil
}

func WriteMetadata(dirPath string, meta Metadata) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dirPath, "metadata.json"), data, 0644)
}

func LoadMetadata(path string) (Metadata, error) {
	var meta Metadata
	data, err := os.ReadFile(path)
	if err != nil {
		return meta, err
	}
	err = json.Unmarshal(data, &meta)
	return meta, err
}

// ListBackups returns all backups for a given engine/host (or all if empty)
// This is a simplified version. In a real app, we might index them.
func ListBackups() ([]Metadata, error) {
	var backups []Metadata
	
	// Walk through backups/
	// structure: backups/<engine>/<backup_dir>/metadata.json
	
	entries, err := os.ReadDir(BackupRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return backups, nil
		}
		return nil, err
	}

	for _, engineDir := range entries {
		if !engineDir.IsDir() {
			continue
		}
		enginePath := filepath.Join(BackupRoot, engineDir.Name())
		
		backupDirs, err := os.ReadDir(enginePath)
		if err != nil {
			continue
		}

		for _, bd := range backupDirs {
			if !bd.IsDir() {
				continue
			}
			metaPath := filepath.Join(enginePath, bd.Name(), "metadata.json")
			meta, err := LoadMetadata(metaPath)
			if err == nil {
				backups = append(backups, meta)
			}
		}
	}

	// Sort by timestamp desc
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp > backups[j].Timestamp
	})

	return backups, nil
}
