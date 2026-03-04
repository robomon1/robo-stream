package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/hypebeast/go-osc/osc"
)

// zoomOSCClient communicates with ZoomOSC via OSC over UDP.
//
// ZoomOSC command/feedback architecture (same as QLab, TouchOSC, Isadora):
//
//	We SEND commands  → ZoomOSC receiving port (default 9090, UDP)
//	We RECEIVE feedback ← ZoomOSC transmission port → our listen port (default 1234, UDP)
//
// ZoomOSC must be configured in its Settings panel:
//
//	Transmission IP:   127.0.0.1  (or the Robo-Stream server IP)
//	Transmission Port: <listenPort>  (default 1234)
//
// On startup we send /zoom/subscribe 2 (All) so ZoomOSC emits user-state
// events (mute/videoOn/handRaised etc.) for all participants including
// ourselves. Without a subscribe call, ZoomOSC defaults to subscribe mode 0
// (None) and may not send those events.
type zoomOSCClient struct {
	sendHost   string
	sendPort   int
	listenPort int

	mu          sync.RWMutex
	muted       bool
	videoOn     bool
	sharing     bool
	handRaised  bool
	lastPingAck time.Time

	// stateKnown becomes true once ZoomOSC sends at least one real state
	// message (mute/video/share/hand). Until then the boolean fields above
	// are their zero-value defaults and should not be trusted as ground truth.
	stateKnown bool

	// Per-field known flags — each becomes true only when ZoomOSC has sent
	// at least one event for that specific state domain.
	//
	// ZoomOSC only sends state-change events, not a full snapshot on connect.
	// If the server starts while a meeting is already in progress (e.g. video
	// was already on before we connected), ZoomOSC will not send a videoOn
	// event until video state actually changes. The per-field flags let
	// ComputeIndicator suppress indicators for fields whose state has not yet
	// been confirmed, rather than trusting a misleading zero value.
	//
	// NOTE: the ZoomOSC pong response does NOT contain audio/video/share state.
	// Pong fields are: pingArg, version, subscribeMode, galTrackMode,
	// inCallStatus, numTargets, numUsersInCall, isPro.
	// State is only available via event-driven outputs (/zoomosc/me/mute etc.)
	// or via the PRO-only /zoomosc/me/list output.
	muteKnown       bool
	videoKnown      bool
	sharingKnown    bool
	handRaisedKnown bool
}

func newZoomOSCClient(sendPort, listenPort int) *zoomOSCClient {
	c := &zoomOSCClient{
		sendHost:   "127.0.0.1",
		sendPort:   sendPort,
		listenPort: listenPort,
	}

	go c.startListener()
	go c.pingLoop()

	// Send subscribe + initial ping shortly after startup.
	// /zoom/subscribe 2 (All) tells ZoomOSC to send us user-state events for
	// all participants. Without this, ZoomOSC defaults to subscribe mode 0
	// (None) and may not emit mute/video change events.
	go func() {
		time.Sleep(500 * time.Millisecond)
		_ = c.sendOSC("/zoom/subscribe", int32(2))
		_ = c.sendOSC("/zoom/ping")
	}()

	return c
}

// Subscribe sends /zoom/subscribe with the given mode.
// Modes: 0=None, 1=TargetList, 2=All, 3=Panelists, 4=OnlyGallery
func (c *zoomOSCClient) Subscribe(mode int32) error {
	return c.sendOSC("/zoom/subscribe", mode)
}

// IsConnected returns true if ZoomOSC has responded to a ping within the last 15 seconds.
func (c *zoomOSCClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return !c.lastPingAck.IsZero() && time.Since(c.lastPingAck) < 15*time.Second
}

// oscState is a snapshot of the current ZoomOSC-reported state.
type oscState struct {
	Muted      bool
	VideoOn    bool
	Sharing    bool
	HandRaised bool
	// StateKnown is true once at least one real state message has been received.
	StateKnown bool
	// Per-field known flags — see zoomOSCClient fields for explanation.
	MuteKnown       bool
	VideoKnown      bool
	SharingKnown    bool
	HandRaisedKnown bool
}

