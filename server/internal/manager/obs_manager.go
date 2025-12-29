package manager

import (
	"fmt"
	"log"
	"sync"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/inputs"
	"github.com/andreykaipov/goobs/api/requests/scenes"
	"github.com/robomon1/robo-stream/server/internal/models"
)

// OBSManager manages OBS WebSocket connection
type OBSManager struct {
	client *goobs.Client
	url    string
	mu     sync.RWMutex
}

// NewOBSManager creates a new OBSManager
func NewOBSManager() *OBSManager {
	return &OBSManager{}
}

// Connect connects to OBS WebSocket
func (om *OBSManager) Connect(url, password string) error {
	om.mu.Lock()
	defer om.mu.Unlock()

	if om.client != nil {
		om.client.Disconnect()
	}

	client, err := goobs.New(url, goobs.WithPassword(password))
	if err != nil {
		return fmt.Errorf("failed to connect to OBS: %w", err)
	}

	om.client = client
	om.url = url
	return nil
}

// Disconnect disconnects from OBS
func (om *OBSManager) Disconnect() error {
	om.mu.Lock()
	defer om.mu.Unlock()

	if om.client != nil {
		om.client.Disconnect()
		om.client = nil
	}
	return nil
}

// IsConnected returns whether connected to OBS
func (om *OBSManager) IsConnected() bool {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return om.client != nil
}

// GetURL returns the OBS WebSocket URL
func (om *OBSManager) GetURL() string {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return om.url
}

// GetScenes returns list of scene names
func (om *OBSManager) GetScenes() ([]string, error) {
	om.mu.RLock()
	client := om.client
	om.mu.RUnlock()

	if client == nil {
		return nil, fmt.Errorf("not connected to OBS")
	}

	resp, err := client.Scenes.GetSceneList()
	if err != nil {
		return nil, err
	}

	sceneNames := make([]string, len(resp.Scenes))
	for i, scene := range resp.Scenes {
		sceneNames[i] = scene.SceneName
	}

	log.Printf("🎬 GetScenes returning %d scenes: %v", len(sceneNames), sceneNames)
	return sceneNames, nil
}

// GetInputs returns list of input names
func (om *OBSManager) GetInputs() ([]string, error) {
	om.mu.RLock()
	client := om.client
	om.mu.RUnlock()

	if client == nil {
		return nil, fmt.Errorf("not connected to OBS")
	}

	resp, err := client.Inputs.GetInputList(&inputs.GetInputListParams{})
	if err != nil {
		return nil, err
	}

	inputNames := make([]string, len(resp.Inputs))
	for i, input := range resp.Inputs {
		inputNames[i] = input.InputName
	}

	log.Printf("🎤 GetInputs returning %d inputs: %v", len(inputNames), inputNames)
	return inputNames, nil
}

// ExecuteAction executes a button action
func (om *OBSManager) ExecuteAction(action models.ButtonAction) error {
	om.mu.RLock()
	client := om.client
	om.mu.RUnlock()

	if client == nil {
		return fmt.Errorf("not connected to OBS")
	}

	switch action.Type {
	case "switch_scene":
		sceneName, ok := action.Params["scene_name"].(string)
		if !ok {
			return fmt.Errorf("missing scene_name parameter")
		}
		_, err := client.Scenes.SetCurrentProgramScene(&scenes.SetCurrentProgramSceneParams{
			SceneName: &sceneName,
		})
		return err

	case "start_stream":
		_, err := client.Stream.StartStream()
		return err

	case "stop_stream":
		_, err := client.Stream.StopStream()
		return err

	case "toggle_stream":
		_, err := client.Stream.ToggleStream()
		return err

	case "start_record":
		_, err := client.Record.StartRecord()
		return err

	case "stop_record":
		_, err := client.Record.StopRecord()
		return err

	case "toggle_record":
		_, err := client.Record.ToggleRecord()
		return err

	case "pause_record":
		_, err := client.Record.PauseRecord()
		return err

	case "resume_record":
		_, err := client.Record.ResumeRecord()
		return err

	case "toggle_input_mute":
		inputName, ok := action.Params["input_name"].(string)
		if !ok {
			return fmt.Errorf("missing input_name parameter")
		}
		_, err := client.Inputs.ToggleInputMute(&inputs.ToggleInputMuteParams{
			InputName: &inputName,
		})
		return err

	case "mute_input":
		inputName, ok := action.Params["input_name"].(string)
		if !ok {
			return fmt.Errorf("missing input_name parameter")
		}
		muted := true
		_, err := client.Inputs.SetInputMute(&inputs.SetInputMuteParams{
			InputName:  &inputName,
			InputMuted: &muted,
		})
		return err

	case "unmute_input":
		inputName, ok := action.Params["input_name"].(string)
		if !ok {
			return fmt.Errorf("missing input_name parameter")
		}
		muted := false
		_, err := client.Inputs.SetInputMute(&inputs.SetInputMuteParams{
			InputName:  &inputName,
			InputMuted: &muted,
		})
		return err

	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
}

// GetStatus returns current OBS status
func (om *OBSManager) GetStatus() (map[string]interface{}, error) {
	om.mu.RLock()
	client := om.client
	om.mu.RUnlock()

	if client == nil {
		return map[string]interface{}{
			"connected": false,
		}, nil
	}

	// Get stream status
	streamResp, err := client.Stream.GetStreamStatus()
	if err != nil {
		return nil, err
	}

	// Get record status
	recordResp, err := client.Record.GetRecordStatus()
	if err != nil {
		return nil, err
	}

	// Get current scene
	sceneResp, err := client.Scenes.GetCurrentProgramScene()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"connected":     true,
		"streaming":     streamResp.OutputActive,
		"recording":     recordResp.OutputActive,
		"current_scene": sceneResp.CurrentProgramSceneName,
	}, nil
}
