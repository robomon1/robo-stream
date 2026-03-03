// Package obs provides the built-in OBS Studio controller.
// It is always compiled into the server binary and registered at startup.
package obs

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/filters"
	"github.com/andreykaipov/goobs/api/requests/inputs"
	"github.com/andreykaipov/goobs/api/requests/sceneitems"
	"github.com/andreykaipov/goobs/api/requests/scenes"
	"github.com/andreykaipov/goobs/api/requests/transitions"
	"github.com/andreykaipov/goobs/api/requests/ui"
	"github.com/robomon1/robo-stream/server/internal/controller"
	"github.com/robomon1/robo-stream/server/internal/models"
	"github.com/robomon1/robo-stream/server/internal/storage"
)

const (
	obsHealthCheckInterval = 15 * time.Second
	obsReconnectDelay      = 10 * time.Second
	obsMaxReconnects       = 3
)

// Controller is the built-in OBS Studio controller.
// In addition to the generic controller.Controller interface it exposes
// OBS-specific query methods (GetScenes, GetInputs, etc.) that the server
// uses for its OBS settings UI.
type Controller struct {
	client  *goobs.Client
	url     string
	storage *storage.Storage
	mu      sync.RWMutex

	// Reconnection state — all guarded by mu.
	password          string
	reconnectAttempts int
	reconnectGaveUp   bool
	cancelMonitor     context.CancelFunc

	// lastStatus caches the most recent result of GetStatus() so that
	// ComputeIndicator can compute simple boolean indicators without
	// additional OBS round-trips.
	lastStatusMu sync.RWMutex
	lastStatus   map[string]interface{}
}

// New creates an OBS controller backed by the provided storage.
func New(st *storage.Storage) *Controller {
	return &Controller{storage: st}
}

// ==================== controller.Controller interface ====================

func (c *Controller) ID() string          { return "obs" }
func (c *Controller) Name() string        { return "OBS Studio" }
func (c *Controller) Description() string { return "Controls OBS Studio via the OBS WebSocket v5 protocol" }
func (c *Controller) Version() string     { return "1.0.0" }

func (c *Controller) Shutdown() error {
	return c.Disconnect()
}

func (c *Controller) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.client != nil
}

func (c *Controller) GetStatus() map[string]interface{} {
	c.mu.RLock()
	client := c.client
	gaveUp := c.reconnectGaveUp
	attempts := c.reconnectAttempts
	c.mu.RUnlock()

	if client == nil {
		if gaveUp {
			return map[string]interface{}{
				"connected": false,
				"error":     "Connection to OBS lost. Reconnect attempts exhausted — click Save & Connect to try again.",
			}
		}
		if attempts > 0 {
			return map[string]interface{}{
				"connected": false,
				"error":     fmt.Sprintf("OBS disconnected — reconnecting… (attempt %d/%d)", attempts, obsMaxReconnects),
			}
		}
		return map[string]interface{}{"connected": false}
	}

	streamResp, err := client.Stream.GetStreamStatus()
	if err != nil {
		return map[string]interface{}{"connected": false}
	}
	recordResp, err := client.Record.GetRecordStatus()
	if err != nil {
		return map[string]interface{}{"connected": false}
	}
	sceneResp, err := client.Scenes.GetCurrentProgramScene()
	if err != nil {
		return map[string]interface{}{"connected": false}
	}

	virtualCamActive := false
	if vcResp, err := client.Outputs.GetVirtualCamStatus(); err == nil {
		virtualCamActive = vcResp.OutputActive
	}
	replayBufferActive := false
	if rbResp, err := client.Outputs.GetReplayBufferStatus(); err == nil {
		replayBufferActive = rbResp.OutputActive
	}
	studioModeActive := false
	if smResp, err := client.Ui.GetStudioModeEnabled(); err == nil {
		studioModeActive = smResp.StudioModeEnabled
	}

	result := map[string]interface{}{
		"connected":            true,
		"streaming":            streamResp.OutputActive,
		"recording":            recordResp.OutputActive,
		"recording_paused":     recordResp.OutputPaused,
		"current_scene":        sceneResp.CurrentProgramSceneName,
		"virtual_cam_active":   virtualCamActive,
		"replay_buffer_active": replayBufferActive,
		"studio_mode_active":   studioModeActive,
	}

	c.lastStatusMu.Lock()
	c.lastStatus = result
	c.lastStatusMu.Unlock()

	return result
}

