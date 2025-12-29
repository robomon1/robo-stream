package manager

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/robomon1/robo-stream/server/internal/models"
	"github.com/robomon1/robo-stream/server/internal/storage"
)

// ConfigManager manages button configurations
type ConfigManager struct {
	storage       *storage.Storage
	buttonManager *ButtonManager
	configs       map[string]*models.Configuration
}

// NewConfigManager creates a new ConfigManager
func NewConfigManager(storage *storage.Storage, buttonManager *ButtonManager) *ConfigManager {
	cm := &ConfigManager{
		storage:       storage,
		buttonManager: buttonManager,
		configs:       make(map[string]*models.Configuration),
	}
	cm.load()
	return cm
}

// load reads configurations from storage
func (cm *ConfigManager) load() error {
	var configs []*models.Configuration
	if err := cm.storage.LoadJSON("configs.json", &configs); err != nil {
		return err
	}
	for _, cfg := range configs {
		cm.configs[cfg.ID] = cfg
	}
	return nil
}

// save writes configurations to storage
func (cm *ConfigManager) save() error {
	configs := make([]*models.Configuration, 0, len(cm.configs))
	for _, cfg := range cm.configs {
		configs = append(configs, cfg)
	}
	return cm.storage.SaveJSON("configs.json", configs)
}

// Create creates a new configuration
func (cm *ConfigManager) Create(config *models.Configuration) error {
	config.ID = uuid.New().String()
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()
	if config.Buttons == nil {
		config.Buttons = make(map[string]string)
	}
	cm.configs[config.ID] = config
	return cm.save()
}

// Get retrieves a configuration by ID
func (cm *ConfigManager) Get(id string) (*models.Configuration, error) {
	cfg, ok := cm.configs[id]
	if !ok {
		return nil, fmt.Errorf("configuration not found: %s", id)
	}
	return cfg, nil
}

// List returns all configurations
func (cm *ConfigManager) List() []*models.Configuration {
	configs := make([]*models.Configuration, 0, len(cm.configs))
	for _, cfg := range cm.configs {
		configs = append(configs, cfg)
	}
	return configs
}

// Update updates an existing configuration
func (cm *ConfigManager) Update(config *models.Configuration) error {
	if _, ok := cm.configs[config.ID]; !ok {
		return fmt.Errorf("configuration not found: %s", config.ID)
	}
	config.UpdatedAt = time.Now()
	cm.configs[config.ID] = config
	return cm.save()
}

// Delete removes a configuration
func (cm *ConfigManager) Delete(id string) error {
	delete(cm.configs, id)
	return cm.save()
}

// SetDefault sets a configuration as the default
func (cm *ConfigManager) SetDefault(id string) error {
	// Clear default flag from all configs
	for _, cfg := range cm.configs {
		cfg.IsDefault = false
	}

	// Set new default
	cfg, ok := cm.configs[id]
	if !ok {
		return fmt.Errorf("configuration not found: %s", id)
	}
	cfg.IsDefault = true

	return cm.save()
}

// GetDefault returns the default configuration
func (cm *ConfigManager) GetDefault() (*models.Configuration, error) {
	for _, cfg := range cm.configs {
		if cfg.IsDefault {
			return cfg, nil
		}
	}
	return nil, fmt.Errorf("no default configuration set")
}

// Resolve converts a configuration to a resolved configuration with full button details
func (cm *ConfigManager) Resolve(id string) (*models.ResolvedConfiguration, error) {
	cfg, err := cm.Get(id)
	if err != nil {
		return nil, err
	}

	resolved := &models.ResolvedConfiguration{
		ID:      cfg.ID,
		Name:    cfg.Name,
		Grid:    cfg.Grid,
		Buttons: make([]models.ResolvedButton, 0),
	}

	// Resolve each button
	for position, buttonID := range cfg.Buttons {
		button, err := cm.buttonManager.Get(buttonID)
		if err != nil {
			continue // Skip if button not found
		}

		// Parse position (btn-0-0 -> row=0, col=0)
		parts := strings.Split(position, "-")
		if len(parts) != 3 {
			continue
		}
		row, _ := strconv.Atoi(parts[1])
		col, _ := strconv.Atoi(parts[2])

		resolvedBtn := models.ResolvedButton{
			ID:     position,
			Row:    row,
			Col:    col,
			Text:   button.Name,
			Icon:   button.Icon,
			Color:  button.Color,
			Action: button.Action,
		}

		resolved.Buttons = append(resolved.Buttons, resolvedBtn)
	}

	return resolved, nil
}
