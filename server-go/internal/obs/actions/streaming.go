package actions

import (
	"fmt"

	"github.com/andreykaipov/goobs"
	// "github.com/andreykaipov/goobs/api/requests/record"
	"github.com/andreykaipov/goobs/api/requests/stream"
	"github.com/sirupsen/logrus"
)

// StreamManager handles streaming and recording operations
type StreamManager struct {
	client *goobs.Client
	logger *logrus.Logger
}

// NewStreamManager creates a new stream manager
func NewStreamManager(client *goobs.Client, logger *logrus.Logger) *StreamManager {
	return &StreamManager{
		client: client,
		logger: logger,
	}
}

// StartStreaming starts streaming
func (sm *StreamManager) StartStreaming() error {
	if sm.client == nil {
		return fmt.Errorf("not connected to OBS")
	}

	sm.logger.Info("Starting stream")

	_, err := sm.client.Stream.StartStream(nil)
	if err != nil {
		sm.logger.Errorf("Failed to start stream: %v", err)
		return fmt.Errorf("failed to start stream: %w", err)
	}

	sm.logger.Info("Successfully started stream")
	return nil
}

// StopStreaming stops streaming
func (sm *StreamManager) StopStreaming() error {
	if sm.client == nil {
		return fmt.Errorf("not connected to OBS")
	}

	sm.logger.Info("Stopping stream")

	_, err := sm.client.Stream.StopStream(nil)
	if err != nil {
		sm.logger.Errorf("Failed to stop stream: %v", err)
		return fmt.Errorf("failed to stop stream: %w", err)
	}

	sm.logger.Info("Successfully stopped stream")
	return nil
}

// ToggleStreaming toggles streaming on/off
func (sm *StreamManager) ToggleStreaming() error {
	if sm.client == nil {
		return fmt.Errorf("not connected to OBS")
	}

	sm.logger.Info("Toggling stream")

	resp, err := sm.client.Stream.ToggleStream(nil)
	if err != nil {
		sm.logger.Errorf("Failed to toggle stream: %v", err)
		return fmt.Errorf("failed to toggle stream: %w", err)
	}

	sm.logger.Infof("Stream toggled: active=%v", resp.OutputActive)
	return nil
}

// GetStreamStatus gets the current streaming status
func (sm *StreamManager) GetStreamStatus() (bool, error) {
	if sm.client == nil {
		return false, fmt.Errorf("not connected to OBS")
	}

	resp, err := sm.client.Stream.GetStreamStatus()
	if err != nil {
		return false, fmt.Errorf("failed to get stream status: %w", err)
	}

	return resp.OutputActive, nil
}

// SendStreamCaption sends a caption/text to the stream
func (sm *StreamManager) SendStreamCaption(caption string) error {
	if sm.client == nil {
		return fmt.Errorf("not connected to OBS")
	}

	req := &stream.SendStreamCaptionParams{
		CaptionText: &caption,
	}

	_, err := sm.client.Stream.SendStreamCaption(req)
	if err != nil {
		return fmt.Errorf("failed to send stream caption: %w", err)
	}

	return nil
}

// StartRecording starts recording
func (sm *StreamManager) StartRecording() error {
	if sm.client == nil {
		return fmt.Errorf("not connected to OBS")
	}

	sm.logger.Info("Starting recording")

	_, err := sm.client.Record.StartRecord(nil)
	if err != nil {
		sm.logger.Errorf("Failed to start recording: %v", err)
		return fmt.Errorf("failed to start recording: %w", err)
	}

	sm.logger.Info("Successfully started recording")
	return nil
}

// StopRecording stops recording
func (sm *StreamManager) StopRecording() error {
	if sm.client == nil {
		return fmt.Errorf("not connected to OBS")
	}

	sm.logger.Info("Stopping recording")

	_, err := sm.client.Record.StopRecord(nil)
	if err != nil {
		sm.logger.Errorf("Failed to stop recording: %v", err)
		return fmt.Errorf("failed to stop recording: %w", err)
	}

	sm.logger.Info("Successfully stopped recording")
	return nil
}

// ToggleRecording toggles recording on/off
func (sm *StreamManager) ToggleRecording() error {
	if sm.client == nil {
		return fmt.Errorf("not connected to OBS")
	}

	sm.logger.Info("Toggling recording")

	resp, err := sm.client.Record.ToggleRecord(nil)
	if err != nil {
		sm.logger.Errorf("Failed to toggle recording: %v", err)
		return fmt.Errorf("failed to toggle recording: %w", err)
	}

	sm.logger.Infof("Recording toggled: active=%v", resp.OutputActive)
	return nil
}

// PauseRecording pauses recording
func (sm *StreamManager) PauseRecording() error {
	if sm.client == nil {
		return fmt.Errorf("not connected to OBS")
	}

	sm.logger.Info("Pausing recording")

	_, err := sm.client.Record.PauseRecord(nil)
	if err != nil {
		sm.logger.Errorf("Failed to pause recording: %v", err)
		return fmt.Errorf("failed to pause recording: %w", err)
	}

	sm.logger.Info("Successfully paused recording")
	return nil
}

// ResumeRecording resumes recording
func (sm *StreamManager) ResumeRecording() error {
	if sm.client == nil {
		return fmt.Errorf("not connected to OBS")
	}

	sm.logger.Info("Resuming recording")

	_, err := sm.client.Record.ResumeRecord(nil)
	if err != nil {
		sm.logger.Errorf("Failed to resume recording: %v", err)
		return fmt.Errorf("failed to resume recording: %w", err)
	}

	sm.logger.Info("Successfully resumed recording")
	return nil
}

// ToggleRecordPause toggles recording pause state
func (sm *StreamManager) ToggleRecordPause() error {
	if sm.client == nil {
		return fmt.Errorf("not connected to OBS")
	}

	sm.logger.Info("Toggling recording pause")

	resp, err := sm.client.Record.ToggleRecordPause(nil)
	if err != nil {
		sm.logger.Errorf("Failed to toggle record pause: %v", err)
		return fmt.Errorf("failed to toggle record pause: %w", err)
	}

	sm.logger.Infof("Recording pause toggled: paused=%v", resp.OutputPaused)
	return nil
}

// GetRecordStatus gets the current recording status
func (sm *StreamManager) GetRecordStatus() (bool, bool, error) {
	if sm.client == nil {
		return false, false, fmt.Errorf("not connected to OBS")
	}

	resp, err := sm.client.Record.GetRecordStatus()
	if err != nil {
		return false, false, fmt.Errorf("failed to get record status: %w", err)
	}

	return resp.OutputActive, resp.OutputPaused, nil
}

// NOTE: Replay Buffer methods are not included as they require additional
// OBS configuration and may not be available in all versions.
// They can be added later based on the actual goobs API for replay buffers.
