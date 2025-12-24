package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// GridConfig defines the button grid layout
type GridConfig struct {
	Rows int `json:"rows"`
	Cols int `json:"cols"`
}

// ActionParams contains parameters for button actions
type ActionParams map[string]interface{}

// ButtonAction defines what happens when a button is pressed
type ButtonAction struct {
	Type   string       `json:"type"`   // switch_scene, toggle_stream, etc.
	Params ActionParams `json:"params"` // action-specific parameters
}

// Button represents a single button in the grid
type Button struct {
	ID     string       `json:"id"`
	Row    int          `json:"row"`
	Col    int          `json:"col"`
	Text   string       `json:"text"`
	Color  string       `json:"color"`
	Icon   string       `json:"icon,omitempty"`
	Action ButtonAction `json:"action"`
}

// ButtonConfig represents the complete button configuration
type ButtonConfig struct {
	Grid    GridConfig `json:"grid"`
	Buttons []Button   `json:"buttons"`
}

// LoadConfig loads button configuration from a JSON file
func LoadConfig(filename string) (*ButtonConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ButtonConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// SaveConfig saves button configuration to a JSON file
func SaveConfig(filename string, config *ButtonConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetButton returns a button by ID
func (c *ButtonConfig) GetButton(id string) *Button {
	for i := range c.Buttons {
		if c.Buttons[i].ID == id {
			return &c.Buttons[i]
		}
	}
	return nil
}

// UpdateButton updates or adds a button
func (c *ButtonConfig) UpdateButton(button Button) {
	for i := range c.Buttons {
		if c.Buttons[i].ID == button.ID {
			c.Buttons[i] = button
			return
		}
	}
	c.Buttons = append(c.Buttons, button)
}

// DeleteButton removes a button by ID
func (c *ButtonConfig) DeleteButton(id string) {
	for i := range c.Buttons {
		if c.Buttons[i].ID == id {
			c.Buttons = append(c.Buttons[:i], c.Buttons[i+1:]...)
			return
		}
	}
}
