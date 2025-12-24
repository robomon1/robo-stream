package main

import (
	"fmt"
	"log"
	"os"

	"github.com/andreykaipov/goobs"
)

func main() {
	// Get connection info from environment or use defaults
	host := getEnv("OBS_HOST", "localhost")
	port := getEnv("OBS_PORT", "4455")
	password := getEnv("OBS_PASSWORD", "")

	address := fmt.Sprintf("%s:%s", host, port)

	fmt.Printf("🔌 Connecting to OBS at %s...\n", address)

	// Connect to OBS
	var client *goobs.Client
	var err error

	if password != "" {
		client, err = goobs.New(
			address,
			goobs.WithPassword(password),
		)
	} else {
		client, err = goobs.New(address)
	}

	if err != nil {
		log.Fatalf("❌ Failed to connect: %v", err)
	}
	defer client.Disconnect()

	fmt.Println("✅ Connected successfully!")

	// Get version
	version, err := client.General.GetVersion()
	if err != nil {
		log.Fatalf("❌ Failed to get version: %v", err)
	}

	fmt.Printf("\n📊 OBS Information:\n")
	fmt.Printf("  OBS Version: %s\n", version.ObsVersion)
	fmt.Printf("  WebSocket Version: %s\n", version.ObsWebSocketVersion)
	fmt.Printf("  Available Requests: %d\n", len(version.AvailableRequests))

	// Get stats
	stats, err := client.General.GetStats()
	if err != nil {
		log.Printf("⚠️  Failed to get stats: %v", err)
	} else {
		fmt.Printf("\n📈 Current Stats:\n")
		fmt.Printf("  FPS: %.2f\n", stats.ActiveFps)
		fmt.Printf("  CPU Usage: %.2f%%\n", stats.CpuUsage)
		fmt.Printf("  Memory Usage: %.2f MB\n", stats.MemoryUsage)
		fmt.Printf("  Render Frames: %d\n", stats.RenderTotalFrames)
	}

	fmt.Println("\n✨ Connection test successful!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
