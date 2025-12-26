package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robomon1/robo-stream/server-go/internal/api"
	"github.com/robomon1/robo-stream/server-go/internal/obs"
	"github.com/robomon1/robo-stream/server-go/internal/obs/actions"
	"github.com/sirupsen/logrus"
)

var (
	// Version information (can be set via ldflags at build time)
	Version   = "1.0.0"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Command-line flags
	obsHost := flag.String("obs-host", getEnvOrDefault("OBS_HOST", "localhost"), "OBS WebSocket host")
	obsPort := flag.Int("obs-port", getEnvOrDefaultInt("OBS_PORT", 4455), "OBS WebSocket port")
	obsPassword := flag.String("obs-password", os.Getenv("OBS_PASSWORD"), "OBS WebSocket password")
	httpPort := flag.Int("port", getEnvOrDefaultInt("SERVER_PORT", 8080), "HTTP API server port")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	showVersion := flag.Bool("version", false, "Show version information")
	testMode := flag.Bool("test", false, "Run in test mode (connect and display info)")

	flag.Parse()

	// Show version
	if *showVersion {
		fmt.Printf("Robo-Stream Server Go\n")
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		fmt.Printf("Git Commit: %s\n", GitCommit)
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

	logger.Info("🚀 Starting Robo-Stream Server Go")
	logger.Infof("Version: %s", Version)

	// Configure OBS connection
	obsConfig := &obs.Config{
		Host:              *obsHost,
		Port:              *obsPort,
		Password:          *obsPassword,
		AutoConnect:       true,
		ReconnectInterval: 5 * time.Second,
	}

	logger.Infof("Connecting to OBS at %s:%d", obsConfig.Host, obsConfig.Port)

	// Create OBS manager
	obsManager := obs.NewManager(obsConfig, logger)

	// Set up connection callback
	obsManager.OnConnect(func() {
		logger.Info("✅ Connected to OBS!")

		// Get OBS version
		version, err := obsManager.GetVersion()
		if err != nil {
			logger.Errorf("Failed to get OBS version: %v", err)
		} else {
			logger.Infof("OBS Version: %s", version)
		}

		// If in test mode, show some info and exit
		if *testMode {
			testOBSConnection(obsManager, logger)
		}
	})

	// Set up error callback
	obsManager.OnError(func(err error) {
		logger.Errorf("OBS error: %v", err)
	})

	// Connect to OBS
	if err := obsManager.Connect(); err != nil {
		logger.Fatalf("Failed to connect to OBS: %v", err)
	}

	// If test mode, wait a bit then exit
	if *testMode {
		time.Sleep(2 * time.Second)
		logger.Info("Test mode complete, exiting")
		return
	}

	// Start HTTP API server
	apiHandler := api.NewHandler(obsManager, logger)

	http.HandleFunc("/api/action", apiHandler.HandleAction)
	http.HandleFunc("/api/status", apiHandler.HandleGetStatus)
	http.HandleFunc("/api/scenes", apiHandler.HandleGetScenes)
	http.HandleFunc("/api/inputs", apiHandler.HandleGetInputs)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	go func() {
		addr := fmt.Sprintf("0.0.0.0:%d", *httpPort)
		listener, err := net.Listen("tcp4", addr)
		if err != nil {
			logger.Fatalf("Failed to create IPv4 listener: %v", err)
		}
		logger.Infof("🌐 HTTP API server starting on http://%s (IPv4)", listener.Addr().String())
		if err := http.Serve(listener, nil); err != nil {
			logger.Fatalf("HTTP server failed: %v", err)
		}
	}()

	logger.Info("Robo-Stream Server running. Press Ctrl+C to stop.")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down...")
	if err := obsManager.Disconnect(); err != nil {
		logger.Errorf("Error disconnecting from OBS: %v", err)
	}
	logger.Info("Goodbye!")
}

// testOBSConnection runs some basic tests to verify OBS connection
func testOBSConnection(manager *obs.Manager, logger *logrus.Logger) {
	if !manager.IsConnected() {
		logger.Error("Not connected to OBS")
		return
	}

	client := manager.Client()
	if client == nil {
		logger.Error("OBS client is nil")
		return
	}

	logger.Info("🧪 Running OBS integration tests...")

	// Test 1: Get scenes
	logger.Info("📋 Getting scene list...")
	scenes, err := actions.GetSceneList(client)
	if err != nil {
		logger.Errorf("Failed to get scenes: %v", err)
	} else {
		logger.Infof("Found %d scenes:", len(scenes))
		for i, scene := range scenes {
			if i < 5 {
				logger.Infof("  - %s", scene)
			}
		}
		if len(scenes) > 5 {
			logger.Infof("  ... and %d more", len(scenes)-5)
		}
	}

	// Test 2: Get current scene
	logger.Info("🎬 Getting current scene...")
	currentScene, err := actions.GetCurrentScene(client)
	if err != nil {
		logger.Errorf("Failed to get current scene: %v", err)
	} else {
		logger.Infof("Current scene: %s", currentScene)
	}

	// Test 3: Get stream status
	logger.Info("📡 Getting stream status...")
	streamStatus, err := actions.GetStreamStatus(client)
	if err != nil {
		logger.Errorf("Failed to get stream status: %v", err)
	} else {
		logger.Infof("Streaming: %v", streamStatus.Active)
	}

	// Test 4: Get recording status
	logger.Info("🔴 Getting recording status...")
	recordStatus, err := actions.GetRecordStatus(client)
	if err != nil {
		logger.Errorf("Failed to get record status: %v", err)
	} else {
		logger.Infof("Recording: %v (paused: %v)", recordStatus.Active, recordStatus.Paused)
	}

	// Test 5: Get inputs
	logger.Info("🎤 Getting input list...")
	inputs, err := actions.GetInputList(client)
	if err != nil {
		logger.Errorf("Failed to get inputs: %v", err)
	} else {
		logger.Infof("Found %d inputs:", len(inputs))
		for i, input := range inputs {
			if i < 5 {
				logger.Infof("  - %s", input)
			}
		}
		if len(inputs) > 5 {
			logger.Infof("  ... and %d more", len(inputs)-5)
		}
	}

	logger.Info("✅ All tests completed!")
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
