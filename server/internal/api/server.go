package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/robomon1/robo-stream/server/internal/manager"
	"github.com/robomon1/robo-stream/server/internal/models"
)

// Server provides HTTP API for clients
type Server struct {
	router         *mux.Router
	configManager  *manager.ConfigManager
	sessionManager *manager.SessionManager
	obsManager     *manager.OBSManager
}

// NewServer creates a new API server
func NewServer(
	cm *manager.ConfigManager,
	sm *manager.SessionManager,
	om *manager.OBSManager,
) *Server {
	s := &Server{
		router:         mux.NewRouter(),
		configManager:  cm,
		sessionManager: sm,
		obsManager:     om,
	}
	s.setupRoutes()
	return s
}

// setupRoutes configures API routes
func (s *Server) setupRoutes() {
	// Enable CORS
	s.router.Use(s.corsMiddleware)

	// Configuration endpoints
	s.router.HandleFunc("/api/configurations", s.listConfigurations).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/api/configurations/default", s.getDefaultConfiguration).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/api/configurations/{id}", s.getConfiguration).Methods("GET", "OPTIONS")

	// Client endpoints
	s.router.HandleFunc("/api/client/register", s.registerClient).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/api/client/config", s.getClientConfig).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/api/client/config/{id}", s.switchClientConfig).Methods("PUT", "OPTIONS")

	// Action endpoint
	s.router.HandleFunc("/api/action", s.executeAction).Methods("POST", "OPTIONS")

	// OBS status endpoints
	s.router.HandleFunc("/api/obs/status", s.getOBSStatus).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/api/obs/scenes", s.getScenes).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/api/obs/inputs", s.getInputs).Methods("GET", "OPTIONS")

	// Health check
	s.router.HandleFunc("/api/health", s.healthCheck).Methods("GET", "OPTIONS")
}

// corsMiddleware handles CORS
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

// Start starts the API server
func (s *Server) Start(addr string) error {
	log.Printf("API server listening on %s", addr)
	return http.ListenAndServe(addr, s.router)
}

// ==================== HANDLERS ====================

// healthCheck returns server health status
func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
		"obs_connected": s.obsManager.IsConnected(),
	})
}

// listConfigurations returns all configurations
func (s *Server) listConfigurations(w http.ResponseWriter, r *http.Request) {
	configs := s.configManager.List()
	s.respondJSON(w, http.StatusOK, configs)
}

// getDefaultConfiguration returns the default configuration with resolved buttons
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

// getConfiguration returns a specific configuration with resolved buttons
func (s *Server) getConfiguration(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	resolved, err := s.configManager.Resolve(id)
	if err != nil {
		s.respondError(w, http.StatusNotFound, err.Error())
		return
	}

	s.respondJSON(w, http.StatusOK, resolved)
}

// registerClient registers a new client or returns existing session
func (s *Server) registerClient(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ClientID   string `json:"client_id"`
		ClientName string `json:"client_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get client IP
	ipAddress := s.getClientIP(r)

	// Check if client already has a session
	existingSession, err := s.sessionManager.GetByClientID(req.ClientID)
	if err == nil {
		// Update existing session
		session, err := s.sessionManager.RegisterOrUpdate(
			req.ClientID,
			req.ClientName,
			existingSession.ConfigID,
			ipAddress,
		)
		if err != nil {
			s.respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Get resolved configuration
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

	// New client - assign default configuration
	defaultConfig, err := s.configManager.GetDefault()
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "no default configuration available")
		return
	}

	// Create new session
	session, err := s.sessionManager.RegisterOrUpdate(
		req.ClientID,
		req.ClientName,
		defaultConfig.ID,
		ipAddress,
	)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get resolved configuration
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

// getClientConfig returns the current configuration for a client
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

	// Update activity
	s.sessionManager.UpdateActivity(sessionID)

	// Get resolved configuration
	resolved, err := s.configManager.Resolve(session.ConfigID)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.respondJSON(w, http.StatusOK, resolved)
}

// switchClientConfig switches a client to a different configuration
func (s *Server) switchClientConfig(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		s.respondError(w, http.StatusBadRequest, "missing X-Session-ID header")
		return
	}

	vars := mux.Vars(r)
	configID := vars["id"]

	// Verify configuration exists
	_, err := s.configManager.Get(configID)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "configuration not found")
		return
	}

	// Update session
	if err := s.sessionManager.UpdateConfig(sessionID, configID); err != nil {
		s.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get resolved configuration
	resolved, err := s.configManager.Resolve(configID)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.respondJSON(w, http.StatusOK, resolved)
}

// executeAction executes an OBS action
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

	// Update activity
	s.sessionManager.UpdateActivity(sessionID)

	// Execute action
	if err := s.obsManager.ExecuteAction(action); err != nil {
		s.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

// getOBSStatus returns current OBS status
func (s *Server) getOBSStatus(w http.ResponseWriter, r *http.Request) {
	status, err := s.obsManager.GetStatus()
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.respondJSON(w, http.StatusOK, status)
}

// getScenes returns list of OBS scenes
func (s *Server) getScenes(w http.ResponseWriter, r *http.Request) {
	scenes, err := s.obsManager.GetScenes()
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.respondJSON(w, http.StatusOK, scenes)
}

// getInputs returns list of OBS inputs
func (s *Server) getInputs(w http.ResponseWriter, r *http.Request) {
	inputs, err := s.obsManager.GetInputs()
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.respondJSON(w, http.StatusOK, inputs)
}

// ==================== HELPERS ====================

// respondJSON writes a JSON response
func (s *Server) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError writes an error response
func (s *Server) respondError(w http.ResponseWriter, status int, message string) {
	s.respondJSON(w, status, map[string]interface{}{
		"error": message,
	})
}

// getClientIP extracts client IP from request
func (s *Server) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	parts := strings.Split(r.RemoteAddr, ":")
	if len(parts) > 0 {
		return parts[0]
	}

	return r.RemoteAddr
}
