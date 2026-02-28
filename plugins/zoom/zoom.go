package main

import (
	"fmt"

	"github.com/robomon1/robo-stream/sdk"
)

// ZoomController implements sdk.PluginImplementation.
// It controls the Zoom desktop client via its local HTTP API on localhost:19421.
type ZoomController struct {
	client *zoomClient
	config map[string]interface{}
}

func newZoomController() *ZoomController {
	return &ZoomController{
		client: newZoomClient(),
		config: make(map[string]interface{}),
	}
}

// ==================== sdk.PluginImplementation ====================

func (z *ZoomController) Info() sdk.PluginInfo {
	return sdk.PluginInfo{
		ID:          "zoom",
		Name:        "Zoom",
		Description: "Controls the Zoom desktop client (mute, camera, leave, share)",
		Version:     "1.0.0",
		Icon:        "video",
		Author:      "Robo-Stream",
	}
}

func (z *ZoomController) Initialize(cfg map[string]interface{}) error {
	z.config = cfg
	// No persistent connection to maintain — the Zoom API is stateless HTTP.
	// Just verify Zoom is reachable.
	if !z.client.IsRunning() {
		// Not an error — Zoom may not be open yet. Log and continue.
		return nil
	}
	return nil
}

func (z *ZoomController) Status() sdk.PluginStatus {
	if !z.client.IsRunning() {
		return sdk.PluginStatus{
			Connected: false,
			Error:     "Zoom desktop client is not running or not in a meeting",
		}
	}

	meeting, err := z.client.GetMeeting()
	if err != nil {
		return sdk.PluginStatus{Connected: false, Error: err.Error()}
	}

	inMeeting := meeting.Status == "IN_MEETING"
	details := map[string]interface{}{
		"in_meeting":      inMeeting,
		"meeting_id":      meeting.ID,
		"meeting_topic":   meeting.Topic,
		"participant_num": meeting.ParticipantNum,
	}

	if inMeeting {
		if audio, err := z.client.GetAudioStatus(); err == nil {
			details["muted"] = audio.Muted
		}
		if video, err := z.client.GetVideoStatus(); err == nil {
			details["video_stopped"] = video.Stopped
		}
		if share, err := z.client.GetShareStatus(); err == nil {
			details["sharing"] = share.Sharing
		}
	}

	return sdk.PluginStatus{Connected: true, Details: details}
}

func (z *ZoomController) ConfigSchema() []sdk.ConfigField {
	// Zoom auto-discovers via localhost:19421, no credentials needed.
	// Expose an optional port override for non-standard setups.
	return []sdk.ConfigField{
		{
			Key:     "port",
			Label:   "Zoom API Port",
			Type:    "number",
			Default: "19421",
			Help:    "Local port of the Zoom desktop client API. Default is 19421.",
		},
	}
}

func (z *ZoomController) GetConfig() map[string]interface{} {
	cfg := make(map[string]interface{}, len(z.config))
	for k, v := range z.config {
		cfg[k] = v
	}
	return cfg
}

func (z *ZoomController) UpdateConfig(cfg map[string]interface{}) error {
	z.config = cfg
	// Apply optional port override
	if port, ok := cfg["port"].(string); ok && port != "" {
		z.client.baseURL = fmt.Sprintf("http://localhost:%s", port)
	}
	return nil
}

func (z *ZoomController) SupportedActions() []sdk.ActionTypeDef {
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

func (z *ZoomController) Execute(action sdk.ExecuteRequest) error {
	switch action.Type {
	case "toggle_audio":
		return z.client.ToggleMute()
	case "mute_audio":
		return z.client.SetMute(true)
	case "unmute_audio":
		return z.client.SetMute(false)
	case "toggle_video":
		return z.client.ToggleVideo()
	case "start_video":
		return z.client.SetVideo(true)
	case "stop_video":
		return z.client.SetVideo(false)
	case "leave_meeting":
		return z.client.LeaveMeeting()
	case "toggle_share":
		return z.client.ToggleShare()
	case "start_share":
		return z.client.StartShare()
	case "stop_share":
		return z.client.StopShare()
	default:
		return fmt.Errorf("unknown Zoom action: %s", action.Type)
	}
}

func (z *ZoomController) DefaultButtons() []sdk.Button {
	c := "zoom"
	return []sdk.Button{
		{
			Name:        "Toggle Mic",
			Description: "Toggle microphone mute",
			Icon:        "mic",
			Color:       "#2d8cff",
			Action:      sdk.ExecuteRequest{Controller: c, Type: "toggle_audio"},
		},
		{
			Name:        "Toggle Camera",
			Description: "Toggle camera on/off",
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
			Description: "Toggle screen sharing",
			Icon:        "monitor",
			Color:       "#27ae60",
			Action:      sdk.ExecuteRequest{Controller: c, Type: "toggle_share"},
		},
	}
}