// ComputeIndicator returns the CSS indicator class for the given button action.
// For simple boolean status fields it reads from the cached lastStatus (set by
// the most recent GetStatus call) to avoid extra OBS round-trips. For
// parameterised actions (source visibility, input mute, filters) it queries OBS
// directly, since those states are button-specific and cannot be pre-fetched as
// a single status blob.
func (c *Controller) ComputeIndicator(action models.ButtonAction) string {
	c.lastStatusMu.RLock()
	status := c.lastStatus
	c.lastStatusMu.RUnlock()

	if status == nil {
		return ""
	}
	connected, _ := status["connected"].(bool)
	if !connected {
		return ""
	}

	streaming, _ := status["streaming"].(bool)
	recording, _ := status["recording"].(bool)
	recordingPaused, _ := status["recording_paused"].(bool)
	currentScene, _ := status["current_scene"].(string)
	virtualCamActive, _ := status["virtual_cam_active"].(bool)
	replayBufferActive, _ := status["replay_buffer_active"].(bool)
	studioModeActive, _ := status["studio_mode_active"].(bool)

	switch action.Type {
	// ── Streaming ─────────────────────────────────────────────────────────
	case "start_stream", "toggle_stream":
		if streaming {
			return "active"
		}
	case "stop_stream":
		if !streaming {
			return "active"
		}

	// ── Recording ─────────────────────────────────────────────────────────
	case "start_record", "toggle_record":
		if recording {
			return "active"
		}
	case "stop_record":
		if !recording {
			return "active"
		}
	case "pause_record":
		if recording && !recordingPaused {
			return "active"
		}
	case "resume_record":
		if recording && recordingPaused {
			return "active"
		}

	// ── Virtual Camera ────────────────────────────────────────────────────
	case "start_virtual_cam", "toggle_virtual_cam":
		if virtualCamActive {
			return "active"
		}
	case "stop_virtual_cam":
		if !virtualCamActive {
			return "active"
		}

	// ── Replay Buffer ─────────────────────────────────────────────────────
	case "start_replay_buffer", "toggle_replay_buffer", "save_replay_buffer":
		if replayBufferActive {
			return "active"
		}
	case "stop_replay_buffer":
		if !replayBufferActive {
			return "active"
		}

	// ── Studio Mode ───────────────────────────────────────────────────────
	case "toggle_studio_mode", "enable_studio_mode", "trigger_transition":
		if studioModeActive {
			return "active"
		}
	case "disable_studio_mode":
		if !studioModeActive {
			return "active"
		}

	// ── Scene switching ───────────────────────────────────────────────────
	case "switch_scene":
		sceneName, _ := action.Params["scene_name"].(string)
		if sceneName != "" && currentScene == sceneName {
			return "active"
		}

	// ── Source visibility ─────────────────────────────────────────────────
	case "toggle_source_visibility", "show_source":
		sceneName, _ := action.Params["scene_name"].(string)
		sourceName, _ := action.Params["source_name"].(string)
		if sceneName != "" && sourceName != "" {
			if visible, err := c.GetSourceVisibility(sceneName, sourceName); err == nil && visible {
				return "active"
			}
		}
	case "hide_source":
		sceneName, _ := action.Params["scene_name"].(string)
		sourceName, _ := action.Params["source_name"].(string)
		if sceneName != "" && sourceName != "" {
			if visible, err := c.GetSourceVisibility(sceneName, sourceName); err == nil && !visible {
				return "active"
			}
		}

	// ── Input mute ────────────────────────────────────────────────────────
	case "toggle_input_mute", "mute_input":
		inputName, _ := action.Params["input_name"].(string)
		if inputName != "" {
			if muted, err := c.GetInputMute(inputName); err == nil && muted {
				return "active"
			}
		}
	case "unmute_input":
		inputName, _ := action.Params["input_name"].(string)
		if inputName != "" {
			if muted, err := c.GetInputMute(inputName); err == nil && !muted {
				return "active"
			}
		}

	// ── Source filters ────────────────────────────────────────────────────
	case "toggle_source_filter", "enable_source_filter":
		sourceName, _ := action.Params["source_name"].(string)
		filterName, _ := action.Params["filter_name"].(string)
		if sourceName != "" && filterName != "" {
			if enabled, err := c.GetSourceFilterEnabled(sourceName, filterName); err == nil && enabled {
				return "active"
			}
		}
	case "disable_source_filter":
		sourceName, _ := action.Params["source_name"].(string)
		filterName, _ := action.Params["filter_name"].(string)
		if sourceName != "" && filterName != "" {
			if enabled, err := c.GetSourceFilterEnabled(sourceName, filterName); err == nil && !enabled {
				return "active"
			}
		}
	}

	return ""
}

