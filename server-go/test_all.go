package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/andreykaipov/goobs"
)

func main() {
	fmt.Println("🧪 Stream-Pi OBS WebSocket Test Suite")
	fmt.Println("=====================================\n")

	// Get connection info
	host := getEnv("OBS_HOST", "localhost")
	port := getEnv("OBS_PORT", "4455")
	password := getEnv("OBS_PASSWORD", "")
	address := fmt.Sprintf("%s:%s", host, port)

	// Test 1: Connection
	fmt.Println("1️⃣  Testing Connection...")
	client, err := connect(address, password)
	if err != nil {
		log.Fatalf("❌ Connection failed: %v", err)
	}
	defer client.Disconnect()
	fmt.Println("   ✅ Connected!\n")

	// Test 2: Version
	fmt.Println("2️⃣  Testing Version...")
	version, err := client.General.GetVersion()
	if err != nil {
		log.Fatalf("❌ Version check failed: %v", err)
	}
	fmt.Printf("   ✅ OBS %s, WebSocket %s\n\n", version.ObsVersion, version.ObsWebSocketVersion)

	// Test 3: Scene List
	fmt.Println("3️⃣  Testing Scene List...")
	scenes, err := client.Scenes.GetSceneList()
	if err != nil {
		log.Fatalf("❌ Scene list failed: %v", err)
	}
	fmt.Printf("   ✅ Found %d scenes:\n", len(scenes.Scenes))
	for i, scene := range scenes.Scenes {
		if i < 5 { // Show first 5
			fmt.Printf("      - %s\n", scene.SceneName)
		}
	}
	if len(scenes.Scenes) > 5 {
		fmt.Printf("      ... and %d more\n", len(scenes.Scenes)-5)
	}
	fmt.Println()

	// Test 4: Current Scene
	fmt.Println("4️⃣  Testing Current Scene...")
	currentScene, err := client.Scenes.GetCurrentProgramScene()
	if err != nil {
		log.Fatalf("❌ Current scene failed: %v", err)
	}
	fmt.Printf("   ✅ Current: %s\n\n", currentScene.CurrentProgramSceneName)

	// Test 5: Stream Status
	fmt.Println("5️⃣  Testing Stream Status...")
	streamStatus, err := client.Stream.GetStreamStatus()
	if err != nil {
		log.Fatalf("❌ Stream status failed: %v", err)
	}
	fmt.Printf("   ✅ Streaming: %v\n", streamStatus.OutputActive)
	if streamStatus.OutputActive {
		fmt.Printf("      Duration: %d ms\n", streamStatus.OutputDuration)
		fmt.Printf("      Bytes: %d\n", streamStatus.OutputBytes)
	}
	fmt.Println()

	// Test 6: Recording Status
	fmt.Println("6️⃣  Testing Recording Status...")
	recStatus, err := client.Record.GetRecordStatus()
	if err != nil {
		log.Fatalf("❌ Record status failed: %v", err)
	}
	fmt.Printf("   ✅ Recording: %v\n", recStatus.OutputActive)
	if recStatus.OutputActive {
		fmt.Printf("      Duration: %d ms\n", recStatus.OutputDuration)
		fmt.Printf("      Paused: %v\n", recStatus.OutputPaused)
	}
	fmt.Println()

	// Test 7: Input List
	fmt.Println("7️⃣  Testing Input List...")
	inputs, err := client.Inputs.GetInputList(nil)
	if err != nil {
		log.Fatalf("❌ Input list failed: %v", err)
	}
	fmt.Printf("   ✅ Found %d inputs:\n", len(inputs.Inputs))
	for i, input := range inputs.Inputs {
		if i < 5 { // Show first 5
			fmt.Printf("      - %s (%s)\n", input.InputName, input.InputKind)
		}
	}
	if len(inputs.Inputs) > 5 {
		fmt.Printf("      ... and %d more\n", len(inputs.Inputs)-5)
	}
	fmt.Println()

	// Test 8: Event Handling
	fmt.Println("8️⃣  Testing Events (waiting 5 seconds)...")
	eventReceived := false
	client.AddEventHandler("CurrentProgramSceneChanged", func(event any) {
		eventReceived = true
		fmt.Println("   ✅ Received scene change event!")
	})

	fmt.Println("      Try switching scenes in OBS...")
	time.Sleep(5 * time.Second)
	if !eventReceived {
		fmt.Println("   ⚠️  No events received (switch scenes manually to test)")
	}
	fmt.Println()

	// Test 9: Stats
	fmt.Println("9️⃣  Testing Stats...")
	stats, err := client.General.GetStats()
	if err != nil {
		log.Fatalf("❌ Stats failed: %v", err)
	}
	fmt.Printf("   ✅ Performance:\n")
	fmt.Printf("      FPS: %.2f\n", stats.ActiveFps)
	fmt.Printf("      CPU: %.2f%%\n", stats.CpuUsage)
	fmt.Printf("      Memory: %.2f MB\n", stats.MemoryUsage)
	fmt.Printf("      Render Frames: %d\n", stats.RenderTotalFrames)
	fmt.Println()

	// Test 10: Replay Buffer Status
	fmt.Println("🔟 Testing Replay Buffer...")
	rbStatus, err := client.ReplayBuffer.GetReplayBufferStatus()
	if err != nil {
		fmt.Printf("   ⚠️  Replay buffer not available: %v\n", err)
	} else {
		fmt.Printf("   ✅ Replay Buffer: %v\n", rbStatus.OutputActive)
	}
	fmt.Println()

	// Summary
	fmt.Println("✅ All tests passed!")
	fmt.Println("\n📊 Test Summary:")
	fmt.Println("   • Connection: ✅")
	fmt.Println("   • Version Check: ✅")
	fmt.Println("   • Scene Management: ✅")
	fmt.Println("   • Streaming: ✅")
	fmt.Println("   • Recording: ✅")
	fmt.Println("   • Inputs: ✅")
	fmt.Println("   • Events: " + getStatus(eventReceived))
	fmt.Println("   • Stats: ✅")
	fmt.Println("\n🎉 OBS WebSocket integration is working correctly!")
}

func connect(address, password string) (*goobs.Client, error) {
	if password != "" {
		return goobs.New(address, goobs.WithPassword(password))
	}
	return goobs.New(address)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getStatus(success bool) string {
	if success {
		return "✅"
	}
	return "⚠️"
}
