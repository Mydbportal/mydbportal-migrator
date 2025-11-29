package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"mydbportal.com/dbmigrate/internal/util"
)

const configFileName = ".dbmigrate.json"

// Default encryption key for prototype (in production, use user passphrase or keyring)
var encryptionKey = []byte("01234567890123456789012345678901") 

type ServerConfig struct {
	ID       string `json:"id"`
	Engine   string `json:"engine"` // mysql, postgres, mongo
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"` // Encrypted
}

type Config struct {
	Sources []ServerConfig `json:"sources"`
	Targets []ServerConfig `json:"targets"`
}

type Manager struct {
	configPath string
	Config     Config
	mu         sync.Mutex
}

func NewManager() (*Manager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(home, configFileName)
	
	mgr := &Manager{
		configPath: path,
		Config:     Config{},
	}
	
	if err := mgr.Load(); err != nil {
		// If file not found, just return empty manager
		if os.IsNotExist(err) {
			return mgr, nil
		}
		return nil, err
	}
	return mgr, nil
}

func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &m.Config)
}

func (m *Manager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := json.MarshalIndent(m.Config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.configPath, data, 0600)
}

func (m *Manager) AddSource(s ServerConfig) error {
	// Encrypt password before adding
	encryptedPass, err := util.Encrypt(s.Password, encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}
	s.Password = encryptedPass
	
	m.Config.Sources = append(m.Config.Sources, s)
	return m.Save()
}

func (m *Manager) GetSource(id string) (ServerConfig, error) {
	for _, s := range m.Config.Sources {
		if s.ID == id {
			// Decrypt password before returning
			decryptedPass, err := util.Decrypt(s.Password, encryptionKey)
			if err != nil {
				return s, fmt.Errorf("failed to decrypt password: %w", err)
			}
			s.Password = decryptedPass
			return s, nil
		}
	}
	return ServerConfig{}, fmt.Errorf("source not found: %s", id)
}

func (m *Manager) ListSources() []ServerConfig {
	return m.Config.Sources
}
