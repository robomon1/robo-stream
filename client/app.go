package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"client/internal/client"
	"client/internal/config"

	"github.com/sirupsen/logrus"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx           context.Context
	apiClient     *client.APIClient
	configuration *config.ResolvedConfiguration
	logger        *logrus.Logger
	serverURL     string
	configDir     string
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
		configDir = filepath.Join(homeDir, "Library", "Application Support", "RoboStream-Client")
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(homeDir, "AppData", "Roaming")
		}
		configDir = filepath.Join(appData, "RoboStream-Client")
	default: // linux, freebsd, etc.
		xdgConfig := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfig == "" {
			xdgConfig = filepath.Join(homeDir, ".config")
		}
		configDir = filepath.Join(xdgConfig, "robo-stream-client")
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return configDir, nil
}

// loadServerURL loads the saved server URL
func loadServerURL(configDir string) string {
	urlFile := filepath.Join(configDir, "server_url.txt")
	data, err := os.ReadFile(urlFile)
	if err != nil {
		return "http://localhost:8080"
	}
	return string(data)
}

// saveServerURL saves the server URL
func saveServerURL(configDir, url string) error {
	urlFile := filepath.Join(configDir, "server_url.txt")
	return os.WriteFile(urlFile, []byte(url), 0644)
}

// NewApp creates a new App application struct
func NewApp() *App {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	configDir, err := getConfigDir()
	if err != nil {
		logger.Fatalf("Could not get config directory: %v", err)
	}

	logger.Infof("Config directory: %s", configDir)

	return &App{
		logger:    logger,
		configDir: configDir,
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.serverURL = loadServerURL(a.configDir)

	a.logger.Infof("Robo-Stream Client starting...")
	a.logger.Infof("Platform: %s/%s", runtime.GOOS, runtime.GOARCH)
	a.logger.Infof("Server URL: %s", a.serverURL)

	a.apiClient = client.NewAPIClient(a.serverURL, a.logger, a.configDir)
	go a.connectAndLoad()
}

// shutdown is called when the app shuts down
func (a *App) shutdown(ctx context.Context) {
	a.logger.Info("Shutting down...")
}

// connectAndLoad connects to server and loads configuration
func (a *App) connectAndLoad() {
	info, err := a.apiClient.GetServerInfo()
	if err != nil {
		a.logger.Errorf("Failed to connect to server: %v", err)
		wailsruntime.EventsEmit(a.ctx, "connection_error", err.Error())
		return
	}

	a.logger.Infof("Connected to server: %v", info)
	wailsruntime.EventsEmit(a.ctx, "connected", info)

	// Register with server and get default configuration
	resolved, err := a.apiClient.Register()
	if err != nil {
		a.logger.Errorf("Failed to register with server: %v", err)
		wailsruntime.EventsEmit(a.ctx, "config_error", err.Error())
		return
	}

	a.configuration = resolved

	a.logger.Infof("Loaded configuration: %s (%dx%d grid, %d buttons)",
		resolved.Name, resolved.Grid.Rows, resolved.Grid.Cols, len(resolved.Buttons))

	wailsruntime.EventsEmit(a.ctx, "configuration_loaded", resolved)
}

// GetConfiguration returns the current configuration
func (a *App) GetConfiguration() *config.ResolvedConfiguration {
	return a.configuration
}

// GetConfigurations returns all available configurations
func (a *App) GetConfigurations() ([]config.Configuration, error) {
	configs, err := a.apiClient.GetConfigurations()
	if err != nil {
		a.logger.Errorf("Failed to get configurations: %v", err)
		return nil, err
	}
	return configs, nil
}

// LoadConfiguration loads a specific configuration
func (a *App) LoadConfiguration(configID string) error {
	resolved, err := a.apiClient.GetConfiguration(configID)
	if err != nil {
		a.logger.Errorf("Failed to load configuration: %v", err)
		return err
	}

	a.configuration = resolved

	a.logger.Infof("Loaded configuration: %s", resolved.Name)
	wailsruntime.EventsEmit(a.ctx, "configuration_loaded", resolved)

	return nil
}

// PressButton executes a button action
// Position is in format "btn-0-0" (btn-row-col)
func (a *App) PressButton(position string) error {
	if a.configuration == nil {
		return fmt.Errorf("no configuration loaded")
	}

	// Parse position string "btn-0-0" to row and col
	parts := strings.Split(position, "-")
	if len(parts) != 3 {
		return fmt.Errorf("invalid position format: %s", position)
	}

	row, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid row in position: %s", position)
	}

	col, err := strconv.Atoi(parts[2])
	if err != nil {
		return fmt.Errorf("invalid col in position: %s", position)
	}

	// Find button at this position
	button := a.configuration.GetButtonAt(row, col)
	if button == nil {
		return fmt.Errorf("no button at position: %s", position)
	}

	a.logger.Infof("Button pressed: %s (action: %s)", button.Text, button.Action.Type)

	err = a.apiClient.ExecuteAction(button.Action)
	if err != nil {
		a.logger.Errorf("Failed to execute action: %v", err)
		return err
	}

	go a.emitStatusUpdate()
	return nil
}

// GetStatus returns current OBS status
func (a *App) GetStatus() (map[string]interface{}, error) {
	return a.apiClient.GetOBSStatus()
}

// GetServerURL returns the configured server URL
func (a *App) GetServerURL() string {
	return a.serverURL
}

// GetCurrentConfiguration returns the currently loaded configuration
func (a *App) GetCurrentConfiguration() *config.ResolvedConfiguration {
	return a.configuration
}

// GetOBSStatus returns current OBS status (recording, streaming, etc)
func (a *App) GetOBSStatus() (map[string]interface{}, error) {
	status, err := a.apiClient.GetOBSStatus()
	if err != nil {
		return map[string]interface{}{
			"connected": false,
			"streaming": false,
			"recording": false,
		}, err
	}
	return status, nil
}

// SetServerURL sets a new server URL and reconnects
func (a *App) SetServerURL(url string) error {
	a.serverURL = url
	a.apiClient = client.NewAPIClient(url, a.logger, a.configDir)

	if err := saveServerURL(a.configDir, url); err != nil {
		a.logger.Warnf("Failed to save server URL: %v", err)
	}

	a.logger.Infof("Server URL updated: %s", url)
	go a.connectAndLoad()
	return nil
}

// Reconnect attempts to reconnect to the server
func (a *App) Reconnect() error {
	go a.connectAndLoad()
	return nil
}

// ToggleFullscreen toggles fullscreen mode
func (a *App) ToggleFullscreen() {
	wailsruntime.WindowToggleMaximise(a.ctx)
}

// emitStatusUpdate fetches status and emits to frontend
func (a *App) emitStatusUpdate() {
	status, err := a.apiClient.GetOBSStatus()
	if err != nil {
		a.logger.Errorf("Failed to get OBS status: %v", err)
		return
	}
	wailsruntime.EventsEmit(a.ctx, "status_update", status)
}
