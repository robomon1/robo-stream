package server

import (
	"fmt"
	"net"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/robomon1/robo-stream/client-go/internal/api"
	"github.com/robomon1/robo-stream/client-go/internal/client"
	"github.com/robomon1/robo-stream/client-go/internal/config"
	"github.com/sirupsen/logrus"
)

// Server represents the web server
type Server struct {
	router    *mux.Router
	obsClient *client.OBSClient
	config    *config.ButtonConfig
	logger    *logrus.Logger
	port      int
}

// NewServer creates a new web server
func NewServer(obsClient *client.OBSClient, buttonConfig *config.ButtonConfig, logger *logrus.Logger, port int) *Server {
	s := &Server{
		router:    mux.NewRouter(),
		obsClient: obsClient,
		config:    buttonConfig,
		logger:    logger,
		port:      port,
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// API routes
	apiHandler := api.NewHandler(s.obsClient, s.config, s.logger)
	
	s.router.HandleFunc("/api/buttons", apiHandler.GetButtons).Methods("GET")
	s.router.HandleFunc("/api/buttons", apiHandler.UpdateButtons).Methods("POST")
	s.router.HandleFunc("/api/buttons/{id}", apiHandler.UpdateButton).Methods("PUT")
	s.router.HandleFunc("/api/buttons/{id}", apiHandler.DeleteButton).Methods("DELETE")
	s.router.HandleFunc("/api/buttons/{id}/press", apiHandler.PressButton).Methods("POST")
	s.router.HandleFunc("/api/scenes", apiHandler.GetScenes).Methods("GET")
	s.router.HandleFunc("/api/inputs", apiHandler.GetInputs).Methods("GET")
	s.router.HandleFunc("/api/status", apiHandler.GetStatus).Methods("GET")
	s.router.HandleFunc("/ws", apiHandler.HandleWebSocket)

	// Serve static files from disk
	staticDir := filepath.Join("web", "static")
	s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// Serve templates (index.html) from disk
	templatesDir := filepath.Join("web", "templates")
	s.router.PathPrefix("/").Handler(http.FileServer(http.Dir(templatesDir)))
}

// Start starts the web server - FORCED TO IPv4
func (s *Server) Start() error {
	addr := fmt.Sprintf("0.0.0.0:%d", s.port)
	s.logger.Infof("Starting Robo-Stream Client web server on http://%s", addr)
	
	// Create TCP4 listener to force IPv4
	listener, err := net.Listen("tcp4", addr)
	if err != nil {
		return fmt.Errorf("failed to create IPv4 listener: %w", err)
	}
	
	s.logger.Infof("✅ Listening on IPv4: %s", listener.Addr().String())
	s.logger.Infof("Access from this computer: http://localhost:%d", s.port)
	
	return http.Serve(listener, s.router)
}
