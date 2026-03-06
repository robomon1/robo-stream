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
			"muted":       state.Muted,
			"video":       state.VideoOn,
			"sharing":     state.Sharing,
			"hand_raised": state.HandRaised,
			// state_known is false until ZoomOSC sends at least one real state
			// message across ANY domain. Kept for backward compatibility.
			"state_known": state.StateKnown,
			// Per-field known flags — each is only true once ZoomOSC has sent
			// at least one event for that specific state domain. ZoomOSC only
			// emits change events, not a full snapshot on connect, so if the
			// server starts mid-meeting the per-field flags stay false for
			// domains that haven't changed yet. ComputeIndicator uses these to
			// suppress indicators for unconfirmed fields rather than trusting
			// zero-value defaults.
			"muted_known":       state.MuteKnown,
			"video_known":       state.VideoKnown,
			"sharing_known":     state.SharingKnown,
			"hand_raised_known": state.HandRaisedKnown,
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
		// ── Audio ──────────────────────────────────────────────────────────
		// toggle_audio: white indicator when mic is live (NOT muted)
		{Type: "toggle_audio", Name: "Toggle Microphone", Description: "Toggle mic mute/unmute",
			IndicatorField: "muted", IndicatorInvert: true},
		// mute_audio: white indicator when mic IS muted
		{Type: "mute_audio", Name: "Mute Microphone", Description: "Mute the microphone",
			IndicatorField: "muted", IndicatorInvert: false},
		// unmute_audio: white indicator when mic IS live (not muted)
		{Type: "unmute_audio", Name: "Unmute Microphone", Description: "Unmute the microphone",
			IndicatorField: "muted", IndicatorInvert: true},

		// ── Video ──────────────────────────────────────────────────────────
		// toggle_video: white indicator when camera IS on
		{Type: "toggle_video", Name: "Toggle Camera", Description: "Toggle camera on/off",
			IndicatorField: "video", IndicatorInvert: false},
		// start_video: white indicator when camera IS on
		{Type: "start_video", Name: "Start Camera", Description: "Turn the camera on",
			IndicatorField: "video", IndicatorInvert: false},
		// stop_video: white indicator when camera IS off
		{Type: "stop_video", Name: "Stop Camera", Description: "Turn the camera off",
			IndicatorField: "video", IndicatorInvert: true},

		// ── Screen Share ───────────────────────────────────────────────────
		// toggle_share / start_share: white indicator when sharing IS active
		{Type: "toggle_share", Name: "Toggle Screen Share", Description: "Start or stop screen sharing",
			IndicatorField: "sharing", IndicatorInvert: false},
		{Type: "start_share", Name: "Start Screen Share", Description: "Start sharing your screen",
			IndicatorField: "sharing", IndicatorInvert: false},
		// stop_share: white indicator when sharing IS stopped
		{Type: "stop_share", Name: "Stop Screen Share", Description: "Stop sharing your screen",
			IndicatorField: "sharing", IndicatorInvert: true},

		// ── Hand Raise ─────────────────────────────────────────────────────
		// raise_hand: white indicator when hand IS raised
		{Type: "raise_hand", Name: "Raise Hand", Description: "Raise your hand",
			IndicatorField: "hand_raised", IndicatorInvert: false},
		// lower_hand: white indicator when hand IS lowered
		{Type: "lower_hand", Name: "Lower Hand", Description: "Lower your hand",
			IndicatorField: "hand_raised", IndicatorInvert: true},
		// NOTE: toggle_hand (/zoom/me/toggleHand) is not supported by ZoomOSC non-PRO.

		// ── Meeting ────────────────────────────────────────────────────────
		// No indicator for leave/end — these are one-shot destructive actions.
		{Type: "leave_meeting", Name: "Leave Meeting", Description: "Leave the current meeting"},
		{Type: "end_meeting", Name: "End Meeting (Host)", Description: "End the meeting for all participants (host only)"},
		// NOTE: spotlight_self / unspotlight_self require host/co-host privileges
		// and are not available to regular participants (PRO feature).
	}
}

