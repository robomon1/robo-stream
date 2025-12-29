package main

import (
	"context"
	"fmt"
	"log"
	"net"
	neturl "net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/robomon1/robo-stream/server/internal/api"
	"github.com/robomon1/robo-stream/server/internal/manager"
	"github.com/robomon1/robo-stream/server/internal/models"
	"github.com/robomon1/robo-stream/server/internal/storage"
)

// App struct
type App struct {
	ctx            context.Context
	storage        *storage.Storage
	buttonManager  *manager.ButtonManager
	configManager  *manager.ConfigManager
	sessionManager *manager.SessionManager
	obsManager     *manager.OBSManager
	apiServer      *api.Server
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Get data directory
	dataDir, err := a.getDataDir()
	if err != nil {
		log.Fatal("Failed to get data directory:", err)
	}

	// Initialize storage
	a.storage, err = storage.New(dataDir)
	if err != nil {
		log.Fatal("Failed to initialize storage:", err)
	}

	// Initialize managers
	a.buttonManager = manager.NewButtonManager(a.storage)
	a.configManager = manager.NewConfigManager(a.storage, a.buttonManager)
	a.sessionManager = manager.NewSessionManager(a.storage)
	a.obsManager = manager.NewOBSManager()

	// Initialize with some default data if needed
	a.initializeDefaults()

	// Auto-connect to OBS on startup
	go func() {
		log.Println("🔌 Attempting to auto-connect to OBS...")
		savedConfig := a.GetSavedOBSConfig()
		
		// Try with saved config first
		err := a.obsManager.Connect(savedConfig.URL, savedConfig.Password)
		if err != nil {
			log.Printf("⚠️  Auto-connect to OBS failed: %v (this is normal if OBS isn't running)", err)
		} else {
			log.Println("✅ Auto-connected to OBS successfully!")
		}
	}()

	// Start API server for clients
	a.apiServer = api.NewServer(a.configManager, a.sessionManager, a.obsManager)
	go func() {
		log.Println("Starting API server on :8080")
		if err := a.apiServer.Start(":8080"); err != nil {
			log.Printf("API server error: %v", err)
		}
	}()

	log.Println("Robo-Stream Server started successfully")
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	if a.obsManager != nil {
		a.obsManager.Disconnect()
	}
	log.Println("Robo-Stream Server shutdown complete")
}

// getDataDir returns the appropriate data directory for the platform
func (a *App) getDataDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	var dataDir string
	switch {
	case filepath.Dir("/") == "/": // Unix-like
		dataDir = filepath.Join(homeDir, ".robo-stream-server")
	default: // Windows
		dataDir = filepath.Join(homeDir, "AppData", "Roaming", "RoboStreamServer")
	}

	return dataDir, nil
}

