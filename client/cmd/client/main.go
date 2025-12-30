package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"client/internal/client"
	"client/internal/config"
	"client/internal/server"

	"github.com/sirupsen/logrus"
)

var (
	Version   = "1.0.0"
	BuildTime = "unknown"
)

func main() {
	// Command-line flags
	serverURL := flag.String("server-url", getEnvOrDefault("SERVER_URL", "http://localhost:8080"), "Robo-Stream server URL")
	port := flag.Int("port", getEnvOrDefaultInt("CLIENT_PORT", 3000), "Web client port")
	configFile := flag.String("config", "configs/buttons.json", "Button configuration file")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	showVersion := flag.Bool("version", false, "Show version information")

	flag.Parse()

	if *showVersion {
		fmt.Printf("Robo-Stream Client\n")
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		os.Exit(0)
	}

	// Setup logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	level, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	logger.Info("🎮 Starting Robo-Stream Client")
	logger.Infof("Version: %s", Version)

	// Load button configuration
	logger.Infof("Loading button configuration from: %s", *configFile)
	buttonConfig, err := config.LoadConfig(*configFile)
	if err != nil {
		logger.Warnf("Failed to load config (will use default): %v", err)
		buttonConfig = getDefaultConfig()

		// Save default config
		if err := config.SaveConfig(*configFile, buttonConfig); err != nil {
			logger.Errorf("Failed to save default config: %v", err)
		}
	}
	logger.Infof("Loaded %d buttons in %dx%d grid", len(buttonConfig.Buttons), buttonConfig.Grid.Rows, buttonConfig.Grid.Cols)

	// Create OBS client
	logger.Infof("Connecting to Robo-Stream server at: %s", *serverURL)
	obsClient := client.NewOBSClient(*serverURL, logger)

	// Create and start web server
	webServer := server.NewServer(obsClient, buttonConfig, logger, *port)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutting down...")
		os.Exit(0)
	}()

	logger.Infof("🌐 Web client available at: http://localhost:%d", *port)
	logger.Info("Press Ctrl+C to stop")

	if err := webServer.Start(); err != nil {
		logger.Fatalf("Failed to start web server: %v", err)
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvOrDefaultInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intVal int
		if _, err := fmt.Sscanf(value, "%d", &intVal); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getDefaultConfig() *config.ButtonConfig {
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
				Text:  "Scene 1",
				Color: "#3498db",
				Action: config.ButtonAction{
					Type: "switch_scene",
					Params: config.ActionParams{
						"scene_name": "Scene 1",
					},
				},
			},
			{
				ID:    "btn-0-1",
				Row:   0,
				Col:   1,
				Text:  "Stream",
				Color: "#e74c3c",
				Action: config.ButtonAction{
					Type: "toggle_stream",
				},
			},
			{
				ID:    "btn-0-2",
				Row:   0,
				Col:   2,
				Text:  "Record",
				Color: "#e67e22",
				Action: config.ButtonAction{
					Type: "toggle_record",
				},
			},
		},
	}
}
