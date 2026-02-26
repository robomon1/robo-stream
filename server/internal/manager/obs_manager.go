package manager

import (
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/filters"
	"github.com/andreykaipov/goobs/api/requests/inputs"
	"github.com/andreykaipov/goobs/api/requests/sceneitems"
	"github.com/andreykaipov/goobs/api/requests/scenes"
	"github.com/andreykaipov/goobs/api/requests/transitions"
	"github.com/andreykaipov/goobs/api/requests/ui"
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
	// ===== SCENES =====
	case "switch_scene":
		sceneName, ok := action.Params["scene_name"].(string)
		if !ok {
			return fmt.Errorf("missing scene_name parameter")
		}
		_, err := client.Scenes.SetCurrentProgramScene(&scenes.SetCurrentProgramSceneParams{
			SceneName: &sceneName,
		})
		return err

	// ===== STREAMING =====
	case "start_stream":
		_, err := client.Stream.StartStream()
		return err

	case "stop_stream":
		_, err := client.Stream.StopStream()
		return err

	case "toggle_stream":
		_, err := client.Stream.ToggleStream()
		return err

	// ===== RECORDING =====
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

	// ===== SOURCE VISIBILITY =====
	case "toggle_source_visibility":
		sceneName, ok := action.Params["scene_name"].(string)
		if !ok {
			return fmt.Errorf("missing scene_name parameter")
		}
		sourceName, ok := action.Params["source_name"].(string)
		if !ok {
			return fmt.Errorf("missing source_name parameter")
		}

		itemResp, err := client.SceneItems.GetSceneItemId(&sceneitems.GetSceneItemIdParams{
			SceneName:  &sceneName,
			SourceName: &sourceName,
		})
		if err != nil {
			return err
		}

		itemID := itemResp.SceneItemId

		stateResp, err := client.SceneItems.GetSceneItemEnabled(&sceneitems.GetSceneItemEnabledParams{
			SceneName:   &sceneName,
			SceneItemId: &itemID,
		})
		if err != nil {
			return err
		}

		newState := !stateResp.SceneItemEnabled

		_, err = client.SceneItems.SetSceneItemEnabled(&sceneitems.SetSceneItemEnabledParams{
			SceneName:        &sceneName,
			SceneItemId:      &itemID,
			SceneItemEnabled: &newState,
		})
		return err

	case "show_source":
		sceneName, ok := action.Params["scene_name"].(string)
		if !ok {
			return fmt.Errorf("missing scene_name parameter")
		}
		sourceName, ok := action.Params["source_name"].(string)
		if !ok {
			return fmt.Errorf("missing source_name parameter")
		}

		itemResp, err := client.SceneItems.GetSceneItemId(&sceneitems.GetSceneItemIdParams{
			SceneName:  &sceneName,
			SourceName: &sourceName,
		})
		if err != nil {
			return err
		}

		itemID := itemResp.SceneItemId
		enabled := true

		_, err = client.SceneItems.SetSceneItemEnabled(&sceneitems.SetSceneItemEnabledParams{
			SceneName:        &sceneName,
			SceneItemId:      &itemID,
			SceneItemEnabled: &enabled,
		})
		return err

	case "hide_source":
		sceneName, ok := action.Params["scene_name"].(string)
		if !ok {
			return fmt.Errorf("missing scene_name parameter")
		}
		sourceName, ok := action.Params["source_name"].(string)
		if !ok {
			return fmt.Errorf("missing source_name parameter")
		}

		itemResp, err := client.SceneItems.GetSceneItemId(&sceneitems.GetSceneItemIdParams{
			SceneName:  &sceneName,
			SourceName: &sourceName,
		})
		if err != nil {
			return err
		}

		itemID := itemResp.SceneItemId
		enabled := false

		_, err = client.SceneItems.SetSceneItemEnabled(&sceneitems.SetSceneItemEnabledParams{
			SceneName:        &sceneName,
			SceneItemId:      &itemID,
			SceneItemEnabled: &enabled,
		})
		return err

	// ===== AUDIO INPUTS =====
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

	case "set_input_volume":
		inputName, ok := action.Params["input_name"].(string)
		if !ok {
			return fmt.Errorf("missing input_name parameter")
		}

		// Volume can come as float64 or string
		var volumePercent float64
		switch v := action.Params["volume"].(type) {
		case float64:
			volumePercent = v
		case string:
			var err error
			volumePercent, err = strconv.ParseFloat(v, 64)
			if err != nil {
				return fmt.Errorf("invalid volume parameter: %s", v)
			}
		default:
			return fmt.Errorf("invalid volume parameter type")
		}

		// Convert percentage (0-100) to multiplier (0.0-1.0)
		volumeMultiplier := volumePercent / 100.0
		if volumeMultiplier < 0 {
			volumeMultiplier = 0
		}
		if volumeMultiplier > 1 {
			volumeMultiplier = 1
		}

		_, err := client.Inputs.SetInputVolume(&inputs.SetInputVolumeParams{
			InputName:      &inputName,
			InputVolumeMul: &volumeMultiplier,
		})
		return err

	// ===== VIRTUAL CAMERA =====
	case "start_virtual_cam":
		_, err := client.Outputs.StartVirtualCam()
		return err

	case "stop_virtual_cam":
		_, err := client.Outputs.StopVirtualCam()
		return err

	case "toggle_virtual_cam":
		_, err := client.Outputs.ToggleVirtualCam()
		return err

	// ===== REPLAY BUFFER =====
	case "start_replay_buffer":
		_, err := client.Outputs.StartReplayBuffer()
		return err

	case "stop_replay_buffer":
		_, err := client.Outputs.StopReplayBuffer()
		return err

	case "save_replay_buffer":
		_, err := client.Outputs.SaveReplayBuffer()
		return err

	case "toggle_replay_buffer":
		_, err := client.Outputs.ToggleReplayBuffer()
		return err

	// ===== FILTERS =====
	case "toggle_source_filter":
		sourceName, ok := action.Params["source_name"].(string)
		if !ok {
			return fmt.Errorf("missing source_name parameter")
		}
		filterName, ok := action.Params["filter_name"].(string)
		if !ok {
			return fmt.Errorf("missing filter_name parameter")
		}

		// Get current state
		stateResp, err := client.Filters.GetSourceFilter(&filters.GetSourceFilterParams{
			SourceName: &sourceName,
			FilterName: &filterName,
		})
		if err != nil {
			return err
		}

		// Toggle state
		newState := !stateResp.FilterEnabled
		_, err = client.Filters.SetSourceFilterEnabled(&filters.SetSourceFilterEnabledParams{
			SourceName:    &sourceName,
			FilterName:    &filterName,
			FilterEnabled: &newState,
		})
		return err

	case "enable_source_filter":
		sourceName, ok := action.Params["source_name"].(string)
		if !ok {
			return fmt.Errorf("missing source_name parameter")
		}
		filterName, ok := action.Params["filter_name"].(string)
		if !ok {
			return fmt.Errorf("missing filter_name parameter")
		}

		enabled := true
		_, err := client.Filters.SetSourceFilterEnabled(&filters.SetSourceFilterEnabledParams{
			SourceName:    &sourceName,
			FilterName:    &filterName,
			FilterEnabled: &enabled,
		})
		return err

	case "disable_source_filter":
		sourceName, ok := action.Params["source_name"].(string)
		if !ok {
			return fmt.Errorf("missing source_name parameter")
		}
		filterName, ok := action.Params["filter_name"].(string)
		if !ok {
			return fmt.Errorf("missing filter_name parameter")
		}

		enabled := false
		_, err := client.Filters.SetSourceFilterEnabled(&filters.SetSourceFilterEnabledParams{
			SourceName:    &sourceName,
			FilterName:    &filterName,
			FilterEnabled: &enabled,
		})
		return err

	// ===== MEDIA CONTROLS =====
	// NOTE: Media controls are not available in goobs v1.3.0
	// Upgrade to goobs v1.4+ to enable these features
	case "play_pause_media", "restart_media", "stop_media", "next_media", "previous_media":
		return fmt.Errorf("media controls require goobs v1.4+, currently using v1.3.0")

	// ===== TRANSITIONS =====
	case "trigger_transition":
		_, err := client.Transitions.TriggerStudioModeTransition()
		return err

	case "set_current_transition":
		transitionName, ok := action.Params["transition_name"].(string)
		if !ok {
			return fmt.Errorf("missing transition_name parameter")
		}
		_, err := client.Transitions.SetCurrentSceneTransition(&transitions.SetCurrentSceneTransitionParams{
			TransitionName: &transitionName,
		})
		return err

	case "set_transition_duration":
		// Duration can come as float64 or string
		var duration float64 // ← Changed from int to float64
		switch d := action.Params["duration"].(type) {
		case float64:
			duration = d // ← No conversion needed now
		case string:
			var err error
			duration, err = strconv.ParseFloat(d, 64) // ← Changed from Atoi to ParseFloat
			if err != nil {
				return fmt.Errorf("invalid duration parameter: %s", d)
			}
		default:
			return fmt.Errorf("invalid duration parameter type")
		}

		_, err := client.Transitions.SetCurrentSceneTransitionDuration(&transitions.SetCurrentSceneTransitionDurationParams{
			TransitionDuration: &duration,
		})
		return err

	// ===== STUDIO MODE =====
	case "toggle_studio_mode":
		// Get current state
		stateResp, err := client.Ui.GetStudioModeEnabled()
		if err != nil {
			return err
		}

		// Toggle state
		newState := !stateResp.StudioModeEnabled
		_, err = client.Ui.SetStudioModeEnabled(&ui.SetStudioModeEnabledParams{
			StudioModeEnabled: &newState,
		})
		return err

	case "enable_studio_mode":
		enabled := true
		_, err := client.Ui.SetStudioModeEnabled(&ui.SetStudioModeEnabledParams{
			StudioModeEnabled: &enabled,
		})
		return err

	case "disable_studio_mode":
		enabled := false
		_, err := client.Ui.SetStudioModeEnabled(&ui.SetStudioModeEnabledParams{
			StudioModeEnabled: &enabled,
		})
		return err

	case "set_preview_scene":
		sceneName, ok := action.Params["scene_name"].(string)
		if !ok {
			return fmt.Errorf("missing scene_name parameter")
		}
		_, err := client.Scenes.SetCurrentPreviewScene(&scenes.SetCurrentPreviewSceneParams{
			SceneName: &sceneName,
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

	// Get virtual camera status
	virtualCamResp, err := client.Outputs.GetVirtualCamStatus()
	virtualCamActive := false
	if err == nil {
		virtualCamActive = virtualCamResp.OutputActive
	}

	// Get replay buffer status
	replayBufferResp, err := client.Outputs.GetReplayBufferStatus()
	replayBufferActive := false
	if err == nil {
		replayBufferActive = replayBufferResp.OutputActive
	}

	// Get studio mode status
	studioModeResp, err := client.Ui.GetStudioModeEnabled()
	studioModeActive := false
	if err == nil {
		studioModeActive = studioModeResp.StudioModeEnabled
	}

	return map[string]interface{}{
		"connected":            true,
		"streaming":            streamResp.OutputActive,
		"recording":            recordResp.OutputActive,
		"recording_paused":     recordResp.OutputPaused,
		"current_scene":        sceneResp.CurrentProgramSceneName,
		"virtual_cam_active":   virtualCamActive,
		"replay_buffer_active": replayBufferActive,
		"studio_mode_active":   studioModeActive,
	}, nil
}

// GetInputMute returns whether an input is muted
func (om *OBSManager) GetInputMute(inputName string) (bool, error) {
	om.mu.RLock()
	client := om.client
	om.mu.RUnlock()

	if client == nil {
		return false, fmt.Errorf("not connected to OBS")
	}

	resp, err := client.Inputs.GetInputMute(&inputs.GetInputMuteParams{
		InputName: &inputName,
	})
	if err != nil {
		return false, err
	}

	return resp.InputMuted, nil
}

// GetSourceFilterEnabled returns whether a source filter is enabled
func (om *OBSManager) GetSourceFilterEnabled(sourceName, filterName string) (bool, error) {
	om.mu.RLock()
	client := om.client
	om.mu.RUnlock()

	if client == nil {
		return false, fmt.Errorf("not connected to OBS")
	}

	resp, err := client.Filters.GetSourceFilter(&filters.GetSourceFilterParams{
		SourceName: &sourceName,
		FilterName: &filterName,
	})
	if err != nil {
		return false, err
	}

	return resp.FilterEnabled, nil
}

// GetSourceVisibility checks if a source is visible in a scene
func (om *OBSManager) GetSourceVisibility(sceneName, sourceName string) (bool, error) {
	om.mu.RLock()
	client := om.client
	om.mu.RUnlock()

	if client == nil {
		return false, fmt.Errorf("not connected to OBS")
	}

	// Get the scene item ID
	itemResp, err := client.SceneItems.GetSceneItemId(&sceneitems.GetSceneItemIdParams{
		SceneName:  &sceneName,
		SourceName: &sourceName,
	})
	if err != nil {
		return false, err
	}

	itemID := itemResp.SceneItemId

	// Get visibility state
	stateResp, err := client.SceneItems.GetSceneItemEnabled(&sceneitems.GetSceneItemEnabledParams{
		SceneName:   &sceneName,
		SceneItemId: &itemID,
	})
	if err != nil {
		return false, err
	}

	return stateResp.SceneItemEnabled, nil
}