// initializeDefaults creates default configuration if none exist
func (a *App) initializeDefaults() {
	// Check if we have any configurations
	configs := a.configManager.List()
	if len(configs) > 0 {
		return // Already initialized
	}

	log.Println("Initializing default configuration...")

	// Create some default buttons
	defaultButtons := []struct {
		name        string
		description string
		icon        string
		color       string
		actionType  string
		params      map[string]interface{}
	}{
		{"Go Live", "Start streaming", "video", "#e74c3c", "start_stream", nil},
		{"Stop Stream", "Stop streaming", "stop-circle", "#95a5a6", "stop_stream", nil},
		{"Start Record", "Start recording", "circle", "#e74c3c", "start_record", nil},
		{"Stop Record", "Stop recording", "stop-circle", "#95a5a6", "stop_record", nil},
		{"Mute Mic", "Mute microphone", "mic-off", "#e67e22", "toggle_input_mute", map[string]interface{}{"input_name": "Mic/Aux"}},
		{"Main Scene", "Switch to main scene", "layout", "#3498db", "switch_scene", map[string]interface{}{"scene_name": "Main"}},
	}

	buttonIDs := make([]string, 0, len(defaultButtons))
	for _, btn := range defaultButtons {
		button := &models.Button{
			Name:        btn.name,
			Description: btn.description,
			Icon:        btn.icon,
			Color:       btn.color,
			Action: models.ButtonAction{
				Type:   btn.actionType,
				Params: btn.params,
			},
		}
		if err := a.buttonManager.Create(button); err != nil {
			log.Printf("Failed to create button %s: %v", btn.name, err)
			continue
		}
		buttonIDs = append(buttonIDs, button.ID)
	}

	// Create default configuration
	defaultConfig := &models.Configuration{
		Name:        "Default",
		Description: "Default configuration",
		Grid: models.GridConfig{
			Rows: 3,
			Cols: 4,
		},
		Buttons:   make(map[string]string),
		IsDefault: true,
	}

	// Assign buttons to positions
	positions := []string{"btn-0-0", "btn-0-1", "btn-1-0", "btn-1-1", "btn-2-0", "btn-2-1"}
	for i, btnID := range buttonIDs {
		if i < len(positions) {
			defaultConfig.Buttons[positions[i]] = btnID
		}
	}

	if err := a.configManager.Create(defaultConfig); err != nil {
		log.Printf("Failed to create default configuration: %v", err)
	}

	log.Println("Default configuration created successfully")
}

// ==================== WAILS BINDINGS ====================

// Button operations
func (a *App) GetButtons() []*models.Button {
	return a.buttonManager.List()
}

func (a *App) GetButton(id string) (*models.Button, error) {
	return a.buttonManager.Get(id)
}

func (a *App) CreateButton(button *models.Button) error {
	return a.buttonManager.Create(button)
}

func (a *App) UpdateButton(button *models.Button) error {
	return a.buttonManager.Update(button)
}

func (a *App) DeleteButton(id string) error {
	return a.buttonManager.Delete(id)
}

// Configuration operations
func (a *App) GetConfigurations() []*models.Configuration {
	return a.configManager.List()
}

func (a *App) GetConfiguration(id string) (*models.Configuration, error) {
	return a.configManager.Get(id)
}

func (a *App) CreateConfiguration(config *models.Configuration) error {
	return a.configManager.Create(config)
}

func (a *App) UpdateConfiguration(config *models.Configuration) error {
	return a.configManager.Update(config)
}

func (a *App) DeleteConfiguration(id string) error {
	return a.configManager.Delete(id)
}

func (a *App) SetDefaultConfiguration(id string) error {
	return a.configManager.SetDefault(id)
}

func (a *App) GetDefaultConfiguration() (*models.Configuration, error) {
	return a.configManager.GetDefault()
}

// Resolve configuration (get with full button details)
func (a *App) ResolveConfiguration(id string) (*models.ResolvedConfiguration, error) {
	return a.configManager.Resolve(id)
}

// Session operations
func (a *App) GetSessions() []*models.ClientSession {
	return a.sessionManager.List()
}

func (a *App) GetSession(sessionID string) (*models.ClientSession, error) {
	return a.sessionManager.Get(sessionID)
}

func (a *App) UpdateClientConfig(sessionID, configID string) error {
	return a.sessionManager.UpdateConfig(sessionID, configID)
}

// OBS operations
func (a *App) ConnectOBS(url, password string) error {
	log.Printf("🔌 ConnectOBS called with RAW url: %q", url)
	
	// Workaround: Wails might be double-encoding the URL
	// Decode it if needed
	decodedURL := url
	if strings.Contains(url, "%2F") {
		var err error
		decodedURL, err = neturl.QueryUnescape(url)
		if err != nil {
			log.Printf("⚠️  Failed to decode URL: %v", err)
		} else {
			log.Printf("🔧 Decoded URL from %q to %q", url, decodedURL)
		}
	}
	
	log.Printf("🔌 Connecting to: %s", decodedURL)
	err := a.obsManager.Connect(decodedURL, password)
	if err != nil {
		log.Printf("❌ ConnectOBS failed: %v", err)
		return err
	}
	log.Printf("✅ ConnectOBS succeeded")
	
	// Save credentials for next time
	config := &models.OBSConfig{
		URL:      decodedURL,
		Password: password,
	}
	if err := a.storage.SaveJSON("obs_config.json", config); err != nil {
		log.Printf("⚠️  Failed to save OBS config: %v", err)
	} else {
		log.Println("💾 OBS config saved")
	}
	
	return nil
}

