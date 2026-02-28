package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// zoomClient talks to the Zoom desktop client's local REST API.
// Zoom exposes this API on localhost:19421 when it is running and a meeting
// is active. It is the same API used by Elgato Stream Deck and Loupedeck.
type zoomClient struct {
	baseURL    string
	httpClient *http.Client
}

func newZoomClient() *zoomClient {
	return &zoomClient{
		baseURL: "http://localhost:19421",
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// meetingInfo is returned by GET /meeting.
type meetingInfo struct {
	ID             string `json:"id"`
	Topic          string `json:"topic"`
	Status         string `json:"status"` // "IN_MEETING", "NOT_IN_MEETING"
	ParticipantNum int    `json:"participant_num"`
}

// audioStatus is returned by GET /mute.
type audioStatus struct {
	Muted bool `json:"muted"`
}

// videoStatus is returned by GET /video.
type videoStatus struct {
	Stopped bool `json:"stopped"` // true = video is off
}

// shareStatus is returned by GET /share.
type shareStatus struct {
	Sharing bool `json:"sharing"`
}

// IsRunning returns true if the Zoom desktop client is running and reachable.
func (z *zoomClient) IsRunning() bool {
	resp, err := z.httpClient.Get(z.baseURL + "/meeting")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode < 500
}

// GetMeeting returns current meeting information.
func (z *zoomClient) GetMeeting() (*meetingInfo, error) {
	return getJSON[meetingInfo](z, "/meeting")
}

// GetAudioStatus returns whether the local participant's microphone is muted.
func (z *zoomClient) GetAudioStatus() (*audioStatus, error) {
	return getJSON[audioStatus](z, "/mute")
}

// GetVideoStatus returns whether the local participant's camera is stopped.
func (z *zoomClient) GetVideoStatus() (*videoStatus, error) {
	return getJSON[videoStatus](z, "/video")
}

// GetShareStatus returns whether the local participant is sharing their screen.
func (z *zoomClient) GetShareStatus() (*shareStatus, error) {
	return getJSON[shareStatus](z, "/share")
}

// SetMute sets the mute state of the local participant's microphone.
func (z *zoomClient) SetMute(muted bool) error {
	return z.postJSON("/mute", map[string]bool{"mute": muted})
}

// ToggleMute toggles the mute state of the local participant's microphone.
func (z *zoomClient) ToggleMute() error {
	status, err := z.GetAudioStatus()
	if err != nil {
		return err
	}
	return z.SetMute(!status.Muted)
}

// SetVideo sets whether the local participant's camera is active.
// video=true turns the camera on; video=false turns it off.
func (z *zoomClient) SetVideo(on bool) error {
	return z.postJSON("/video", map[string]bool{"video": on})
}

// ToggleVideo toggles the local participant's camera.
func (z *zoomClient) ToggleVideo() error {
	status, err := z.GetVideoStatus()
	if err != nil {
		return err
	}
	return z.SetVideo(status.Stopped) // if stopped, turn on; if on, turn off
}

// LeaveMeeting leaves the current meeting.
func (z *zoomClient) LeaveMeeting() error {
	return z.postJSON("/end", map[string]bool{"leave": true})
}

// StartShare starts screen sharing.
func (z *zoomClient) StartShare() error {
	return z.postJSON("/share", map[string]bool{"start": true})
}

// StopShare stops screen sharing.
func (z *zoomClient) StopShare() error {
	return z.postJSON("/share", map[string]bool{"start": false})
}

// ToggleShare toggles screen sharing.
func (z *zoomClient) ToggleShare() error {
	status, err := z.GetShareStatus()
	if err != nil {
		return err
	}
	return z.postJSON("/share", map[string]bool{"start": !status.Sharing})
}

// ==================== internal helpers ====================

func getJSON[T any](z *zoomClient, path string) (*T, error) {
	resp, err := z.httpClient.Get(z.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("zoom API unavailable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("zoom API %s returned %d: %s", path, resp.StatusCode, body)
	}
	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("zoom API %s: decode error: %w", path, err)
	}
	return &result, nil
}

func (z *zoomClient) postJSON(path string, body interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	resp, err := z.httpClient.Post(z.baseURL+path, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("zoom API unavailable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("zoom API %s returned %d: %s", path, resp.StatusCode, b)
	}
	return nil
}
