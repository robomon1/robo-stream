package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/robomon1/stream-pi-go/client-go/internal/client"
	"github.com/robomon1/stream-pi-go/client-go/internal/config"
	"github.com/sirupsen/logrus"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx        context.Context
	obsClient  *client.OBSClient
	config     *config.ButtonConfig
	logger     *logrus.Logger
	serverURL  string
	configPath string
}

// getConfigDir returns the OS-appropriate config directory
func getConfigDir() (string, error) {
	var configDir string
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "darwin": // macOS
		configDir = filepath.Join(homeDir, "Library", "Application Support", "StreamPi")
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(homeDir, "AppData", "Roaming")
		}
		configDir = filepath.Join(appData, "StreamPi")
	default: // linux, freebsd, etc.
		xdgConfig := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfig == "" {
			xdgConfig = filepath.Join(homeDir, ".config")
		}
		configDir = filepath.Join(xdgConfig, "streampi")
	}

	// Ensure directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return configDir, nil
}

// NewApp creates a new App application struct
func NewApp() *App {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Get OS-appropriate config directory
	configDir, err := getConfigDir()
	if err != nil {
		logger.Fatalf("Could not get config directory: %v", err)
	}
	
	configPath := filepath.Join(configDir, "buttons.json")
	logger.Infof("Config directory: %s", configDir)

	return &App{
		logger:     logger,
		configPath: configPath,
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	
	// Get server URL from environment or use default
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}
	a.serverURL = serverURL

	a.logger.Infof("Stream-Pi Deck starting...")
	a.logger.Infof("Platform: %s/%s", runtime.GOOS, runtime.GOARCH)
	a.logger.Infof("Connecting to server: %s", serverURL)
	a.logger.Infof("Config file: %s", a.configPath)
	
	// Create OBS client
	a.obsClient = client.NewOBSClient(serverURL, a.logger)

	// Load configuration
	loadedConfig, err := config.LoadConfig(a.configPath)
	if err != nil {
		a.logger.Warnf("Failed to load config: %v, using default", err)
		loadedConfig = a.getDefaultConfig()
		if err := config.SaveConfig(a.configPath, loadedConfig); err != nil {
			a.logger.Errorf("Failed to save default config: %v", err)
		}
	}
	a.config = loadedConfig

	a.logger.Infof("Loaded %d buttons in %dx%d grid", len(a.config.Buttons), a.config.Grid.Rows, a.config.Grid.Cols)
}

// shutdown is called when the app shuts down
func (a *App) shutdown(ctx context.Context) {
	a.logger.Info("Shutting down...")
}

// GetConfig returns the current button configuration
func (a *App) GetConfig() *config.ButtonConfig {
	return a.config
}

// GetScenes returns available OBS scenes
func (a *App) GetScenes() ([]string, error) {
	return a.obsClient.GetScenes()
}

// GetInputs returns available OBS inputs
func (a *App) GetInputs() ([]string, error) {
	return a.obsClient.GetInputs()
}

// GetStatus returns current OBS status
func (a *App) GetStatus() (map[string]interface{}, error) {
	return a.obsClient.GetStatus()
}

// PressButton executes a button action
func (a *App) PressButton(buttonID string) error {
	button := a.config.GetButton(buttonID)
	if button == nil {
		return fmt.Errorf("button not found: %s", buttonID)
	}

	a.logger.Infof("Button pressed: %s (action: %s)", button.Text, button.Action.Type)

	_, err := a.obsClient.ExecuteAction(button.Action.Type, button.Action.Params)
	if err != nil {
		a.logger.Errorf("Failed to execute action: %v", err)
		return err
	}

	// Emit event to update status
	go a.emitStatusUpdate()

	return nil
}

// SaveButton saves a button configuration
func (a *App) SaveButton(buttonData string) error {
	a.logger.Infof("SaveButton called")
	
	var button config.Button
	if err := json.Unmarshal([]byte(buttonData), &button); err != nil {
		a.logger.Errorf("Failed to parse button: %v", err)
		return fmt.Errorf("failed to parse button: %w", err)
	}

	a.config.UpdateButton(button)

	if err := config.SaveConfig(a.configPath, a.config); err != nil {
		a.logger.Errorf("Failed to save config: %v", err)
		return fmt.Errorf("failed to save config: %w", err)
	}

	a.logger.Infof("Button saved: %s", button.ID)
	return nil
}

// DeleteButton deletes a button
func (a *App) DeleteButton(buttonID string) error {
	a.logger.Infof("DeleteButton called for: %s", buttonID)
	
	a.config.DeleteButton(buttonID)

	if err := config.SaveConfig(a.configPath, a.config); err != nil {
		a.logger.Errorf("Failed to save config after delete: %v", err)
		return fmt.Errorf("failed to save config: %w", err)
	}

	a.logger.Infof("Button deleted: %s", buttonID)
	return nil
}

// UpdateGrid updates the grid configuration
func (a *App) UpdateGrid(rows, cols int) error {
	a.logger.Infof("UpdateGrid called: %dx%d", rows, cols)
	
	a.config.Grid.Rows = rows
	a.config.Grid.Cols = cols

	if err := config.SaveConfig(a.configPath, a.config); err != nil {
		a.logger.Errorf("Failed to save config after grid update: %v", err)
		return fmt.Errorf("failed to save config: %w", err)
	}

	a.logger.Infof("Grid updated: %dx%d", rows, cols)
	return nil
}

// ToggleFullscreen toggles fullscreen mode
func (a *App) ToggleFullscreen() {
	wailsruntime.WindowToggleMaximise(a.ctx)
}

// GetServerURL returns the configured server URL
func (a *App) GetServerURL() string {
	return a.serverURL
}

// SetServerURL sets a new server URL
func (a *App) SetServerURL(url string) error {
	a.serverURL = url
	a.obsClient = client.NewOBSClient(url, a.logger)
	a.logger.Infof("Server URL updated: %s", url)
	return nil
}

// GetConfigPath returns the config file path
func (a *App) GetConfigPath() string {
	return a.configPath
}

// emitStatusUpdate fetches status and emits to frontend
func (a *App) emitStatusUpdate() {
	status, err := a.obsClient.GetStatus()
	if err != nil {
		a.logger.Errorf("Failed to get status: %v", err)
		return
	}

	wailsruntime.EventsEmit(a.ctx, "status_update", status)
}

func (a *App) getDefaultConfig() *config.ButtonConfig {
	return &config.ButtonConfig{
		Grid: config.GridConfig{
			Rows: 3,
			Cols: 4,
		},
		Buttons: []config.Button{
			{
				ID:    "btn-0-0",
				Row:   0,
				Col:   0,
				Text:  "Stream",
				Color: "#e74c3c",
				Action: config.ButtonAction{
					Type: "toggle_stream",
				},
			},
			{
				ID:    "btn-0-1",
				Row:   0,
				Col:   1,
				Text:  "Record",
				Color: "#e67e22",
				Action: config.ButtonAction{
					Type: "toggle_record",
				},
			},
		},
	}
}