func (c *Controller) GetConfigSchema() []controller.ConfigField {
	return []controller.ConfigField{
		{Key: "url", Label: "WebSocket URL", Type: "string", Required: true, Default: "localhost:4455", Help: "OBS WebSocket address (host:port)"},
		{Key: "password", Label: "Password", Type: "password", Help: "OBS WebSocket password (leave blank if disabled)"},
	}
}

func (c *Controller) GetCurrentConfig() map[string]interface{} {
	// Environment variables take precedence
	if envURL := os.Getenv("OBS_WEBSOCKET_URL"); envURL != "" {
		return map[string]interface{}{
			"url":      envURL,
			"password": os.Getenv("OBS_WEBSOCKET_PASSWORD"),
		}
	}

	var cfg models.OBSConfig
	if err := c.storage.LoadJSON("obs_config.json", &cfg); err != nil || cfg.URL == "" {
		return map[string]interface{}{"url": "localhost:4455", "password": ""}
	}
	return map[string]interface{}{"url": cfg.URL, "password": cfg.Password}
}

func (c *Controller) SaveConfig(cfg map[string]interface{}) error {
	url, _ := cfg["url"].(string)
	password, _ := cfg["password"].(string)
	return c.storage.SaveJSON("obs_config.json", models.OBSConfig{URL: url, Password: password})
}

func (c *Controller) GetDefaultButtons() []*models.Button {
	return []*models.Button{
		{Name: "Go Live", Description: "Start streaming", Icon: "video", Color: "#e74c3c",
			Action: models.ButtonAction{Controller: "obs", Type: "start_stream"}},
		{Name: "Stop Stream", Description: "Stop streaming", Icon: "stop-circle", Color: "#95a5a6",
			Action: models.ButtonAction{Controller: "obs", Type: "stop_stream"}},
		{Name: "Start Record", Description: "Start recording", Icon: "circle", Color: "#e74c3c",
			Action: models.ButtonAction{Controller: "obs", Type: "start_record"}},
		{Name: "Stop Record", Description: "Stop recording", Icon: "stop-circle", Color: "#95a5a6",
			Action: models.ButtonAction{Controller: "obs", Type: "stop_record"}},
		{Name: "Mute Mic", Description: "Mute microphone", Icon: "mic-off", Color: "#e67e22",
			Action: models.ButtonAction{Controller: "obs", Type: "toggle_input_mute",
				Params: map[string]interface{}{"input_name": "Mic/Aux"}}},
		{Name: "Scene", Description: "Switch to main scene", Icon: "layout", Color: "#3498db",
			Action: models.ButtonAction{Controller: "obs", Type: "switch_scene",
				Params: map[string]interface{}{"scene_name": "Scene"}}},
	}
}

