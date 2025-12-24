package actions

import (
	"fmt"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/scenes"
	"github.com/sirupsen/logrus"
)

// SceneManager handles scene-related OBS actions
type SceneManager struct {
	client *goobs.Client
	logger *logrus.Logger
}

// NewSceneManager creates a new scene manager
func NewSceneManager(client *goobs.Client, logger *logrus.Logger) *SceneManager {
	if logger == nil {
		logger = logrus.New()
	}
	return &SceneManager{
		client: client,
		logger: logger,
	}
}

// SetCurrentScene sets the current program scene
// Maps from OBS WebSocket 4.x: setCurrentScene -> 5.x: SetCurrentProgramScene
func (sm *SceneManager) SetCurrentScene(sceneName string) error {
	if sm.client == nil {
		return fmt.Errorf("not connected to OBS")
	}

	sm.logger.Infof("Setting current scene to: %s", sceneName)

	req := &scenes.SetCurrentProgramSceneParams{
		SceneName: &sceneName,
	}

	_, err := sm.client.Scenes.SetCurrentProgramScene(req)
	if err != nil {
		sm.logger.Errorf("Failed to set current scene: %v", err)
		return fmt.Errorf("failed to set current scene: %w", err)
	}

	sm.logger.Infof("Successfully set current scene to: %s", sceneName)
	return nil
}

// SetPreviewScene sets the current preview scene (studio mode)
// Maps from OBS WebSocket 4.x: setPreviewScene -> 5.x: SetCurrentPreviewScene
func (sm *SceneManager) SetPreviewScene(sceneName string) error {
	if sm.client == nil {
		return fmt.Errorf("not connected to OBS")
	}

	sm.logger.Infof("Setting preview scene to: %s", sceneName)

	req := &scenes.SetCurrentPreviewSceneParams{
		SceneName: &sceneName,
	}

	_, err := sm.client.Scenes.SetCurrentPreviewScene(req)
	if err != nil {
		sm.logger.Errorf("Failed to set preview scene: %v", err)
		return fmt.Errorf("failed to set preview scene: %w", err)
	}

	sm.logger.Infof("Successfully set preview scene to: %s", sceneName)
	return nil
}

// GetCurrentScene gets the current program scene name
func (sm *SceneManager) GetCurrentScene() (string, error) {
	if sm.client == nil {
		return "", fmt.Errorf("not connected to OBS")
	}

	resp, err := sm.client.Scenes.GetCurrentProgramScene()
	if err != nil {
		return "", fmt.Errorf("failed to get current scene: %w", err)
	}

	return resp.CurrentProgramSceneName, nil
}

// GetPreviewScene gets the current preview scene name
func (sm *SceneManager) GetPreviewScene() (string, error) {
	if sm.client == nil {
		return "", fmt.Errorf("not connected to OBS")
	}

	resp, err := sm.client.Scenes.GetCurrentPreviewScene()
	if err != nil {
		return "", fmt.Errorf("failed to get preview scene: %w", err)
	}

	return resp.CurrentPreviewSceneName, nil
}

// GetSceneList gets list of all available scenes
func (sm *SceneManager) GetSceneList() ([]string, error) {
	if sm.client == nil {
		return nil, fmt.Errorf("not connected to OBS")
	}

	resp, err := sm.client.Scenes.GetSceneList()
	if err != nil {
		return nil, fmt.Errorf("failed to get scene list: %w", err)
	}

	sceneNames := make([]string, len(resp.Scenes))
	for i, scene := range resp.Scenes {
		sceneNames[i] = scene.SceneName
	}

	return sceneNames, nil
}

// CreateScene creates a new scene
func (sm *SceneManager) CreateScene(sceneName string) error {
	if sm.client == nil {
		return fmt.Errorf("not connected to OBS")
	}

	sm.logger.Infof("Creating scene: %s", sceneName)

	req := &scenes.CreateSceneParams{
		SceneName: &sceneName,
	}

	_, err := sm.client.Scenes.CreateScene(req)
	if err != nil {
		sm.logger.Errorf("Failed to create scene: %v", err)
		return fmt.Errorf("failed to create scene: %w", err)
	}

	sm.logger.Infof("Successfully created scene: %s", sceneName)
	return nil
}

// RemoveScene removes a scene
func (sm *SceneManager) RemoveScene(sceneName string) error {
	if sm.client == nil {
		return fmt.Errorf("not connected to OBS")
	}

	sm.logger.Infof("Removing scene: %s", sceneName)

	req := &scenes.RemoveSceneParams{
		SceneName: &sceneName,
	}

	_, err := sm.client.Scenes.RemoveScene(req)
	if err != nil {
		sm.logger.Errorf("Failed to remove scene: %v", err)
		return fmt.Errorf("failed to remove scene: %w", err)
	}

	sm.logger.Infof("Successfully removed scene: %s", sceneName)
	return nil
}