// State returns a snapshot of the last-known state from ZoomOSC feedback.
func (c *zoomOSCClient) State() oscState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return oscState{
		Muted:           c.muted,
		VideoOn:         c.videoOn,
		Sharing:         c.sharing,
		HandRaised:      c.handRaised,
		StateKnown:      c.stateKnown,
		MuteKnown:       c.muteKnown,
		VideoKnown:      c.videoKnown,
		SharingKnown:    c.sharingKnown,
		HandRaisedKnown: c.handRaisedKnown,
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Command helpers
// ──────────────────────────────────────────────────────────────────────────────

func (c *zoomOSCClient) Ping() error        { return c.sendOSC("/zoom/ping") }
func (c *zoomOSCClient) Mute() error        { return c.sendOSC("/zoom/me/mute") }
func (c *zoomOSCClient) UnMute() error      { return c.sendOSC("/zoom/me/unMute") }
func (c *zoomOSCClient) ToggleMute() error  { return c.sendOSC("/zoom/me/toggleMute") }
func (c *zoomOSCClient) VideoOn() error     { return c.sendOSC("/zoom/me/videoOn") }
func (c *zoomOSCClient) VideoOff() error    { return c.sendOSC("/zoom/me/videoOff") }
func (c *zoomOSCClient) ToggleVideo() error { return c.sendOSC("/zoom/me/toggleVideo") }

// LeaveMeeting sends a global leave-meeting command (not scoped to a user).
func (c *zoomOSCClient) LeaveMeeting() error { return c.sendOSC("/zoom/leaveMeeting") }

// EndMeeting ends the meeting for all participants (host only).
func (c *zoomOSCClient) EndMeeting() error { return c.sendOSC("/zoom/endMeeting") }

// Screen share.
// StartShare uses startScreenSharePrimary (non-PRO). The PRO startScreenShare
// requires a screenID argument. stopShare stops any active share (screen,
// window, camera, or audio).
func (c *zoomOSCClient) StartShare() error { return c.sendOSC("/zoom/me/startScreenSharePrimary") }
func (c *zoomOSCClient) StopShare() error  { return c.sendOSC("/zoom/me/stopShare") }

// ToggleShare sends stop or start depending on current tracked state.
// ZoomOSC has no toggleShare command, so we synthesize it.
func (c *zoomOSCClient) ToggleShare() error {
	c.mu.RLock()
	sharing := c.sharing
	c.mu.RUnlock()
	if sharing {
		return c.StopShare()
	}
	return c.StartShare()
}

// Hand raise / lower.
func (c *zoomOSCClient) RaiseHand() error  { return c.sendOSC("/zoom/me/raiseHand") }
func (c *zoomOSCClient) LowerHand() error  { return c.sendOSC("/zoom/me/lowerHand") }
func (c *zoomOSCClient) ToggleHand() error { return c.sendOSC("/zoom/me/toggleHand") }

// Spotlight — pin your own video for all participants (host or co-host only).
func (c *zoomOSCClient) SpotlightSelf() error   { return c.sendOSC("/zoom/me/spot") }
func (c *zoomOSCClient) UnspotlightSelf() error { return c.sendOSC("/zoom/me/unSpot") }

// applyOptimisticAction updates local state immediately after a command is
// sent. ZoomOSC only emits events when state changes, so if the server
// connects mid-meeting we may not have confirmed state yet. Setting the
// optimistic value after a button press keeps indicators responsive and sets
// the field_known flag so ComputeIndicator can show a meaningful result while
// we wait for ZoomOSC to confirm via a real event.
func (c *zoomOSCClient) applyOptimisticAction(actionType string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch actionType {
	case "toggle_audio":
		c.muted = !c.muted
		c.muteKnown = true
	case "mute_audio":
		c.muted = true
		c.muteKnown = true
	case "unmute_audio":
		c.muted = false
		c.muteKnown = true
	case "toggle_video":
		c.videoOn = !c.videoOn
		c.videoKnown = true
	case "start_video":
		c.videoOn = true
		c.videoKnown = true
	case "stop_video":
		c.videoOn = false
		c.videoKnown = true
	case "toggle_share":
		c.sharing = !c.sharing
		c.sharingKnown = true
	case "start_share":
		c.sharing = true
		c.sharingKnown = true
	case "stop_share":
		c.sharing = false
		c.sharingKnown = true
	case "raise_hand":
		c.handRaised = true
		c.handRaisedKnown = true
	case "lower_hand":
		c.handRaised = false
		c.handRaisedKnown = true
	case "toggle_hand":
		c.handRaised = !c.handRaised
		c.handRaisedKnown = true
	default:
		return
	}

	c.stateKnown = c.muteKnown || c.videoKnown || c.sharingKnown || c.handRaisedKnown
}

// ──────────────────────────────────────────────────────────────────────────────
// Internal
// ──────────────────────────────────────────────────────────────────────────────

// sendOSC sends a single OSC message to ZoomOSC.
func (c *zoomOSCClient) sendOSC(addr string, args ...interface{}) error {
	client := osc.NewClient(c.sendHost, c.sendPort)
	msg := osc.NewMessage(addr, args...)
	return client.Send(msg)
}

// pingLoop sends /zoom/ping to ZoomOSC every 5 seconds to detect connectivity.
func (c *zoomOSCClient) pingLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		_ = c.sendOSC("/zoom/ping")
	}
}