func (c *Controller) SupportedActionTypes() []controller.ActionTypeDefinition {
	return []controller.ActionTypeDefinition{
		{Type: "switch_scene", Name: "Switch Scene", Params: []controller.ParamDef{
			{Key: "scene_name", Label: "Scene Name", Type: "string", Required: true},
		}},
		{Type: "set_preview_scene", Name: "Set Preview Scene (Studio Mode)", Params: []controller.ParamDef{
			{Key: "scene_name", Label: "Scene Name", Type: "string", Required: true},
		}},
		{Type: "start_stream", Name: "Start Stream"},
		{Type: "stop_stream", Name: "Stop Stream"},
		{Type: "toggle_stream", Name: "Toggle Stream"},
		{Type: "start_record", Name: "Start Recording"},
		{Type: "stop_record", Name: "Stop Recording"},
		{Type: "toggle_record", Name: "Toggle Recording"},
		{Type: "pause_record", Name: "Pause Recording"},
		{Type: "resume_record", Name: "Resume Recording"},
		{Type: "toggle_source_visibility", Name: "Toggle Source Visibility", Params: []controller.ParamDef{
			{Key: "scene_name", Label: "Scene Name", Type: "string", Required: true},
			{Key: "source_name", Label: "Source Name", Type: "string", Required: true},
		}},
		{Type: "show_source", Name: "Show Source", Params: []controller.ParamDef{
			{Key: "scene_name", Label: "Scene Name", Type: "string", Required: true},
			{Key: "source_name", Label: "Source Name", Type: "string", Required: true},
		}},
		{Type: "hide_source", Name: "Hide Source", Params: []controller.ParamDef{
			{Key: "scene_name", Label: "Scene Name", Type: "string", Required: true},
			{Key: "source_name", Label: "Source Name", Type: "string", Required: true},
		}},
		{Type: "toggle_input_mute", Name: "Toggle Input Mute", Params: []controller.ParamDef{
			{Key: "input_name", Label: "Input Name", Type: "string", Required: true},
		}},
		{Type: "mute_input", Name: "Mute Input", Params: []controller.ParamDef{
			{Key: "input_name", Label: "Input Name", Type: "string", Required: true},
		}},
		{Type: "unmute_input", Name: "Unmute Input", Params: []controller.ParamDef{
			{Key: "input_name", Label: "Input Name", Type: "string", Required: true},
		}},
		{Type: "set_input_volume", Name: "Set Input Volume", Params: []controller.ParamDef{
			{Key: "input_name", Label: "Input Name", Type: "string", Required: true},
			{Key: "volume", Label: "Volume (0-100)", Type: "number", Required: true, Default: float64(100)},
		}},
		{Type: "start_virtual_cam", Name: "Start Virtual Camera"},
		{Type: "stop_virtual_cam", Name: "Stop Virtual Camera"},
		{Type: "toggle_virtual_cam", Name: "Toggle Virtual Camera"},
		{Type: "start_replay_buffer", Name: "Start Replay Buffer"},
		{Type: "stop_replay_buffer", Name: "Stop Replay Buffer"},
		{Type: "save_replay_buffer", Name: "Save Replay Buffer"},
		{Type: "toggle_replay_buffer", Name: "Toggle Replay Buffer"},
		{Type: "toggle_source_filter", Name: "Toggle Source Filter", Params: []controller.ParamDef{
			{Key: "source_name", Label: "Source Name", Type: "string", Required: true},
			{Key: "filter_name", Label: "Filter Name", Type: "string", Required: true},
		}},
		{Type: "enable_source_filter", Name: "Enable Source Filter", Params: []controller.ParamDef{
			{Key: "source_name", Label: "Source Name", Type: "string", Required: true},
			{Key: "filter_name", Label: "Filter Name", Type: "string", Required: true},
		}},
		{Type: "disable_source_filter", Name: "Disable Source Filter", Params: []controller.ParamDef{
			{Key: "source_name", Label: "Source Name", Type: "string", Required: true},
			{Key: "filter_name", Label: "Filter Name", Type: "string", Required: true},
		}},
		{Type: "trigger_transition", Name: "Trigger Studio Mode Transition"},
		{Type: "set_current_transition", Name: "Set Transition", Params: []controller.ParamDef{
			{Key: "transition_name", Label: "Transition Name", Type: "string", Required: true},
		}},
		{Type: "set_transition_duration", Name: "Set Transition Duration", Params: []controller.ParamDef{
			{Key: "duration", Label: "Duration (ms)", Type: "number", Required: true},
		}},
		{Type: "toggle_studio_mode", Name: "Toggle Studio Mode"},
		{Type: "enable_studio_mode", Name: "Enable Studio Mode"},
		{Type: "disable_studio_mode", Name: "Disable Studio Mode"},
	}
}

