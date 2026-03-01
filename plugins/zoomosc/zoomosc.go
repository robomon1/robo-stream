package main

import (
	"fmt"

	"github.com/robomon1/robo-stream/sdk"
)

const (
	defaultZoomOSCPort = 9090 // ZoomOSC receiving port (we SEND to this)
	defaultListenPort  = 1234 // Our feedback port (ZoomOSC SENDS to this)
)

// ZoomOSCController implements sdk.PluginImplementation.
// It controls the Zoom desktop client by sending OSC commands to ZoomOSC
// (https://www.liminalet.com/zoomosc) and receiving OSC feedback from it.
type ZoomOSCController struct {
	client *zoomOSCClient
	config map[string]interface{}
}

func newZoomOSCController() *ZoomOSCController {
	return &ZoomOSCController{
		client: newZoomOSCClient(defaultZoomOSCPort, defaultListenPort),
		config: make(map[string]interface{}),
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// sdk.PluginImplementation
// ──────────────────────────────────────────────────────────────────────────────

func (z *ZoomOSCController) Info() sdk.PluginInfo {
	return sdk.PluginInfo{
		ID:          "zoomosc",
		Name:        "ZoomOSC",
		Description: "Controls Zoom via ZoomOSC (mute, camera, leave, share). Requires ZoomOSC by Liminal.",
		Version:     "1.0.0",
		Icon:        "video",
		Author:      "Robo-Stream",
	}
}

func (z *ZoomOSCController) Initialize(cfg map[string]interface{}) error {
	z.config = cfg
	return z.applyConfig(cfg)
}

func (z *ZoomOSCController) Status() sdk.PluginStatus {
	if !z.client.IsConnected() {
		return sdk.PluginStatus{
			Connected: false,
			Error:     "ZoomOSC is not responding. Ensure ZoomOSC is running and its Transmission IP/Port point to this plugin.",
		}
	}

	state := z.client.State()
	return sdk.PluginStatus{
		Connected: true,
		Details: map[string]interface{}{
			"muted":   state.Muted,
			"video":   state.VideoOn,
			"sharing": state.Sharing,
		},
	}
}

func (z *ZoomOSCController) ConfigSchema() []sdk.ConfigField {
	return []sdk.ConfigField{
		{
			Key:     "zoomosc_port",
			Label:   "ZoomOSC Receiving Port",
			Type:    "number",
			Default: fmt.Sprintf("%d", defaultZoomOSCPort),
			Help:    "UDP port ZoomOSC listens on for commands. Match the 'Receiving Port' in ZoomOSC settings (default 9090).",
		},
		{
			Key:     "listen_port",
			Label:   "Feedback Listen Port",
			Type:    "number",
			Default: fmt.Sprintf("%d", defaultListenPort),
			Help:    "UDP port this plugin listens on for ZoomOSC feedback. Set ZoomOSC 'Transmission Port' to this value (default 1234).",
		},
	}
}

func (z *ZoomOSCController) GetConfig() map[string]interface{} {
	cfg := make(map[string]interface{}, len(z.config))
	for k, v := range z.config {
		cfg[k] = v
	}
	return cfg
}

func (z *ZoomOSCController) UpdateConfig(cfg map[string]interface{}) error {
	z.config = cfg
	return z.applyConfig(cfg)
}

func (z *ZoomOSCController) SupportedActions() []sdk.ActionTypeDef {
	return []sdk.ActionTypeDef{
		{Type: "toggle_audio", Name: "Toggle Microphone", Description: "Toggle mic mute/unmute"},
		{Type: "mute_audio", Name: "Mute Microphone", Description: "Mute the microphone"},
		{Type: "unmute_audio", Name: "Unmute Microphone", Description: "Unmute the microphone"},
		{Type: "toggle_video", Name: "Toggle Camera", Description: "Toggle camera on/off"},
		{Type: "start_video", Name: "Start Camera", Description: "Turn the camera on"},
		{Type: "stop_video", Name: "Stop Camera", Description: "Turn the camera off"},
		{Type: "leave_meeting", Name: "Leave Meeting", Description: "Leave the current meeting"},
		{Type: "toggle_share", Name: "Toggle Screen Share", Description: "Start or stop screen sharing"},
		{Type: "start_share", Name: "Start Screen Share", Description: "Start sharing your screen"},
		{Type: "stop_share", Name: "Stop Screen Share", Description: "Stop sharing your screen"},
	}
}

func (z *ZoomOSCController) Execute(action sdk.ExecuteRequest) error {
	switch action.Type {
	case "toggle_audio":
		return z.client.ToggleMute()
	case "mute_audio":
		return z.client.Mute()
	case "unmute_audio":
		return z.client.UnMute()
	case "toggle_video":
		return z.client.ToggleVideo()
	case "start_video":
		return z.client.VideoOn()
	case "stop_video":
		return z.client.VideoOff()
	case "leave_meeting":
		return z.client.LeaveMeeting()
	case "toggle_share":
		return z.client.ToggleShare()
	case "start_share":
		return z.client.StartShare()
	case "stop_share":
		return z.client.StopShare()
	default:
		return fmt.Errorf("unknown ZoomOSC action: %s", action.Type)
	}
}

func (z *ZoomOSCController) DefaultButtons() []sdk.Button {
	c := "zoomosc"
	return []sdk.Button{
		{
			Name:        "Toggle Mic",
			Description: "Toggle microphone mute via ZoomOSC",
			Icon:        "mic",
			Color:       "#2d8cff",
			Action:      sdk.ExecuteRequest{Controller: c, Type: "toggle_audio"},
		},
		{
			Name:        "Toggle Camera",
			Description: "Toggle camera on/off via ZoomOSC",
			Icon:        "video",
			Color:       "#2d8cff",
			Action:      sdk.ExecuteRequest{Controller: c, Type: "toggle_video"},
		},
		{
			Name:        "Leave Meeting",
			Description: "Leave the current Zoom meeting",
			Icon:        "phone-off",
			Color:       "#e74c3c",
			Action:      sdk.ExecuteRequest{Controller: c, Type: "leave_meeting"},
		},
		{
			Name:        "Share Screen",
			Description: "Toggle screen sharing via ZoomOSC",
			Icon:        "monitor",
			Color:       "#27ae60",
			Action:      sdk.ExecuteRequest{Controller: c, Type: "toggle_share"},
		},
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Internal helpers
// ──────────────────────────────────────────────────────────────────────────────

func (z *ZoomOSCController) applyConfig(cfg map[string]interface{}) error {
	// Port config changes require restarting the client; for now, just update
	// the send port (the listen port can't change without restarting the server).
	if v, ok := cfg["zoomosc_port"]; ok {
		switch p := v.(type) {
		case float64:
			z.client.sendPort = int(p)
		case int:
			z.client.sendPort = p
		case string:
			var port int
			if _, err := fmt.Sscanf(p, "%d", &port); err == nil {
				z.client.sendPort = port
			}
		}
	}
	return nil
}