func (a *App) DisconnectOBS() error {
	log.Println("🔌 DisconnectOBS called")
	return a.obsManager.Disconnect()
}

func (a *App) GetOBSStatus() map[string]interface{} {
	status := map[string]interface{}{
		"connected": a.obsManager.IsConnected(),
		"url":       a.obsManager.GetURL(),
	}
	log.Printf("📊 GetOBSStatus: %+v", status)
	return status
}

// GetSavedOBSConfig returns the saved OBS credentials
func (a *App) GetSavedOBSConfig() *models.OBSConfig {
	// Check environment variables first
	envURL := os.Getenv("OBS_WEBSOCKET_URL")
	envPassword := os.Getenv("OBS_WEBSOCKET_PASSWORD")
	
	if envURL != "" {
		log.Printf("📋 Using OBS config from environment variables")
		return &models.OBSConfig{
			URL:      envURL,
			Password: envPassword,
		}
	}
	
	// Fall back to saved config file
	var config models.OBSConfig
	if err := a.storage.LoadJSON("obs_config.json", &config); err != nil {
		log.Printf("📋 No saved OBS config found, using defaults")
		return &models.OBSConfig{
			URL:      "localhost:4455",
			Password: "",
		}
	}
	log.Printf("📋 Loaded saved OBS config: url=%s", config.URL)
	return &config
}

func (a *App) GetScenes() ([]string, error) {
	log.Println("📞 GetScenes() called from frontend")
	return a.obsManager.GetScenes()
}

func (a *App) GetInputs() ([]string, error) {
	log.Println("📞 GetInputs() called from frontend")
	return a.obsManager.GetInputs()
}

func (a *App) ExecuteAction(action models.ButtonAction) error {
	return a.obsManager.ExecuteAction(action)
}

// Test configuration by executing all actions in preview mode
func (a *App) TestConfiguration(configID string) error {
	config, err := a.configManager.Resolve(configID)
	if err != nil {
		return err
	}

	if !a.obsManager.IsConnected() {
		return fmt.Errorf("not connected to OBS")
	}

	log.Printf("Testing configuration: %s (%d buttons)", config.Name, len(config.Buttons))
	return nil
}

// getLocalIPs returns all non-loopback IPv4 addresses
func (a *App) getLocalIPs() []string {
	var ips []string
	
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("Failed to get network interfaces: %v", err)
		return ips
	}
	
	for _, iface := range interfaces {
		// Skip down interfaces
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		
		// Skip loopback
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			
			// Skip loopback and IPv6
			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}
			
			ips = append(ips, ip.String())
		}
	}
	
	return ips
}

// Get server info
func (a *App) GetServerInfo() map[string]interface{} {
	ips := a.getLocalIPs()
	clientURLs := make([]string, len(ips))
	for i, ip := range ips {
		clientURLs[i] = fmt.Sprintf("http://%s:8080", ip)
	}
	
	return map[string]interface{}{
		"version":         "1.0.0",
		"api_port":        8080,
		"ip_addresses":    ips,
		"client_urls":     clientURLs,
		"obs_connected":   a.obsManager.IsConnected(),
		"active_sessions": len(a.sessionManager.List()),
		"configurations":  len(a.configManager.List()),
		"buttons":         len(a.buttonManager.List()),
	}
}

// TestBinding - Simple test to verify Wails bindings are working
func (a *App) TestBinding(message string) string {
	response := fmt.Sprintf("✅ Wails binding works! You sent: %s", message)
	log.Println(response)
	return response
}
