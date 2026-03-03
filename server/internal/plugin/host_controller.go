package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robomon1/robo-stream/server/internal/controller"
	"github.com/robomon1/robo-stream/server/internal/models"
	"github.com/robomon1/robo-stream/server/internal/storage"
)

// Local mirror types for the plugin HTTP protocol.
// These match the JSON schema defined in sdk/types.go without importing
// the sdk module from the server.

type pluginInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type pluginStatus struct {
	Connected bool                   `json:"connected"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

type pluginActionDef struct {
	Type            string           `json:"type"`
	Name            string           `json:"name"`
	Description     string           `json:"description,omitempty"`
	Params          []pluginParamDef `json:"params,omitempty"`
	IndicatorField  string           `json:"indicator_field,omitempty"`
	IndicatorInvert bool             `json:"indicator_invert,omitempty"`
}

type pluginParamDef struct {
	Key      string      `json:"key"`
	Label    string      `json:"label"`
	Type     string      `json:"type"`
	Options  []string    `json:"options,omitempty"`
	Required bool        `json:"required,omitempty"`
	Default  interface{} `json:"default,omitempty"`
}

type pluginConfigField struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Type     string `json:"type"`
	Required bool   `json:"required,omitempty"`
	Default  string `json:"default,omitempty"`
	Help     string `json:"help,omitempty"`
}

type pluginButton struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Icon        string          `json:"icon,omitempty"`
	Color       string          `json:"color,omitempty"`
	Action      pluginActionRef `json:"action"`
}

type pluginActionRef struct {
	Controller string                 `json:"controller"`
	Type       string                 `json:"type"`
	Params     map[string]interface{} `json:"params,omitempty"`
}

// HostController wraps a plugin subprocess and implements controller.Controller
// by making HTTP calls to the plugin's local server.
type HostController struct {
	id         string
	port       int
	pluginDir  string
	storage    *storage.Storage
	httpClient *http.Client
	// statusClient uses a shorter timeout for frequent health-check calls.
	statusClient *http.Client
	// Cached metadata fetched from the plugin on creation
	info pluginInfo
	// Cached connected state — updated by GetStatus(); read by IsConnected()
	// so that GetControllers() never has to make a blocking HTTP call.
	connectedMu sync.RWMutex
	connected   bool

	// statusCacheMu guards statusCache and statusCachedAt.
	statusCacheMu  sync.RWMutex
	statusCache    map[string]interface{}
	statusCachedAt time.Time

	// actionDefsMu guards actionDefsCache and actionDefsLoadedAt.
	// Action defs are fetched from the plugin's /actions endpoint and cached
	// so that ComputeIndicator doesn't need to make an HTTP round-trip per button.
	actionDefsMu      sync.RWMutex
	actionDefsCache   map[string]pluginActionDef // keyed by action type string
	actionDefsLoadedAt time.Time
}

// actionDefsCacheTTL is how long action definitions are cached before re-fetching.
const actionDefsCacheTTL = 60 * time.Second

func newHostController(id string, port int, pluginDir string, st *storage.Storage) *HostController {
	h := &HostController{
		id:        id,
		port:      port,
		pluginDir: pluginDir,
		storage:   st,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		// Status/health checks use a tighter timeout so frequent polls fail fast
		// rather than blocking for 10 s when the plugin is unresponsive.
		statusClient: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
	// Best-effort: fetch info on creation (not required for operation)
	if info, err := h.fetchInfo(); err == nil {
		h.info = info
	} else {
		h.info = pluginInfo{ID: id}
	}
	return h
}

func (h *HostController) baseURL() string {
	return fmt.Sprintf("http://127.0.0.1:%d", h.port)
}

// ==================== controller.Controller interface ====================

func (h *HostController) ID() string          { return h.info.ID }
func (h *HostController) Name() string        { return h.info.Name }
func (h *HostController) Description() string { return h.info.Description }
func (h *HostController) Version() string     { return h.info.Version }

func (h *HostController) Shutdown() error {
	// Plugin process lifecycle is managed by Manager; nothing to do here.
	return nil
}

// IsConnected returns the last-known connected state without making any HTTP
// calls. The cached value is refreshed by GetStatus(), which the polling loop
// calls every few seconds. This keeps GetControllers() instant.
func (h *HostController) IsConnected() bool {
	h.connectedMu.RLock()
	defer h.connectedMu.RUnlock()
	return h.connected
}

func (h *HostController) GetStatus() map[string]interface{} {
	var status pluginStatus
	// Use the short-timeout statusClient so stale/crashed plugins fail fast.
	if err := h.getWith(h.statusClient, "/status", &status); err != nil {
		h.connectedMu.Lock()
		h.connected = false
		h.connectedMu.Unlock()
		disconnected := map[string]interface{}{"connected": false, "error": err.Error()}
		h.statusCacheMu.Lock()
		h.statusCache = disconnected
		h.statusCachedAt = time.Now()
		h.statusCacheMu.Unlock()
		return disconnected
	}
	h.connectedMu.Lock()
	h.connected = status.Connected
	h.connectedMu.Unlock()

	result := map[string]interface{}{"connected": status.Connected}
	if status.Error != "" {
		result["error"] = status.Error
	}
	for k, v := range status.Details {
		result[k] = v
	}

	// Cache the result for use by ComputeIndicator.
	h.statusCacheMu.Lock()
	h.statusCache = result
	h.statusCachedAt = time.Now()
	h.statusCacheMu.Unlock()

	return result
}

// cachedStatus returns the most recent status map without making any HTTP calls.
func (h *HostController) cachedStatus() map[string]interface{} {
	h.statusCacheMu.RLock()
	defer h.statusCacheMu.RUnlock()
	if h.statusCache != nil {
		return h.statusCache
	}
	return map[string]interface{}{}
}

// getActionDefs returns the plugin's action type definitions, using a 60-second
// in-memory cache to avoid hitting the plugin's /actions endpoint on every
// ComputeIndicator call.
func (h *HostController) getActionDefs() map[string]pluginActionDef {
	h.actionDefsMu.RLock()
	if h.actionDefsCache != nil && time.Since(h.actionDefsLoadedAt) < actionDefsCacheTTL {
		defs := h.actionDefsCache
		h.actionDefsMu.RUnlock()
		return defs
	}
	h.actionDefsMu.RUnlock()

	// Fetch fresh definitions from the plugin.
	var rawDefs []pluginActionDef
	if err := h.get("/actions", &rawDefs); err != nil {
		// Return stale cache if available.
		h.actionDefsMu.RLock()
		defs := h.actionDefsCache
		h.actionDefsMu.RUnlock()
		return defs
	}

	defs := make(map[string]pluginActionDef, len(rawDefs))
	for _, d := range rawDefs {
		defs[d.Type] = d
	}

	h.actionDefsMu.Lock()
	h.actionDefsCache = defs
	h.actionDefsLoadedAt = time.Now()
	h.actionDefsMu.Unlock()

	return defs
}

// ComputeIndicator returns the CSS class for the given button action based on
// the plugin's last-known status. Uses the IndicatorField / IndicatorInvert
// rules declared in the plugin's SupportedActions definitions.
func (h *HostController) ComputeIndicator(action models.ButtonAction) string {
	defs := h.getActionDefs()
	if defs == nil {
		return ""
	}
	def, ok := defs[action.Type]
	if !ok || def.IndicatorField == "" {
		return ""
	}

	status := h.cachedStatus()

	// CRITICAL: guard against Go zero-value booleans when the plugin has not
	// yet received real state feedback from the external service. ZoomOSC (and
	// similar plugins) only emit state-change events — they do NOT broadcast
	// current state on connect. If the server starts mid-meeting, fields that
	// haven't changed remain at their zero value (false) even though they may
	// actually be true.
	//
	// Strategy: prefer a field-specific "{field}_known" flag if the plugin
	// provides one (e.g. "video_known" for the "video" IndicatorField). Fall
	// back to the global "state_known" flag for plugins that only track
	// overall state. Return "" for any field that hasn't been confirmed yet.
	fieldKnownKey := def.IndicatorField + "_known"
	if fieldKnownRaw, exists := status[fieldKnownKey]; exists {
		if fieldKnown, isBool := fieldKnownRaw.(bool); isBool && !fieldKnown {
			return ""
		}
	} else if stateKnownRaw, exists := status["state_known"]; exists {
		if stateKnown, isBool := stateKnownRaw.(bool); isBool && !stateKnown {
			return ""
		}
	}

	val, exists := status[def.IndicatorField]
	if !exists {
		return ""
	}

	boolVal, ok := val.(bool)
	if !ok {
		return ""
	}

	isActive := boolVal
	if def.IndicatorInvert {
		isActive = !isActive
	}
	if isActive {
		return "active"
	}
	return ""
}

func (h *HostController) SupportedActionTypes() []controller.ActionTypeDefinition {
	var defs []pluginActionDef
	if err := h.get("/actions", &defs); err != nil {
		return nil
	}
	result := make([]controller.ActionTypeDefinition, len(defs))
	for i, d := range defs {
		params := make([]controller.ParamDef, len(d.Params))
		for j, p := range d.Params {
			params[j] = controller.ParamDef{
				Key:      p.Key,
				Label:    p.Label,
				Type:     p.Type,
				Options:  p.Options,
				Required: p.Required,
				Default:  p.Default,
			}
		}
		result[i] = controller.ActionTypeDefinition{
			Type:            d.Type,
			Name:            d.Name,
			Description:     d.Description,
			Params:          params,
			IndicatorField:  d.IndicatorField,
			IndicatorInvert: d.IndicatorInvert,
		}
	}
	return result
}

func (h *HostController) GetConfigSchema() []controller.ConfigField {
	var fields []pluginConfigField
	if err := h.get("/config/schema", &fields); err != nil {
		return nil
	}
	result := make([]controller.ConfigField, len(fields))
	for i, f := range fields {
		result[i] = controller.ConfigField{
			Key:      f.Key,
			Label:    f.Label,
			Type:     f.Type,
			Required: f.Required,
			Default:  f.Default,
			Help:     f.Help,
		}
	}
	return result
}

func (h *HostController) GetCurrentConfig() map[string]interface{} {
	cfg := h.loadPersistedConfig()
	return cfg
}

func (h *HostController) SaveConfig(cfg map[string]interface{}) error {
	if err := h.persistConfig(cfg); err != nil {
		return err
	}
	return h.initialize(cfg)
}

func (h *HostController) GetDefaultButtons() []*models.Button {
	var buttons []pluginButton
	if err := h.get("/buttons", &buttons); err != nil {
		return nil
	}
	result := make([]*models.Button, len(buttons))
	for i, b := range buttons {
		result[i] = &models.Button{
			ID:          uuid.New().String(),
			Name:        b.Name,
			Description: b.Description,
			Icon:        b.Icon,
			Color:       b.Color,
			Action: models.ButtonAction{
				Controller: b.Action.Controller,
				Type:       b.Action.Type,
				Params:     b.Action.Params,
			},
		}
	}
	return result
}

func (h *HostController) ExecuteAction(action models.ButtonAction) error {
	body := map[string]interface{}{
		"controller": action.Controller,
		"type":       action.Type,
		"params":     action.Params,
	}
	return h.post("/execute", body, nil)
}

// ==================== internal helpers ====================

func (h *HostController) initialize(cfg map[string]interface{}) error {
	return h.post("/initialize", map[string]interface{}{"config": cfg}, nil)
}

func (h *HostController) fetchInfo() (pluginInfo, error) {
	var info pluginInfo
	err := h.get("/info", &info)
	return info, err
}

func (h *HostController) loadPersistedConfig() map[string]interface{} {
	var cfg map[string]interface{}
	// Config is stored relative to the storage data dir, under plugins/{id}/config.json
	filename := fmt.Sprintf("%s/config.json", h.id)
	h.storage.LoadJSON(filename, &cfg)
	if cfg == nil {
		cfg = make(map[string]interface{})
	}
	return cfg
}

func (h *HostController) persistConfig(cfg map[string]interface{}) error {
	filename := fmt.Sprintf("%s/config.json", h.id)
	return h.storage.SaveJSON(filename, cfg)
}

func (h *HostController) get(path string, out interface{}) error {
	return h.getWith(h.httpClient, path, out)
}

func (h *HostController) getWith(client *http.Client, path string, out interface{}) error {
	resp, err := client.Get(h.baseURL() + path)
	if err != nil {
		return fmt.Errorf("plugin %s GET %s: %w", h.id, path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("plugin %s GET %s: status %d: %s", h.id, path, resp.StatusCode, body)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (h *HostController) post(path string, body interface{}, out interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	resp, err := h.httpClient.Post(h.baseURL()+path, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("plugin %s POST %s: %w", h.id, path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		var errResp struct {
			Error string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return fmt.Errorf("%s", errResp.Error)
		}
		return fmt.Errorf("plugin %s POST %s: status %d", h.id, path, resp.StatusCode)
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
