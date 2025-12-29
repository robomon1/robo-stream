package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Storage handles persistent data storage using JSON files
type Storage struct {
	dataDir string
	mu      sync.RWMutex
}

// New creates a new Storage instance
func New(dataDir string) (*Storage, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	return &Storage{dataDir: dataDir}, nil
}

// LoadJSON loads data from a JSON file
func (s *Storage) LoadJSON(filename string, v interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.dataDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, return empty
		}
		return err
	}

	if len(data) == 0 {
		return nil // Empty file
	}

	return json.Unmarshal(data, v)
}

// SaveJSON saves data to a JSON file
func (s *Storage) SaveJSON(filename string, v interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(s.dataDir, filename)
	return os.WriteFile(path, data, 0644)
}

// GetDataDir returns the data directory path
func (s *Storage) GetDataDir() string {
	return s.dataDir
}