func (z *ZoomOSCController) Execute(action sdk.ExecuteRequest) error {
	var err error
	switch action.Type {
	// Audio
	case "toggle_audio":
		err = z.client.ToggleMute()
	case "mute_audio":
		err = z.client.Mute()
	case "unmute_audio":
		err = z.client.UnMute()
	// Video
	case "toggle_video":
		err = z.client.ToggleVideo()
	case "start_video":
		err = z.client.VideoOn()
	case "stop_video":
		err = z.client.VideoOff()
	// Screen share
	case "toggle_share":
		err = z.client.ToggleShare()
	case "start_share":
		err = z.client.StartShare()
	case "stop_share":
		err = z.client.StopShare()
	// Hand raise
	case "raise_hand":
		err = z.client.RaiseHand()
	case "lower_hand":
		err = z.client.LowerHand()
	// Meeting
	case "leave_meeting":
		err = z.client.LeaveMeeting()
	case "end_meeting":
		err = z.client.EndMeeting()
	default:
		return fmt.Errorf("unknown ZoomOSC action: %s", action.Type)
	}

	if err == nil {
		// Apply optimistic state immediately so indicators respond before
		// ZoomOSC confirms the change via an event. ZoomOSC will send the
		// real /zoomosc/me/mute, videoOn, etc. event shortly after, which
		// overwrites the optimistic value with the confirmed truth.
		// NOTE: we do NOT send a post-action ping here. ZoomOSC's pong does
		// not carry audio/video/share state (only inCallStatus, numTargets,
		// numUsersInCall, isPro), so pinging would not help. State updates
		// come exclusively from event-driven outputs.
		z.client.applyOptimisticAction(action.Type)
	}
	return err
}

func (z *ZoomOSCController) DefaultButtons() []sdk.Button {
	c := "zoomosc"
	return []sdk.Button{
		{
			Name:        "Toggle Audio",
			Description: "Toggle microphone mute/unmute via ZoomOSC",
			Icon:        "mic",
			Color:       "#2d8cff",
			Action:      sdk.ExecuteRequest{Controller: c, Type: "toggle_audio"},
		},
		{
			Name:        "Mute Audio",
			Description: "Mute the microphone via ZoomOSC",
			Icon:        "mic-off",
			Color:       "#2d8cff",
			Action:      sdk.ExecuteRequest{Controller: c, Type: "mute_audio"},
		},
		{
			Name:        "Unmute Audio",
			Description: "Unmute the microphone via ZoomOSC",
			Icon:        "mic",
			Color:       "#2d8cff",
			Action:      sdk.ExecuteRequest{Controller: c, Type: "unmute_audio"},
		},
		{
			Name:        "Toggle Video",
			Description: "Toggle camera on/off via ZoomOSC",
			Icon:        "video",
			Color:       "#2d8cff",
			Action:      sdk.ExecuteRequest{Controller: c, Type: "toggle_video"},
		},
		{
			Name:        "Stop Video",
			Description: "Turn the camera off via ZoomOSC",
			Icon:        "video",
			Color:       "#2d8cff",
			Action:      sdk.ExecuteRequest{Controller: c, Type: "stop_video"},
		},
		{
			Name:        "Start Video",
			Description: "Turn the camera on via ZoomOSC",
			Icon:        "video",
			Color:       "#2d8cff",
			Action:      sdk.ExecuteRequest{Controller: c, Type: "start_video"},
		},
		{
			Name:        "Raise Hand",
			Description: "Raise your hand via ZoomOSC",
			Icon:        "hand",
			Color:       "#2d8cff",
			Action:      sdk.ExecuteRequest{Controller: c, Type: "raise_hand"},
		},
		{
			Name:        "Lower Hand",
			Description: "Lower your hand via ZoomOSC",
			Icon:        "hand",
			Color:       "#2d8cff",
			Action:      sdk.ExecuteRequest{Controller: c, Type: "lower_hand"},
		},
		{
			Name:        "Stop Screen Share",
			Description: "Stop sharing your screen via ZoomOSC",
			Icon:        "monitor",
			Color:       "#2d8cff",
			Action:      sdk.ExecuteRequest{Controller: c, Type: "stop_share"},
		},
		{
			Name:        "Start Screen Share",
			Description: "Start sharing your screen via ZoomOSC",
			Icon:        "monitor",
			Color:       "#2d8cff",
			Action:      sdk.ExecuteRequest{Controller: c, Type: "start_share"},
		},
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Internal helpers
// ──────────────────────────────────────────────────────────────────────────────

func (z *ZoomOSCController) applyConfig(cfg map[string]interface{}) error {
	// Port config changes require restarting the client; apply only the send
	// port at runtime.
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