// ExecuteAction dispatches a button action to the OBS WebSocket API.
func (c *Controller) ExecuteAction(action models.ButtonAction) error {
	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()

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

	case "set_preview_scene":
		sceneName, ok := action.Params["scene_name"].(string)
		if !ok {
			return fmt.Errorf("missing scene_name parameter")
		}
		_, err := client.Scenes.SetCurrentPreviewScene(&scenes.SetCurrentPreviewSceneParams{
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
			SceneName: &sceneName, SourceName: &sourceName,
		})
		if err != nil {
			return err
		}
		itemID := itemResp.SceneItemId
		stateResp, err := client.SceneItems.GetSceneItemEnabled(&sceneitems.GetSceneItemEnabledParams{
			SceneName: &sceneName, SceneItemId: &itemID,
		})
		if err != nil {
			return err
		}
		newState := !stateResp.SceneItemEnabled
		_, err = client.SceneItems.SetSceneItemEnabled(&sceneitems.SetSceneItemEnabledParams{
			SceneName: &sceneName, SceneItemId: &itemID, SceneItemEnabled: &newState,
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
			SceneName: &sceneName, SourceName: &sourceName,
		})
		if err != nil {
			return err
		}
		itemID := itemResp.SceneItemId
		enabled := true
		_, err = client.SceneItems.SetSceneItemEnabled(&sceneitems.SetSceneItemEnabledParams{
			SceneName: &sceneName, SceneItemId: &itemID, SceneItemEnabled: &enabled,
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
			SceneName: &sceneName, SourceName: &sourceName,
		})
		if err != nil {
			return err
		}
		itemID := itemResp.SceneItemId
		enabled := false
		_, err = client.SceneItems.SetSceneItemEnabled(&sceneitems.SetSceneItemEnabledParams{
			SceneName: &sceneName, SceneItemId: &itemID, SceneItemEnabled: &enabled,
		})
		return err

	// ===== AUDIO INPUTS =====
	case "toggle_input_mute":
		inputName, ok := action.Params["input_name"].(string)
		if !ok {
			return fmt.Errorf("missing input_name parameter")
		}
		_, err := client.Inputs.ToggleInputMute(&inputs.ToggleInputMuteParams{InputName: &inputName})
		return err

	case "mute_input":
		inputName, ok := action.Params["input_name"].(string)
		if !ok {
			return fmt.Errorf("missing input_name parameter")
		}
		muted := true
		_, err := client.Inputs.SetInputMute(&inputs.SetInputMuteParams{InputName: &inputName, InputMuted: &muted})
		return err

	case "unmute_input":
		inputName, ok := action.Params["input_name"].(string)
		if !ok {
			return fmt.Errorf("missing input_name parameter")
		}
		muted := false
		_, err := client.Inputs.SetInputMute(&inputs.SetInputMuteParams{InputName: &inputName, InputMuted: &muted})
		return err

	case "set_input_volume":
		inputName, ok := action.Params["input_name"].(string)
		if !ok {
			return fmt.Errorf("missing input_name parameter")
		}
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
		vol := volumePercent / 100.0
		if vol < 0 {
			vol = 0
		}
		if vol > 1 {
			vol = 1
		}
		_, err := client.Inputs.SetInputVolume(&inputs.SetInputVolumeParams{
			InputName: &inputName, InputVolumeMul: &vol,
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
		stateResp, err := client.Filters.GetSourceFilter(&filters.GetSourceFilterParams{
			SourceName: &sourceName, FilterName: &filterName,
		})
		if err != nil {
			return err
		}
		newState := !stateResp.FilterEnabled
		_, err = client.Filters.SetSourceFilterEnabled(&filters.SetSourceFilterEnabledParams{
			SourceName: &sourceName, FilterName: &filterName, FilterEnabled: &newState,
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
			SourceName: &sourceName, FilterName: &filterName, FilterEnabled: &enabled,
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
			SourceName: &sourceName, FilterName: &filterName, FilterEnabled: &enabled,
		})
		return err

	// ===== MEDIA CONTROLS =====
	case "play_pause_media", "restart_media", "stop_media", "next_media", "previous_media":
		return fmt.Errorf("media controls require goobs v1.4+")

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
		var duration float64
		switch d := action.Params["duration"].(type) {
		case float64:
			duration = d
		case string:
			var err error
			duration, err = strconv.ParseFloat(d, 64)
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
		stateResp, err := client.Ui.GetStudioModeEnabled()
		if err != nil {
			return err
		}
		newState := !stateResp.StudioModeEnabled
		_, err = client.Ui.SetStudioModeEnabled(&ui.SetStudioModeEnabledParams{StudioModeEnabled: &newState})
		return err

	case "enable_studio_mode":
		enabled := true
		_, err := client.Ui.SetStudioModeEnabled(&ui.SetStudioModeEnabledParams{StudioModeEnabled: &enabled})
		return err

	case "disable_studio_mode":
		enabled := false
		_, err := client.Ui.SetStudioModeEnabled(&ui.SetStudioModeEnabledParams{StudioModeEnabled: &enabled})
		return err

	default:
		return fmt.Errorf("unknown OBS action type: %s", action.Type)
	}
}

// ==================== OBS-specific methods ====================
// These are not part of the generic Controller interface; app.go accesses
// them via a type assertion on the value returned from registry.Get("obs").

// Connect establishes the OBS WebSocket connection.
// On success it starts a background health-check monitor that will attempt
// up to obsMaxReconnects automatic reconnects if the connection drops.
func (c *Controller) Connect(url, password string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Cancel any previously running monitor.
	if c.cancelMonitor != nil {
		c.cancelMonitor()
		c.cancelMonitor = nil
	}

	if c.client != nil {
		c.client.Disconnect()
		c.client = nil
	}

	client, err := goobs.New(url, goobs.WithPassword(password))
	if err != nil {
		return fmt.Errorf("failed to connect to OBS: %w", err)
	}

	c.client = client
	c.url = url
	c.password = password
	c.reconnectAttempts = 0
	c.reconnectGaveUp = false

	ctx, cancel := context.WithCancel(context.Background())
	c.cancelMonitor = cancel
	go c.connectionMonitor(ctx, url, password)

	return nil
}

// Disconnect closes the OBS WebSocket connection and stops the monitor.
func (c *Controller) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cancelMonitor != nil {
		c.cancelMonitor()
		c.cancelMonitor = nil
	}
	if c.client != nil {
		c.client.Disconnect()
		c.client = nil
	}
	c.reconnectAttempts = 0
	c.reconnectGaveUp = false
	return nil
}

// connectionMonitor runs in a goroutine after a successful Connect.
// It periodically checks the OBS connection and, if it drops, attempts
// up to obsMaxReconnects reconnects before giving up.
// The goroutine exits when ctx is cancelled (i.e. on Disconnect or a new Connect).
func (c *Controller) connectionMonitor(ctx context.Context, url, password string) {
	ticker := time.NewTicker(obsHealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.mu.RLock()
			client := c.client
			c.mu.RUnlock()

			if client == nil {
				return // Already disconnected externally.
			}

			if _, err := client.Stream.GetStreamStatus(); err == nil {
				continue // Still healthy.
			}

			// Connection dropped — nil the client so the UI shows disconnected.
			log.Printf("obs: connection to OBS lost, will attempt to reconnect")
			c.mu.Lock()
			if c.client != nil {
				c.client.Disconnect()
				c.client = nil
			}
			c.mu.Unlock()

			// Try up to obsMaxReconnects times.
			for attempt := 1; attempt <= obsMaxReconnects; attempt++ {
				select {
				case <-ctx.Done():
					return
				case <-time.After(obsReconnectDelay):
				}

				log.Printf("obs: reconnect attempt %d/%d…", attempt, obsMaxReconnects)
				newClient, connErr := goobs.New(url, goobs.WithPassword(password))
				if connErr == nil {
					c.mu.Lock()
					c.client = newClient
					c.reconnectAttempts = 0
					c.mu.Unlock()
					log.Printf("obs: reconnected to OBS successfully")
					break // Resume health-check loop.
				}

				log.Printf("obs: reconnect attempt %d/%d failed: %v", attempt, obsMaxReconnects, connErr)
				c.mu.Lock()
				c.reconnectAttempts = attempt
				if attempt == obsMaxReconnects {
					c.reconnectGaveUp = true
				}
				c.mu.Unlock()

				if attempt == obsMaxReconnects {
					log.Printf("obs: gave up reconnecting to OBS after %d attempts", obsMaxReconnects)
					return
				}
			}
		}
	}
}

