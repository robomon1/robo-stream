package actions

import (
	"fmt"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/inputs"
	"github.com/andreykaipov/goobs/api/requests/sceneitems"
	"github.com/sirupsen/logrus"
)

// SourceManager handles source and audio-related OBS actions
type SourceManager struct {
	client *goobs.Client
	logger *logrus.Logger
}

// NewSourceManager creates a new source manager
func NewSourceManager(client *goobs.Client, logger *logrus.Logger) *SourceManager {
	if logger == nil {
		logger = logrus.New()
	}
	return &SourceManager{
		client: client,
		logger: logger,
	}
}

// SetMute sets the mute state of an input
// Maps from OBS WebSocket 4.x: setMute -> 5.x: SetInputMute
func (sm *SourceManager) SetMute(inputName string, muted bool) error {
	if sm.client == nil {
		return fmt.Errorf("not connected to OBS")
	}

	sm.logger.Infof("Setting mute for %s to %v", inputName, muted)

	req := &inputs.SetInputMuteParams{
		InputName: &inputName,
		InputMuted: &muted,
	}

	_, err := sm.client.Inputs.SetInputMute(req)
	if err != nil {
		sm.logger.Errorf("Failed to set mute: %v", err)
		return fmt.Errorf("failed to set mute: %w", err)
	}

	sm.logger.Infof("Successfully set mute for %s to %v", inputName, muted)
	return nil
}

// ToggleMute toggles the mute state of an input
// Maps from OBS WebSocket 4.x: toggleMute -> 5.x: ToggleInputMute
func (sm *SourceManager) ToggleMute(inputName string) error {
	if sm.client == nil {
		return fmt.Errorf("not connected to OBS")
	}

	sm.logger.Infof("Toggling mute for %s", inputName)

	req := &inputs.ToggleInputMuteParams{
		InputName: &inputName,
	}

	resp, err := sm.client.Inputs.ToggleInputMute(req)
	if err != nil {
		sm.logger.Errorf("Failed to toggle mute: %v", err)
		return fmt.Errorf("failed to toggle mute: %w", err)
	}

	sm.logger.Infof("Toggled mute for %s: now muted=%v", inputName, resp.InputMuted)
	return nil
}

// GetMute gets the mute state of an input
func (sm *SourceManager) GetMute(inputName string) (bool, error) {
	if sm.client == nil {
		return false, fmt.Errorf("not connected to OBS")
	}

	req := &inputs.GetInputMuteParams{
		InputName: &inputName,
	}

	resp, err := sm.client.Inputs.GetInputMute(req)
	if err != nil {
		return false, fmt.Errorf("failed to get mute state: %w", err)
	}

	return resp.InputMuted, nil
}

// SetVolume sets the volume of an input
// Maps from OBS WebSocket 4.x: setVolume -> 5.x: SetInputVolume
// Volume is in dB (-100.0 to 26.0, with 0.0 being unity)
func (sm *SourceManager) SetVolume(inputName string, volumeDb float64) error {
	if sm.client == nil {
		return fmt.Errorf("not connected to OBS")
	}

	sm.logger.Infof("Setting volume for %s to %f dB", inputName, volumeDb)

	req := &inputs.SetInputVolumeParams{
		InputName:    &inputName,
		InputVolumeDb: &volumeDb,
	}

	_, err := sm.client.Inputs.SetInputVolume(req)
	if err != nil {
		sm.logger.Errorf("Failed to set volume: %v", err)
		return fmt.Errorf("failed to set volume: %w", err)
	}

	sm.logger.Infof("Successfully set volume for %s to %f dB", inputName, volumeDb)
	return nil
}

// GetVolume gets the volume of an input in dB
func (sm *SourceManager) GetVolume(inputName string) (float64, error) {
	if sm.client == nil {
		return 0, fmt.Errorf("not connected to OBS")
	}

	req := &inputs.GetInputVolumeParams{
		InputName: &inputName,
	}

	resp, err := sm.client.Inputs.GetInputVolume(req)
	if err != nil {
		return 0, fmt.Errorf("failed to get volume: %w", err)
	}

	return resp.InputVolumeDb, nil
}

