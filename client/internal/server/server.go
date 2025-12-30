package server

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"client/internal/api"
	"client/internal/client"
	"client/internal/config"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

//go:embed ../../web/static/* ../../web/templates/*
var embeddedFiles embed.FS

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

	// Serve static files and templates
	staticFS, err := fs.Sub(embeddedFiles, "web/static")
	if err == nil {
		s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	}

	templatesFS, err := fs.Sub(embeddedFiles, "web/templates")
	if err == nil {
		s.router.PathPrefix("/").Handler(http.FileServer(http.FS(templatesFS)))
	}
}

// Start starts the web server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	s.logger.Infof("Starting Robo-Stream Client web server on http://localhost%s", addr)
	return http.ListenAndServe(addr, s.router)
}
