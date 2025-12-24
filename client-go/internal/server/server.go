package server

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/robomon1/stream-pi-go/client-go/internal/api"
	"github.com/robomon1/stream-pi-go/client-go/internal/client"
	"github.com/robomon1/stream-pi-go/client-go/internal/config"
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

// Start starts the web server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	s.logger.Infof("Starting Stream-Pi Client web server on http://localhost%s", addr)
	return http.ListenAndServe(addr, s.router)
}