// SetSourceVisibility sets the visibility of a scene item
// Maps from OBS WebSocket 4.x: setSourceRender -> 5.x: SetSceneItemEnabled
func (sm *SourceManager) SetSourceVisibility(sceneName string, sceneItemId int, visible bool) error {
	if sm.client == nil {
		return fmt.Errorf("not connected to OBS")
	}

	sm.logger.Infof("Setting visibility for item %d in scene %s to %v", sceneItemId, sceneName, visible)

	req := &sceneitems.SetSceneItemEnabledParams{
		SceneName:       &sceneName,
		SceneItemId:     &sceneItemId,
		SceneItemEnabled: &visible,
	}

	_, err := sm.client.SceneItems.SetSceneItemEnabled(req)
	if err != nil {
		sm.logger.Errorf("Failed to set source visibility: %v", err)
		return fmt.Errorf("failed to set source visibility: %w", err)
	}

	sm.logger.Infof("Successfully set visibility for item %d in scene %s to %v", sceneItemId, sceneName, visible)
	return nil
}

// GetSourceVisibility gets the visibility of a scene item
func (sm *SourceManager) GetSourceVisibility(sceneName string, sceneItemId int) (bool, error) {
	if sm.client == nil {
		return false, fmt.Errorf("not connected to OBS")
	}

	req := &sceneitems.GetSceneItemEnabledParams{
		SceneName:   &sceneName,
		SceneItemId: &sceneItemId,
	}

	resp, err := sm.client.SceneItems.GetSceneItemEnabled(req)
	if err != nil {
		return false, fmt.Errorf("failed to get source visibility: %w", err)
	}

	return resp.SceneItemEnabled, nil
}

// ToggleSourceVisibility toggles the visibility of a scene item
func (sm *SourceManager) ToggleSourceVisibility(sceneName string, sceneItemId int) error {
	if sm.client == nil {
		return fmt.Errorf("not connected to OBS")
	}

	// Get current state
	visible, err := sm.GetSourceVisibility(sceneName, sceneItemId)
	if err != nil {
		return err
	}

	// Toggle
	return sm.SetSourceVisibility(sceneName, sceneItemId, !visible)
}

// GetSceneItemId gets the ID of a scene item by its name
func (sm *SourceManager) GetSceneItemId(sceneName string, sourceName string) (int, error) {
	if sm.client == nil {
		return 0, fmt.Errorf("not connected to OBS")
	}

	req := &sceneitems.GetSceneItemIdParams{
		SceneName:  &sceneName,
		SourceName: &sourceName,
	}

	resp, err := sm.client.SceneItems.GetSceneItemId(req)
	if err != nil {
		return 0, fmt.Errorf("failed to get scene item ID: %w", err)
	}

	return resp.SceneItemId, nil
}

// GetInputList gets a list of all inputs
func (sm *SourceManager) GetInputList() ([]string, error) {
	if sm.client == nil {
		return nil, fmt.Errorf("not connected to OBS")
	}

	// Get all inputs
	resp, err := sm.client.Inputs.GetInputList(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get input list: %w", err)
	}

	inputNames := make([]string, len(resp.Inputs))
	for i, input := range resp.Inputs {
		inputNames[i] = input.InputName
	}

	return inputNames, nil
}

// SetInputSettings updates the settings of an input
func (sm *SourceManager) SetInputSettings(inputName string, settings map[string]interface{}) error {
	if sm.client == nil {
		return fmt.Errorf("not connected to OBS")
	}

	sm.logger.Infof("Setting settings for input %s", inputName)

	req := &inputs.SetInputSettingsParams{
		InputName:     &inputName,
		InputSettings: settings,
	}

	_, err := sm.client.Inputs.SetInputSettings(req)
	if err != nil {
		sm.logger.Errorf("Failed to set input settings: %v", err)
		return fmt.Errorf("failed to set input settings: %w", err)
	}

	sm.logger.Infof("Successfully set settings for input %s", inputName)
	return nil
}

// GetInputSettings gets the settings of an input
func (sm *SourceManager) GetInputSettings(inputName string) (map[string]interface{}, error) {
	if sm.client == nil {
		return nil, fmt.Errorf("not connected to OBS")
	}

	req := &inputs.GetInputSettingsParams{
		InputName: &inputName,
	}

	resp, err := sm.client.Inputs.GetInputSettings(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get input settings: %w", err)
	}

	return resp.InputSettings, nil
}
