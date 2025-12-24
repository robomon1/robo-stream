package actions

import (
	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/inputs"
	"github.com/andreykaipov/goobs/api/requests/sceneitems"
	"github.com/andreykaipov/goobs/api/requests/scenes"
)

// Scene actions
func SetCurrentScene(client *goobs.Client, sceneName string) error {
	params := &scenes.SetCurrentProgramSceneParams{
		SceneName: &sceneName,
	}
	_, err := client.Scenes.SetCurrentProgramScene(params)
	return err
}

func GetCurrentScene(client *goobs.Client) (string, error) {
	resp, err := client.Scenes.GetCurrentProgramScene()
	if err != nil {
		return "", err
	}
	return resp.CurrentProgramSceneName, nil
}

func GetSceneList(client *goobs.Client) ([]string, error) {
	resp, err := client.Scenes.GetSceneList()
	if err != nil {
		return nil, err
	}

	scenes := make([]string, len(resp.Scenes))
	for i, scene := range resp.Scenes {
		scenes[i] = scene.SceneName
	}
	return scenes, nil
}

// Streaming actions
func ToggleStreaming(client *goobs.Client) error {
	_, err := client.Stream.ToggleStream(nil)
	return err
}

func StartStreaming(client *goobs.Client) error {
	_, err := client.Stream.StartStream(nil)
	return err
}

func StopStreaming(client *goobs.Client) error {
	_, err := client.Stream.StopStream(nil)
	return err
}

type StreamStatus struct {
	Active bool
}

func GetStreamStatus(client *goobs.Client) (*StreamStatus, error) {
	resp, err := client.Stream.GetStreamStatus()
	if err != nil {
		return nil, err
	}
	return &StreamStatus{Active: resp.OutputActive}, nil
}

// Recording actions
func ToggleRecording(client *goobs.Client) error {
	_, err := client.Record.ToggleRecord(nil)
	return err
}

func StartRecording(client *goobs.Client) error {
	_, err := client.Record.StartRecord(nil)
	return err
}

func StopRecording(client *goobs.Client) error {
	_, err := client.Record.StopRecord(nil)
	return err
}

func PauseRecording(client *goobs.Client) error {
	_, err := client.Record.PauseRecord(nil)
	return err
}

func ResumeRecording(client *goobs.Client) error {
	_, err := client.Record.ResumeRecord(nil)
	return err
}

type RecordStatus struct {
	Active bool
	Paused bool
}

func GetRecordStatus(client *goobs.Client) (*RecordStatus, error) {
	resp, err := client.Record.GetRecordStatus()
	if err != nil {
		return nil, err
	}
	return &RecordStatus{
		Active: resp.OutputActive,
		Paused: resp.OutputPaused,
	}, nil
}

// Audio input actions
func ToggleInputMute(client *goobs.Client, inputName string) error {
	params := &inputs.ToggleInputMuteParams{
		InputName: &inputName,
	}
	_, err := client.Inputs.ToggleInputMute(params)
	return err
}

func SetInputMute(client *goobs.Client, inputName string, muted bool) error {
	params := &inputs.SetInputMuteParams{
		InputName:  &inputName,
		InputMuted: &muted,
	}
	_, err := client.Inputs.SetInputMute(params)
	return err
}

func GetInputList(client *goobs.Client) ([]string, error) {
	resp, err := client.Inputs.GetInputList(nil)
	if err != nil {
		return nil, err
	}

	inputNames := make([]string, len(resp.Inputs))
	for i, input := range resp.Inputs {
		inputNames[i] = input.InputName
	}
	return inputNames, nil
}

// Source visibility actions
func SetSourceVisibility(client *goobs.Client, sourceName string, visible bool) error {
	// Get current scene first
	currentScene, err := GetCurrentScene(client)
	if err != nil {
		return err
	}

	// Get scene item ID
	itemID, err := GetSceneItemId(client, currentScene, sourceName)
	if err != nil {
		return err
	}

	params := &sceneitems.SetSceneItemEnabledParams{
		SceneName:        &currentScene,
		SceneItemId:      &itemID,
		SceneItemEnabled: &visible,
	}
	_, err = client.SceneItems.SetSceneItemEnabled(params)
	return err
}

func GetSceneItemId(client *goobs.Client, sceneName, sourceName string) (int, error) {
	params := &sceneitems.GetSceneItemIdParams{
		SceneName:  &sceneName,
		SourceName: &sourceName,
	}
	resp, err := client.SceneItems.GetSceneItemId(params)
	if err != nil {
		return 0, err
	}

	// SceneItemId is already an int in goobs
	return resp.SceneItemId, nil
}
