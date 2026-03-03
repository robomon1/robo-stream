package api

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/robomon1/robo-stream/server/internal/controller"
	obsctrl "github.com/robomon1/robo-stream/server/internal/controller/obs"
	"github.com/robomon1/robo-stream/server/internal/manager"
	"github.com/robomon1/robo-stream/server/internal/models"
	"github.com/robomon1/robo-stream/server/internal/plugin"
)

// Server provides the HTTP API consumed by remote clients and the Wails UI.
type Server struct {
	router         *mux.Router
	configManager  *manager.ConfigManager
	sessionManager *manager.SessionManager
	registry       *controller.Registry
	obsController  *obsctrl.Controller
	pluginManager  *plugin.Manager
}

// NewServer creates and configures the API server.
func NewServer(
	cm *manager.ConfigManager,
	sm *manager.SessionManager,
	reg *controller.Registry,
	obsCtrl *obsctrl.Controller,
	pm *plugin.Manager,
) *Server {
	s := &Server{
		router:         mux.NewRouter(),
		configManager:  cm,
		sessionManager: sm,
		registry:       reg,
		obsController:  obsCtrl,
		pluginManager:  pm,
	}
	s.setupRoutes()
	return s
}

// setupRoutes registers all HTTP routes.
func (s *Server) setupRoutes() {
	s.router.Use(s.corsMiddleware)

	// Configuration endpoints
	s.router.HandleFunc("/api/configurations", s.listConfigurations).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/api/configurations/default", s.getDefaultConfiguration).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/api/configurations/{id}", s.getConfiguration).Methods("GET", "OPTIONS")

	// Client session endpoints
	s.router.HandleFunc("/api/client/register", s.registerClient).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/api/client/config", s.getClientConfig).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/api/client/config/{id}", s.switchClientConfig).Methods("PUT", "OPTIONS")

	// Action execution
	s.router.HandleFunc("/api/action", s.executeAction).Methods("POST", "OPTIONS")

	// Session button indicators — plugin-agnostic indicator states for all
	// buttons in the current session's active configuration.
	s.router.HandleFunc("/api/session/button-indicators", s.getButtonIndicators).Methods("GET", "OPTIONS")

	// Generic controller endpoints
	s.router.HandleFunc("/api/controllers", s.listControllers).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/api/controllers/{id}/status", s.getControllerStatus).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/api/controllers/{id}/config/schema", s.getControllerConfigSchema).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/api/controllers/{id}/config", s.getControllerConfig).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/api/controllers/{id}/config", s.updateControllerConfig).Methods("PUT", "OPTIONS")
	s.router.HandleFunc("/api/controllers/{id}/actions", s.listControllerActions).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/api/controllers/{id}/restart", s.restartController).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/api/controllers/{id}", s.uninstallController).Methods("DELETE", "OPTIONS")

	// OBS-specific query endpoints (kept for client compatibility)
	s.router.HandleFunc("/api/obs/status", s.getOBSStatus).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/api/obs/scenes", s.getScenes).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/api/obs/inputs", s.getInputs).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/api/obs/source-visibility", s.getSourceVisibility).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/api/obs/input-mute", s.getInputMute).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/api/obs/source-filter", s.getSourceFilterEnabled).Methods("GET", "OPTIONS")

	s.router.HandleFunc("/api/health", s.healthCheck).Methods("GET", "OPTIONS")
}

// corsMiddleware adds CORS headers to every response.
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Session-ID, X-Client-ID")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Start begins listening on the given address.
func (s *Server) Start(addr string) error {
	log.Printf("API server listening on %s", addr)
	listener, err := net.Listen("tcp4", addr)
	if err != nil {
		log.Fatal(err)
	}
	return http.Serve(listener, s.router)
}

// ==================== HANDLERS ====================

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":        "ok",
		"obs_connected": s.obsController.IsConnected(),
	})
}

func (s *Server) listConfigurations(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, s.configManager.ListSummaries())
}

func (s *Server) getDefaultConfiguration(w http.ResponseWriter, r *http.Request) {
	config, err := s.configManager.GetDefault()
	if err != nil {
		s.respondError(w, http.StatusNotFound, err.Error())
		return
	}
	resolved, err := s.configManager.Resolve(config.ID)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.respondJSON(w, http.StatusOK, resolved)
}

func (s *Server) getConfiguration(w http.ResponseWriter, r *http.Request) {
	resolved, err := s.configManager.Resolve(mux.Vars(r)["id"])
	if err != nil {
		s.respondError(w, http.StatusNotFound, err.Error())
		return
	}
	s.respondJSON(w, http.StatusOK, resolved)
}

