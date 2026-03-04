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
	"time"

	"github.com/robomon1/robo-stream/server/internal/api"
	"github.com/robomon1/robo-stream/server/internal/controller"
	obsctrl "github.com/robomon1/robo-stream/server/internal/controller/obs"
	"github.com/robomon1/robo-stream/server/internal/manager"
	"github.com/robomon1/robo-stream/server/internal/models"
	"github.com/robomon1/robo-stream/server/internal/plugin"
	"github.com/robomon1/robo-stream/server/internal/storage"
)

// App is the root application struct exposed to the Wails runtime.
type App struct {
	ctx            context.Context
	storage        *storage.Storage
	buttonManager  *manager.ButtonManager
	configManager  *manager.ConfigManager
	sessionManager *manager.SessionManager
	registry       *controller.Registry
	obsController  *obsctrl.Controller
	pluginManager  *plugin.Manager
	apiServer      *api.Server
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{}
}

// startup is called when the Wails application starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	dataDir, err := a.getDataDir()
	if err != nil {
		log.Fatal("Failed to get data directory:", err)
	}

	a.storage, err = storage.New(dataDir)
	if err != nil {
		log.Fatal("Failed to initialize storage:", err)
	}

	// Core managers (unchanged)
	a.buttonManager = manager.NewButtonManager(a.storage)
	a.configManager = manager.NewConfigManager(a.storage, a.buttonManager)
	a.sessionManager = manager.NewSessionManager(a.storage)

	// Controller registry with built-in OBS controller
	a.registry = controller.NewRegistry()
	a.obsController = obsctrl.New(a.storage)
	if err := a.registry.Register(a.obsController); err != nil {
		log.Fatal("Failed to register OBS controller:", err)
	}

	// Plugin manager discovers and starts external controller plugins
	pluginsDir := filepath.Join(dataDir, "plugins")
	a.pluginManager = plugin.New(pluginsDir, a.storage, a.registry)

	a.initializeDefaults()

	go a.sessionCleanupLoop()

	// Discover and start any installed plugins
	go func() {
		if err := a.pluginManager.DiscoverAndStart(); err != nil {
			log.Printf("⚠️  Plugin discovery error: %v", err)
		}
	}()

	// Auto-connect OBS with saved credentials
	go func() {
		log.Println("🔌 Attempting to auto-connect to OBS...")
		cfg := a.obsController.GetCurrentConfig()
		url, _ := cfg["url"].(string)
		password, _ := cfg["password"].(string)
		if err := a.obsController.Connect(url, password); err != nil {
			log.Printf("⚠️  Auto-connect to OBS failed: %v (this is normal if OBS isn't running)", err)
		} else {
			log.Println("✅ Auto-connected to OBS successfully!")
		}
	}()

	// Start HTTP API server
	a.apiServer = api.NewServer(a.configManager, a.sessionManager, a.registry, a.obsController, a.pluginManager)
	go func() {
		log.Println("Starting API server on 0.0.0.0:8080")
		if err := a.apiServer.Start("0.0.0.0:8080"); err != nil {
			log.Printf("API server error: %v", err)
		}
	}()

	log.Println("Robo-Stream Server started successfully")
}

func (a *App) sessionCleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	inactiveTimeout := 30 * time.Minute
	cleanup := func() {
		if err := a.sessionManager.CleanupInactive(inactiveTimeout); err != nil {
			log.Printf("⚠️  Failed to cleanup inactive sessions: %v", err)
		} else {
			log.Printf("🧹 Session cleanup complete (%d active sessions)", len(a.sessionManager.List()))
		}
	}
	cleanup()
	for range ticker.C {
		cleanup()
	}
}

// shutdown is called when the Wails application closes.
func (a *App) shutdown(ctx context.Context) {
	a.registry.ShutdownAll()
	a.pluginManager.Shutdown()
	log.Println("Robo-Stream Server shutdown complete")
}

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

