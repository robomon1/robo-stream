package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/hypebeast/go-osc/osc"
)

// zoomOSCClient communicates with ZoomOSC via OSC over UDP.
//
// ZoomOSC command/feedback architecture:
//
//	We SEND commands  → ZoomOSC receiving port (default 9090, UDP)
//	We RECEIVE feedback ← ZoomOSC transmission port → our listen port (default 1234, UDP)
//
// ZoomOSC must be configured in its Settings panel:
//
//	Transmission IP:   127.0.0.1  (or the Robo-Stream server IP)
//	Transmission Port: <listenPort>  (default 1234)
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
	// at least one event for that specific state domain. ZoomOSC only sends
	// state-change events, not the full current state on connect. If the
	// server starts while a meeting is already in progress (e.g. video was
	// already on before we connected), ZoomOSC will never send a videoOn
	// event and videoOn stays false. The per-field flags let ComputeIndicator
	// suppress indicators for fields whose state has not yet been confirmed,
	// rather than trusting a misleading zero value.
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

	// Send an initial ping shortly after startup.
	go func() {
		time.Sleep(500 * time.Millisecond)
		_ = c.sendOSC("/zoom/ping")
	}()

	return c
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
	// StateKnown is true once at least one real state message has been received
	// from ZoomOSC. Until then the other fields are zero-value defaults, not
	// confirmed values, and callers should not treat them as ground truth.
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

// Ping sends a /zoom/ping to ZoomOSC, which will respond with a /zoomosc/pong
// containing current state. Call this after executing an action so the next
// status poll reflects the updated state quickly.
func (c *zoomOSCClient) Ping() error { return c.sendOSC("/zoom/ping") }

func (c *zoomOSCClient) Mute() error        { return c.sendOSC("/zoom/me/mute") }
func (c *zoomOSCClient) UnMute() error      { return c.sendOSC("/zoom/me/unMute") }
func (c *zoomOSCClient) ToggleMute() error  { return c.sendOSC("/zoom/me/toggleMute") }
func (c *zoomOSCClient) VideoOn() error     { return c.sendOSC("/zoom/me/videoOn") }
func (c *zoomOSCClient) VideoOff() error    { return c.sendOSC("/zoom/me/videoOff") }
func (c *zoomOSCClient) ToggleVideo() error { return c.sendOSC("/zoom/me/toggleVideo") }

// LeaveMeeting sends a global leave-meeting command (not scoped to a user).
func (c *zoomOSCClient) LeaveMeeting() error { return c.sendOSC("/zoom/leaveMeeting") }

// EndMeeting ends the meeting for all participants (host only).
func (c *zoomOSCClient) EndMeeting() error { return c.sendOSC("/zoom/meeting/end") }

// Screen share — ZoomOSC 4.x uses /zoom/me/startShare and /zoom/me/stopShare.
func (c *zoomOSCClient) StartShare() error  { return c.sendOSC("/zoom/me/startShare") }
func (c *zoomOSCClient) StopShare() error   { return c.sendOSC("/zoom/me/stopShare") }
func (c *zoomOSCClient) ToggleShare() error { return c.sendOSC("/zoom/me/toggleShare") }

// Hand raise / lower.
func (c *zoomOSCClient) RaiseHand() error  { return c.sendOSC("/zoom/me/raiseHand") }
func (c *zoomOSCClient) LowerHand() error  { return c.sendOSC("/zoom/me/lowerHand") }
func (c *zoomOSCClient) ToggleHand() error { return c.sendOSC("/zoom/me/handToggle") }

// Spotlight — pin your own video for all participants (host or co-host only).
func (c *zoomOSCClient) SpotlightSelf() error   { return c.sendOSC("/zoom/me/spotlight") }
func (c *zoomOSCClient) UnspotlightSelf() error { return c.sendOSC("/zoom/me/unSpotlight") }

// ──────────────────────────────────────────────────────────────────────────────
// Internal
// ──────────────────────────────────────────────────────────────────────────────

// sendOSC sends a single OSC message (no arguments) to ZoomOSC.
func (c *zoomOSCClient) sendOSC(addr string, args ...interface{}) error {
	client := osc.NewClient(c.sendHost, c.sendPort)
	msg := osc.NewMessage(addr, args...)
	return client.Send(msg)
}