// GetURL returns the currently connected OBS WebSocket URL.
func (c *Controller) GetURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.url
}

// GetScenes returns the list of OBS scene names.
func (c *Controller) GetScenes() ([]string, error) {
	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()
	if client == nil {
		return nil, fmt.Errorf("not connected to OBS")
	}
	resp, err := client.Scenes.GetSceneList()
	if err != nil {
		return nil, err
	}
	names := make([]string, len(resp.Scenes))
	for i, s := range resp.Scenes {
		names[i] = s.SceneName
	}
	log.Printf("🎬 GetScenes returning %d scenes: %v", len(names), names)
	return names, nil
}

// GetInputs returns the list of OBS input names.
func (c *Controller) GetInputs() ([]string, error) {
	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()
	if client == nil {
		return nil, fmt.Errorf("not connected to OBS")
	}
	resp, err := client.Inputs.GetInputList(&inputs.GetInputListParams{})
	if err != nil {
		return nil, err
	}
	names := make([]string, len(resp.Inputs))
	for i, inp := range resp.Inputs {
		names[i] = inp.InputName
	}
	log.Printf("🎤 GetInputs returning %d inputs: %v", len(names), names)
	return names, nil
}

// GetInputMute returns whether an input is currently muted.
func (c *Controller) GetInputMute(inputName string) (bool, error) {
	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()
	if client == nil {
		return false, fmt.Errorf("not connected to OBS")
	}
	resp, err := client.Inputs.GetInputMute(&inputs.GetInputMuteParams{InputName: &inputName})
	if err != nil {
		return false, err
	}
	return resp.InputMuted, nil
}