// startListener opens a UDP socket and listens for OSC feedback from ZoomOSC.
// This is a blocking call and should be run in a goroutine.
func (c *zoomOSCClient) startListener() {
	d := &allMsgDispatcher{handler: c.handleMessage}

	server := &osc.Server{
		Addr:        fmt.Sprintf("0.0.0.0:%d", c.listenPort),
		Dispatcher:  d,
		ReadTimeout: 0,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Printf("zoomosc: OSC listener on port %d stopped: %v", c.listenPort, err)
	}
}

// fmtArgs formats OSC message arguments for logging.
func fmtArgs(msg *osc.Message) string {
	if len(msg.Arguments) == 0 {
		return "(no args)"
	}
	s := ""
	for i, a := range msg.Arguments {
		if i > 0 {
			s += " "
		}
		s += fmt.Sprintf("%T:%v", a, a)
	}
	return s
}

// handleMessage updates internal state from a ZoomOSC feedback message.
func (c *zoomOSCClient) handleMessage(msg *osc.Message) {
	c.mu.Lock()
	defer c.mu.Unlock()

	log.Printf("zoomosc: OSC rx: %s  args=%s", msg.Address, fmtArgs(msg))

	// Any message from ZoomOSC means it is alive.
	c.lastPingAck = time.Now()

	// realState becomes true for any message that carries actual user state.
	// Used below to set the stateKnown flag.
	realState := true

	switch msg.Address {

	// ── Audio ──────────────────────────────────────────────────────────────
	// ZoomOSC sends /zoomosc/me/mute when we (or the host) mute our mic, and
	// /zoomosc/me/unMute when we unmute. These are the only reliable non-PRO
	// sources of audio state — the pong response does NOT contain mute state.
	case "/zoomosc/me/mute",
		"/zoomosc/me/audioMuted":
		c.muted = true
		c.muteKnown = true

	case "/zoomosc/me/unMute",
		"/zoomosc/me/unmute",
		"/zoomosc/me/audioUnmuted":
		c.muted = false
		c.muteKnown = true

	// ── Video ──────────────────────────────────────────────────────────────
	// ZoomOSC sends /zoomosc/me/videoOn / videoOff when camera state changes.
	// Same caveat as audio: not in pong, event-driven only (non-PRO).
	case "/zoomosc/me/videoOn",
		"/zoomosc/me/videoEnabled":
		c.videoOn = true
		c.videoKnown = true

	case "/zoomosc/me/videoOff",
		"/zoomosc/me/videoDisabled":
		c.videoOn = false
		c.videoKnown = true

	// ── Screen Share ───────────────────────────────────────────────────────
	// ZoomOSC does not document a standardised share-state event in the free
	// tier. We handle a few common paths seen across ZoomOSC versions.
	// Share state is only confirmable via the PRO /zoomosc/me/list output.
	case "/zoomosc/me/videoShareStarted",
		"/zoomosc/me/sharingScreen":
		c.sharing = true
		c.sharingKnown = true

	case "/zoomosc/me/videoShareStopped",
		"/zoomosc/me/stoppedSharingScreen":
		c.sharing = false
		c.sharingKnown = true

	// ── Hand Raise ─────────────────────────────────────────────────────────
	case "/zoomosc/me/handRaised",
		"/zoomosc/me/raiseHand":
		c.handRaised = true
		c.handRaisedKnown = true

	case "/zoomosc/me/handLowered",
		"/zoomosc/me/lowerHand":
		c.handRaised = false
		c.handRaisedKnown = true

	// ── Pong ───────────────────────────────────────────────────────────────
	// Documented pong format (ZoomOSC 4.6):
	//   [0] any    — pingArg (zero if none sent)
	//   [1] string — zoomOSCversion (e.g. "ZOSC_4.6.1_MAC")
	//   [2] int    — subscribeMode (0=None,1=TargetList,2=All,3=Panelists,4=OnlyGallery)
	//   [3] int    — galTrackMode
	//   [4] int    — inCallStatus (0=not in meeting, 1=in meeting)
	//   [5] int    — number of targets currently selected
	//   [6] int    — number of users in call
	//   [7] int    — isPro (1=true, 0=false)
	//
	// IMPORTANT: pong does NOT carry audio mute, video on/off, or share state.
	// We only use arg[4] (inCallStatus) to clear stale state when leaving a
	// meeting. Audio/video state comes from /zoomosc/me/mute, videoOn, etc.
	case "/zoomosc/pong":
		realState = false
		if len(msg.Arguments) >= 5 {
			if inCallStatus, ok := toInt64(msg.Arguments[4]); ok && inCallStatus == 0 {
				// Not in a meeting — clear all known flags so stale state from
				// a previous session does not bleed into the next one.
				c.muteKnown = false
				c.videoKnown = false
				c.sharingKnown = false
				c.handRaisedKnown = false
				c.stateKnown = false
			}
		}

	case "/zoomosc/ping":
		realState = false

	// ── Meeting status ─────────────────────────────────────────────────────
	// Zoom SDK status codes: 0=idle, 1=connecting, 3=in meeting,
	// 4=disconnecting, 6=failed, 7=ended by host.
	// Any status other than 3 means we left (or never joined) the meeting.
	case "/zoomosc/meetingStatusChanged":
		realState = false
		if len(msg.Arguments) >= 1 {
			if status, ok := toInt64(msg.Arguments[0]); ok && status != 3 {
				c.muteKnown = false
				c.videoKnown = false
				c.sharingKnown = false
				c.handRaisedKnown = false
				c.stateKnown = false
			}
		}

	default:
		// Non-fatal; many ZoomOSC outputs (gallery order, user events for
		// other participants, etc.) are legitimately unhandled.
	}

	if realState {
		c.stateKnown = true
	}
}

func toInt64(v interface{}) (int64, bool) {
	switch n := v.(type) {
	case int:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return n, true
	case uint:
		return int64(n), true
	case uint32:
		return int64(n), true
	case uint64:
		return int64(n), true
	case float32:
		return int64(n), true
	case float64:
		return int64(n), true
	case bool:
		if n {
			return 1, true
		}
		return 0, true
	case string:
		i, err := strconv.ParseInt(n, 10, 64)
		return i, err == nil
	default:
		return 0, false
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// allMsgDispatcher — routes every OSC packet to a single handler function.
// ──────────────────────────────────────────────────────────────────────────────

type allMsgDispatcher struct {
	handler func(*osc.Message)
}

func (d *allMsgDispatcher) Dispatch(p osc.Packet) {
	switch v := p.(type) {
	case *osc.Message:
		d.handler(v)
	case *osc.Bundle:
		for _, msg := range v.Messages {
			d.handler(msg)
		}
	}
}