// pingLoop sends /zoom/ping to ZoomOSC every 5 seconds.
// ZoomOSC responds with /zoomosc/ping, which updates lastPingAck.
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
	// allMsgDispatcher is a custom Dispatcher that routes every OSC message
	// through a single handler, letting us update state for any ZoomOSC feedback.
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

	// Log every received message — critical for diagnosing path mismatches
	// between ZoomOSC versions. Remove this log once paths are confirmed.
	log.Printf("zoomosc: OSC rx: %s  args=%s", msg.Address, fmtArgs(msg))

	// Any message from ZoomOSC means it is alive.
	c.lastPingAck = time.Now()

	// realState becomes true for any message that carries actual state
	// (not just pings). Used below to set the stateKnown flag.
	realState := true

	switch msg.Address {
	// Audio state — ZoomOSC sends mute/unmute feedback on mic toggle.
	// Multiple aliases cover ZoomOSC 4.x vs 5.x path differences.
	case "/zoomosc/me/mute",
		"/zoomosc/me/audioMuted":
		c.muted = true
		c.muteKnown = true
	case "/zoomosc/me/unMute",   // ZoomOSC's own camelCase
		"/zoomosc/me/unmute",    // lowercase variant seen in some builds
		"/zoomosc/me/audioUnmuted":
		c.muted = false
		c.muteKnown = true

	// Video state
	case "/zoomosc/me/videoOn",
		"/zoomosc/me/videoEnabled":
		c.videoOn = true
		c.videoKnown = true
	case "/zoomosc/me/videoOff",
		"/zoomosc/me/videoDisabled":
		c.videoOn = false
		c.videoKnown = true

	// Screen share state (ZoomOSC uses different path names across versions)
	case "/zoomosc/me/shareOn",
		"/zoomosc/me/startShare",
		"/zoomosc/me/sharingScreen":
		c.sharing = true
		c.sharingKnown = true
	case "/zoomosc/me/shareOff",
		"/zoomosc/me/stopShare":
		c.sharing = false
		c.sharingKnown = true

	// Hand raise state
	case "/zoomosc/me/handRaised",
		"/zoomosc/me/raiseHand":
		c.handRaised = true
		c.handRaisedKnown = true
	case "/zoomosc/me/handLowered",
		"/zoomosc/me/lowerHand":
		c.handRaised = false
		c.handRaisedKnown = true

	// Pong — ZoomOSC 4.x packs the full local-user state into every pong
	// response. It does NOT send discrete /zoomosc/me/videoOn etc. events on
	// connect, so the pong is the only reliable source of initial state.
	//
	// Observed format (ZoomOSC 4.6.1 MAC):
	//   [0] int32  — undocumented (always 0)
	//   [1] string — version string ("ZOSC_4.6.1_MAC")
	//   [2] int32  — undocumented (always 0)
	//   [3] int32  — undocumented (always 1)
	//   [4] int32  — in_meeting  (0 = no, 1 = yes)
	//   [5] int32  — audio_muted (0 = live, 1 = muted)
	//   [6] int32  — video_on    (0 = off,  1 = on)
	//   [7] int32  — sharing     (0 = no,   1 = yes)
	case "/zoomosc/pong":
		realState = false // default; overridden below when in a meeting
		if len(msg.Arguments) >= 8 {
			inMeeting, ok0 := msg.Arguments[4].(int32)
			muted, ok1 := msg.Arguments[5].(int32)
			videoOn, ok2 := msg.Arguments[6].(int32)
			sharing, ok3 := msg.Arguments[7].(int32)
			if ok0 && ok1 && ok2 && ok3 {
				if inMeeting != 0 {
					// In a meeting — parse and store authoritative state.
					c.muted = muted != 0
					c.videoOn = videoOn != 0
					c.sharing = sharing != 0
					c.muteKnown = true
					c.videoKnown = true
					c.sharingKnown = true
					realState = true
				} else {
					// Not in a meeting — clear known flags so stale state from
					// a previous session does not bleed into the next one.
					c.muteKnown = false
					c.videoKnown = false
					c.sharingKnown = false
					c.handRaisedKnown = false
					c.stateKnown = false
				}
			}
		}

	case "/zoomosc/ping":
		realState = false

	// Meeting status changes — use status 3 ("in meeting") as the authoritative
	// in-meeting signal. Any other status means we are not (or no longer) in a
	// meeting; clear known flags so old state does not persist.
	// ZoomOSC status codes (from Zoom SDK): 0=idle, 1=connecting, 3=in meeting,
	// 4=disconnecting, 6=failed, 7=ended by host.
	case "/zoomosc/meetingStatusChanged":
		realState = false
		if len(msg.Arguments) >= 1 {
			if status, ok := msg.Arguments[0].(int32); ok && status != 3 {
				c.muteKnown = false
				c.videoKnown = false
				c.sharingKnown = false
				c.handRaisedKnown = false
				c.stateKnown = false
			}
		}

	default:
		// Log unknown paths to help users troubleshoot ZoomOSC configuration.
		log.Printf("zoomosc: unhandled feedback: %s", msg.Address)
	}

	if realState {
		c.stateKnown = true
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