// GetSourceFilterEnabled returns whether a named filter on a source is enabled.
func (c *Controller) GetSourceFilterEnabled(sourceName, filterName string) (bool, error) {
	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()
	if client == nil {
		return false, fmt.Errorf("not connected to OBS")
	}
	resp, err := client.Filters.GetSourceFilter(&filters.GetSourceFilterParams{
		SourceName: &sourceName, FilterName: &filterName,
	})
	if err != nil {
		return false, err
	}
	return resp.FilterEnabled, nil
}

// GetSourceVisibility returns whether a source is visible in the given scene.
func (c *Controller) GetSourceVisibility(sceneName, sourceName string) (bool, error) {
	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()
	if client == nil {
		return false, fmt.Errorf("not connected to OBS")
	}
	itemResp, err := client.SceneItems.GetSceneItemId(&sceneitems.GetSceneItemIdParams{
		SceneName: &sceneName, SourceName: &sourceName,
	})
	if err != nil {
		return false, err
	}
	itemID := itemResp.SceneItemId
	stateResp, err := client.SceneItems.GetSceneItemEnabled(&sceneitems.GetSceneItemEnabledParams{
		SceneName: &sceneName, SceneItemId: &itemID,
	})
	if err != nil {
		return false, err
	}
	return stateResp.SceneItemEnabled, nil
}