// initializeDefaults creates the default OBS configuration on first run.
func (a *App) initializeDefaults() {
	if len(a.configManager.List()) > 0 {
		return
	}
	log.Println("Initializing default configuration...")

	buttonIDs := make([]string, 0)
	for _, btn := range a.obsController.GetDefaultButtons() {
		if err := a.buttonManager.Create(btn); err != nil {
			log.Printf("Failed to create default button %s: %v", btn.Name, err)
			continue
		}
		buttonIDs = append(buttonIDs, btn.ID)
	}

	defaultConfig := &models.Configuration{
		Name:        "Default",
		Description: "Default configuration",
		Grid:        models.GridConfig{Rows: 3, Cols: 4},
		Buttons:     make(map[string]string),
		IsDefault:   true,
	}
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

// ----- Button operations -----

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

// ----- Configuration operations -----

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

func (a *App) ResolveConfiguration(id string) (*models.ResolvedConfiguration, error) {
	return a.configManager.Resolve(id)
}

// ----- Session operations -----

func (a *App) GetSessions() []*models.ClientSession {
	return a.sessionManager.List()
}

func (a *App) GetSession(sessionID string) (*models.ClientSession, error) {
	return a.sessionManager.Get(sessionID)
}

func (a *App) UpdateClientConfig(sessionID, configID string) error {
	return a.sessionManager.UpdateConfig(sessionID, configID)
}

// ----- Controller operations (generic) -----

// GetControllers returns status info for all registered controllers.
func (a *App) GetControllers() []map[string]interface{} {
	controllers := a.registry.List()
	result := make([]map[string]interface{}, len(controllers))
	for i, c := range controllers {
		result[i] = map[string]interface{}{
			"id":          c.ID(),
			"name":        c.Name(),
			"description": c.Description(),
			"version":     c.Version(),
			"connected":   c.IsConnected(),
		}
	}
	return result
}

// GetControllerStatus returns the status map for the named controller.
func (a *App) GetControllerStatus(id string) map[string]interface{} {
	c, ok := a.registry.Get(id)
	if !ok {
		return map[string]interface{}{"error": fmt.Sprintf("controller %q not found", id)}
	}
	return c.GetStatus()
}

// GetControllerConfig returns the current (non-sensitive) configuration for the named controller.
func (a *App) GetControllerConfig(id string) map[string]interface{} {
	c, ok := a.registry.Get(id)
	if !ok {
		return nil
	}
	return c.GetCurrentConfig()
}

// GetControllerConfigSchema returns the config field definitions for the named controller.
func (a *App) GetControllerConfigSchema(id string) []controller.ConfigField {
	c, ok := a.registry.Get(id)
	if !ok {
		return nil
	}
	return c.GetConfigSchema()
}

// ConnectController connects or reconfigures the named controller.
// For OBS, config should include "url" and optionally "password".
// For plugin controllers, config is saved and forwarded via /initialize.
func (a *App) ConnectController(id string, config map[string]interface{}) error {
	switch id {
	case "obs":
		return a.connectOBS(config)
	default:
		return a.pluginManager.UpdateConfig(id, config)
	}
}

// connectOBS handles the OBS-specific connection flow.
func (a *App) connectOBS(config map[string]interface{}) error {
	url, _ := config["url"].(string)
	password, _ := config["password"].(string)

	log.Printf("🔌 ConnectController(obs) called with url: %q", url)

	// Wails sometimes double-encodes URLs
	if strings.Contains(url, "%2F") {
		if decoded, err := neturl.QueryUnescape(url); err == nil {
			log.Printf("🔧 Decoded URL from %q to %q", url, decoded)
			url = decoded
		}
	}

	if err := a.obsController.Connect(url, password); err != nil {
		log.Printf("❌ OBS connect failed: %v", err)
		return err
	}
	log.Printf("✅ OBS connected to %s", url)

	// Persist credentials
	if err := a.obsController.SaveConfig(map[string]interface{}{"url": url, "password": password}); err != nil {
		log.Printf("⚠️  Failed to save OBS config: %v", err)
	}
	return nil
}

// DisconnectController disconnects the named controller.
func (a *App) DisconnectController(id string) error {
	log.Printf("🔌 DisconnectController(%s) called", id)
	c, ok := a.registry.Get(id)
	if !ok {
		return fmt.Errorf("controller %q not found", id)
	}
	return c.Shutdown()
}

// ExecuteAction routes a button action through the controller registry.
func (a *App) ExecuteAction(action models.ButtonAction) error {
	return a.registry.ExecuteAction(action)
}

// GetConfigurationButtonIndicators returns per-button indicator classes for all
// buttons in the given configuration.
func (a *App) GetConfigurationButtonIndicators(configID string) (map[string]string, error) {
	cfg, err := a.configManager.Get(configID)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	buttonActions := make(map[string]models.ButtonAction, len(cfg.Buttons))
	for _, buttonID := range cfg.Buttons {
		btn, err := a.buttonManager.Get(buttonID)
		if err != nil {
			continue
		}
		buttonActions[buttonID] = btn.Action
		ctrlID := btn.Action.Controller
		if ctrlID == "" {
			ctrlID = "obs"
		}
		seen[ctrlID] = true
	}

	for ctrlID := range seen {
		if ctrl, ok := a.registry.Get(ctrlID); ok {
			ctrl.GetStatus()
		}
	}

	indicators := make(map[string]string, len(buttonActions))
	for buttonID, action := range buttonActions {
		ctrlID := action.Controller
		if ctrlID == "" {
			ctrlID = "obs"
		}
		if ctrl, ok := a.registry.Get(ctrlID); ok {
			indicators[buttonID] = ctrl.ComputeIndicator(action)
		} else {
			indicators[buttonID] = ""
		}
	}
	return indicators, nil
}

// ControllerActionGroup groups a controller's supported action types for the button editor.
type ControllerActionGroup struct {
	ControllerID   string                            `json:"controller_id"`
	ControllerName string                            `json:"controller_name"`
	Connected      bool                              `json:"connected"`
	Actions        []controller.ActionTypeDefinition `json:"actions"`
}

// GetAllActionTypes returns supported action types from every registered controller,
// grouped by controller. Used by the button editor to populate the action picker.
func (a *App) GetAllActionTypes() []ControllerActionGroup {
	controllers := a.registry.List()
	result := make([]ControllerActionGroup, 0, len(controllers))
	for _, c := range controllers {
		result = append(result, ControllerActionGroup{
			ControllerID:   c.ID(),
			ControllerName: c.Name(),
			Connected:      c.IsConnected(),
			Actions:        c.SupportedActionTypes(),
		})
	}
	return result
}

// ----- OBS-specific query bindings -----
// These are kept as named methods because the Wails frontend has OBS-specific
// settings panels that query scenes, inputs, and source states.

func (a *App) GetOBSStatus() map[string]interface{} {
	return a.obsController.GetStatus()
}

func (a *App) GetScenes() ([]string, error) {
	log.Println("📞 GetScenes() called from frontend")
	return a.obsController.GetScenes()
}

func (a *App) GetInputs() ([]string, error) {
	log.Println("📞 GetInputs() called from frontend")
	return a.obsController.GetInputs()
}

func (a *App) GetSourceVisibility(sceneName, sourceName string) (bool, error) {
	return a.obsController.GetSourceVisibility(sceneName, sourceName)
}

// ----- Plugin management bindings -----

// InstallPlugin copies the binary at binaryPath into the plugins directory
// and starts the plugin. Call this from the Wails file-picker handler.
func (a *App) InstallPlugin(binaryPath string) error {
	log.Printf("📦 Installing plugin from: %s", binaryPath)
	return a.pluginManager.Install(binaryPath)
}

// GetInstalledPlugins returns info about all running plugin processes.
func (a *App) GetInstalledPlugins() []map[string]interface{} {
	return a.pluginManager.List()
}

// UninstallPlugin stops and removes the named plugin.
func (a *App) UninstallPlugin(id string) error {
	return a.pluginManager.Uninstall(id)
}

// ----- Server info -----

func (a *App) GetServerInfo() map[string]interface{} {
	ips := a.getLocalIPs()
	clientURLs := make([]string, len(ips))
	for i, ip := range ips {
		clientURLs[i] = fmt.Sprintf("http://%s:8080", ip)
	}

	controllerStatus := make(map[string]bool)
	for _, c := range a.registry.List() {
		controllerStatus[c.ID()] = c.IsConnected()
	}

	return map[string]interface{}{
		"version":         Version,
		"api_port":        8080,
		"ip_addresses":    ips,
		"client_urls":     clientURLs,
		"obs_connected":   a.obsController.IsConnected(),
		"controllers":     controllerStatus,
		"active_sessions": len(a.sessionManager.List()),
		"configurations":  len(a.configManager.List()),
		"buttons":         len(a.buttonManager.List()),
	}
}

func (a *App) TestBinding(message string) string {
	response := fmt.Sprintf("✅ Wails binding works! You sent: %s", message)
	log.Println(response)
	return response
}

func (a *App) TestConfiguration(configID string) error {
	config, err := a.configManager.Resolve(configID)
	if err != nil {
		return err
	}
	if !a.obsController.IsConnected() {
		return fmt.Errorf("not connected to OBS")
	}
	log.Printf("Testing configuration: %s (%d buttons)", config.Name, len(config.Buttons))
	return nil
}

func (a *App) getLocalIPs() []string {
	var ips []string
	interfaces, err := net.Interfaces()
	if err != nil {
		return ips
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
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
			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}
			ips = append(ips, ip.String())
		}
	}
	return ips
}
