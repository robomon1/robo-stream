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
	lastPingAck time.Time
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
	Muted   bool
	VideoOn bool
	Sharing bool
}

// State returns a snapshot of the last-known state from ZoomOSC feedback.
func (c *zoomOSCClient) State() oscState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return oscState{
		Muted:   c.muted,
		VideoOn: c.videoOn,
		Sharing: c.sharing,
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Command helpers
// ──────────────────────────────────────────────────────────────────────────────

func (c *zoomOSCClient) Mute() error        { return c.sendOSC("/zoom/me/mute") }
func (c *zoomOSCClient) UnMute() error      { return c.sendOSC("/zoom/me/unMute") }
func (c *zoomOSCClient) ToggleMute() error  { return c.sendOSC("/zoom/me/toggleMute") }
func (c *zoomOSCClient) VideoOn() error     { return c.sendOSC("/zoom/me/videoOn") }
func (c *zoomOSCClient) VideoOff() error    { return c.sendOSC("/zoom/me/videoOff") }
func (c *zoomOSCClient) ToggleVideo() error { return c.sendOSC("/zoom/me/toggleVideo") }

// LeaveMeeting sends a global leave-meeting command (not scoped to a user).
func (c *zoomOSCClient) LeaveMeeting() error { return c.sendOSC("/zoom/leaveMeeting") }

// Screen share — ZoomOSC 4.x uses /zoom/me/startShare and /zoom/me/stopShare.
func (c *zoomOSCClient) StartShare() error  { return c.sendOSC("/zoom/me/startShare") }
func (c *zoomOSCClient) StopShare() error   { return c.sendOSC("/zoom/me/stopShare") }
func (c *zoomOSCClient) ToggleShare() error { return c.sendOSC("/zoom/me/toggleShare") }

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

// handleMessage updates internal state from a ZoomOSC feedback message.
func (c *zoomOSCClient) handleMessage(msg *osc.Message) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Any message from ZoomOSC means it is alive.
	c.lastPingAck = time.Now()

	switch msg.Address {
	// Audio state
	case "/zoomosc/me/mute":
		c.muted = true
	case "/zoomosc/me/unMute":
		c.muted = false

	// Video state
	case "/zoomosc/me/videoOn":
		c.videoOn = true
	case "/zoomosc/me/videoOff":
		c.videoOn = false

	// Screen share state (ZoomOSC uses different path names across versions)
	case "/zoomosc/me/shareOn", "/zoomosc/me/startShare":
		c.sharing = true
	case "/zoomosc/me/shareOff", "/zoomosc/me/stopShare":
		c.sharing = false

	// Ping/pong — already handled by the lastPingAck update above.
	case "/zoomosc/ping", "/zoomosc/pong":
		// no-op; state already updated

	default:
		// Log unknown paths to help users troubleshoot ZoomOSC configuration.
		log.Printf("zoomosc: unhandled feedback: %s", msg.Address)
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