func (s *Server) registerClient(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ClientID   string `json:"client_id"`
		ClientName string `json:"client_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	ipAddress := s.getClientIP(r)

	existingSession, err := s.sessionManager.GetByClientID(req.ClientID)
	if err == nil {
		session, err := s.sessionManager.RegisterOrUpdate(req.ClientID, req.ClientName, existingSession.ConfigID, ipAddress)
		if err != nil {
			s.respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		resolved, err := s.configManager.Resolve(session.ConfigID)
		if err != nil {
			s.respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		s.respondJSON(w, http.StatusOK, map[string]interface{}{
			"session_id": session.SessionID,
			"config_id":  session.ConfigID,
			"config":     resolved,
		})
		return
	}

	defaultConfig, err := s.configManager.GetDefault()
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "no default configuration available")
		return
	}
	session, err := s.sessionManager.RegisterOrUpdate(req.ClientID, req.ClientName, defaultConfig.ID, ipAddress)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resolved, err := s.configManager.Resolve(defaultConfig.ID)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"session_id": session.SessionID,
		"config_id":  defaultConfig.ID,
		"config":     resolved,
	})
}

func (s *Server) getClientConfig(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		s.respondError(w, http.StatusBadRequest, "missing X-Session-ID header")
		return
	}
	session, err := s.sessionManager.Get(sessionID)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "session not found")
		return
	}
	s.sessionManager.UpdateActivity(sessionID)
	resolved, err := s.configManager.Resolve(session.ConfigID)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.respondJSON(w, http.StatusOK, resolved)
}

func (s *Server) switchClientConfig(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		s.respondError(w, http.StatusBadRequest, "missing X-Session-ID header")
		return
	}
	configID := mux.Vars(r)["id"]
	if _, err := s.configManager.Get(configID); err != nil {
		s.respondError(w, http.StatusNotFound, "configuration not found")
		return
	}
	if err := s.sessionManager.UpdateConfig(sessionID, configID); err != nil {
		s.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resolved, err := s.configManager.Resolve(configID)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.respondJSON(w, http.StatusOK, resolved)
}

func (s *Server) executeAction(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		s.respondError(w, http.StatusBadRequest, "missing X-Session-ID header")
		return
	}
	var action models.ButtonAction
	if err := json.NewDecoder(r.Body).Decode(&action); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	s.sessionManager.UpdateActivity(sessionID)
	if err := s.registry.ExecuteAction(action); err != nil {
		s.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

// getButtonIndicators returns a map of buttonID → indicator class ("active", "warn",
// or "") for every button in the calling session's active configuration.
// This lets clients display indicators without any plugin-specific code: they
// simply apply the returned class to the matching button element.
//
// Algorithm:
//  1. Look up the session and resolve its active configuration.
//  2. Pre-warm the status cache for every distinct controller used in the config
//     by calling GetStatus() once per controller.  This ensures ComputeIndicator
//     reads a fresh status snapshot rather than stale (or nil) cache data.
//  3. Call ComputeIndicator(action) for each button and collect the results.
func (s *Server) getButtonIndicators(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		s.respondError(w, http.StatusBadRequest, "missing X-Session-ID header")
		return
	}

	session, err := s.sessionManager.Get(sessionID)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "session not found")
		return
	}
	s.sessionManager.UpdateActivity(sessionID)

	resolved, err := s.configManager.Resolve(session.ConfigID)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Collect all distinct controller IDs used by buttons in this config.
	seen := make(map[string]bool)
	for _, btn := range resolved.Buttons {
		ctrlID := btn.Action.Controller
		if ctrlID == "" {
			ctrlID = "obs" // default for legacy buttons
		}
		seen[ctrlID] = true
	}

	// Pre-warm each controller's internal status cache with a single GetStatus call.
	// ComputeIndicator then reads from the cache — no extra round-trips per button.
	for ctrlID := range seen {
		if ctrl, ok := s.registry.Get(ctrlID); ok {
			ctrl.GetStatus()
		}
	}

	// Compute indicator for every button.
	indicators := make(map[string]string, len(resolved.Buttons))
	for _, btn := range resolved.Buttons {
		ctrlID := btn.Action.Controller
		if ctrlID == "" {
			ctrlID = "obs"
		}
		if ctrl, ok := s.registry.Get(ctrlID); ok {
			indicators[btn.ID] = ctrl.ComputeIndicator(btn.Action)
		} else {
			indicators[btn.ID] = ""
		}
	}

	s.respondJSON(w, http.StatusOK, indicators)
}

// ==================== CONTROLLER ENDPOINTS ====================

func (s *Server) listControllers(w http.ResponseWriter, r *http.Request) {
	controllers := s.registry.List()
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
	s.respondJSON(w, http.StatusOK, result)
}

func (s *Server) getControllerStatus(w http.ResponseWriter, r *http.Request) {
	c, ok := s.registry.Get(mux.Vars(r)["id"])
	if !ok {
		s.respondError(w, http.StatusNotFound, "controller not found")
		return
	}
	s.respondJSON(w, http.StatusOK, c.GetStatus())
}

func (s *Server) getControllerConfigSchema(w http.ResponseWriter, r *http.Request) {
	c, ok := s.registry.Get(mux.Vars(r)["id"])
	if !ok {
		s.respondError(w, http.StatusNotFound, "controller not found")
		return
	}
	s.respondJSON(w, http.StatusOK, c.GetConfigSchema())
}

func (s *Server) getControllerConfig(w http.ResponseWriter, r *http.Request) {
	c, ok := s.registry.Get(mux.Vars(r)["id"])
	if !ok {
		s.respondError(w, http.StatusNotFound, "controller not found")
		return
	}
	cfg := c.GetCurrentConfig()
	// Redact password fields
	for k := range cfg {
		if strings.Contains(strings.ToLower(k), "password") || strings.Contains(strings.ToLower(k), "secret") {
			cfg[k] = "••••••••"
		}
	}
	s.respondJSON(w, http.StatusOK, cfg)
}

func (s *Server) updateControllerConfig(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	c, ok := s.registry.Get(id)
	if !ok {
		s.respondError(w, http.StatusNotFound, "controller not found")
		return
	}
	var cfg map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := c.SaveConfig(cfg); err != nil {
		s.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (s *Server) listControllerActions(w http.ResponseWriter, r *http.Request) {
	c, ok := s.registry.Get(mux.Vars(r)["id"])
	if !ok {
		s.respondError(w, http.StatusNotFound, "controller not found")
		return
	}
	s.respondJSON(w, http.StatusOK, c.SupportedActionTypes())
}

func (s *Server) restartController(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "obs" {
		s.respondError(w, http.StatusBadRequest, "built-in controllers cannot be restarted via API")
		return
	}
	if err := s.pluginManager.Restart(id); err != nil {
		s.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (s *Server) uninstallController(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "obs" {
		s.respondError(w, http.StatusBadRequest, "built-in controllers cannot be uninstalled")
		return
	}
	if err := s.pluginManager.Uninstall(id); err != nil {
		s.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.respondJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

// ==================== OBS-SPECIFIC ENDPOINTS ====================

func (s *Server) getOBSStatus(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, s.obsController.GetStatus())
}

func (s *Server) getScenes(w http.ResponseWriter, r *http.Request) {
	scenes, err := s.obsController.GetScenes()
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.respondJSON(w, http.StatusOK, scenes)
}

func (s *Server) getInputs(w http.ResponseWriter, r *http.Request) {
	inputs, err := s.obsController.GetInputs()
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.respondJSON(w, http.StatusOK, inputs)
}

func (s *Server) getSourceVisibility(w http.ResponseWriter, r *http.Request) {
	sceneName := r.URL.Query().Get("scene")
	sourceName := r.URL.Query().Get("source")
	if sceneName == "" || sourceName == "" {
		s.respondError(w, http.StatusBadRequest, "missing scene or source parameter")
		return
	}
	visible, err := s.obsController.GetSourceVisibility(sceneName, sourceName)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.respondJSON(w, http.StatusOK, map[string]interface{}{"visible": visible})
}

func (s *Server) getInputMute(w http.ResponseWriter, r *http.Request) {
	inputName := r.URL.Query().Get("input")
	if inputName == "" {
		s.respondError(w, http.StatusBadRequest, "missing input parameter")
		return
	}
	muted, err := s.obsController.GetInputMute(inputName)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.respondJSON(w, http.StatusOK, map[string]interface{}{"muted": muted})
}

func (s *Server) getSourceFilterEnabled(w http.ResponseWriter, r *http.Request) {
	sourceName := r.URL.Query().Get("source")
	filterName := r.URL.Query().Get("filter")
	if sourceName == "" || filterName == "" {
		s.respondError(w, http.StatusBadRequest, "missing source or filter parameter")
		return
	}
	enabled, err := s.obsController.GetSourceFilterEnabled(sourceName, filterName)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.respondJSON(w, http.StatusOK, map[string]interface{}{"enabled": enabled})
}

// ==================== HELPERS ====================

func (s *Server) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) respondError(w http.ResponseWriter, status int, message string) {
	s.respondJSON(w, status, map[string]interface{}{"error": message})
}

func (s *Server) getClientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return strings.TrimSpace(strings.Split(fwd, ",")[0])
	}
	if real := r.Header.Get("X-Real-IP"); real != "" {
		return real
	}
	parts := strings.Split(r.RemoteAddr, ":")
	if len(parts) > 0 {
		return parts[0]
	}
	return r.RemoteAddr
}
