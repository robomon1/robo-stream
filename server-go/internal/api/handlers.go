package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/robomon1/stream-pi-go/server-go/internal/obs"
	"github.com/robomon1/stream-pi-go/server-go/internal/obs/actions"
	"github.com/sirupsen/logrus"
)

// Handler handles API requests
type Handler struct {
	manager *obs.Manager
	logger  *logrus.Logger
}

// NewHandler creates a new API handler
func NewHandler(manager *obs.Manager, logger *logrus.Logger) *Handler {
	return &Handler{
		manager: manager,
		logger:  logger,
	}
}

// ActionRequest represents a client action request
type ActionRequest struct {
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params,omitempty"`
}

// ActionResponse represents an action response
type ActionResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// HandleAction executes an OBS action
func (h *Handler) HandleAction(w http.ResponseWriter, r *http.Request) {
	var req ActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.logger.Infof("Executing action: %s with params: %v", req.Action, req.Params)

	client := h.manager.Client()
	if client == nil {
		respondError(w, "Not connected to OBS", http.StatusServiceUnavailable)
		return
	}

	var err error
	var data interface{}

	switch req.Action {
	case "switch_scene":
		sceneName := getStringParam(req.Params, "scene_name")
		err = actions.SetCurrentScene(client, sceneName)

	case "toggle_stream":
		err = actions.ToggleStreaming(client)

	case "start_stream":
		err = actions.StartStreaming(client)

	case "stop_stream":
		err = actions.StopStreaming(client)

	case "toggle_record":
		err = actions.ToggleRecording(client)

	case "start_record":
		err = actions.StartRecording(client)

	case "stop_record":
		err = actions.StopRecording(client)

	case "pause_record":
		err = actions.PauseRecording(client)

	case "resume_record":
		err = actions.ResumeRecording(client)

	case "toggle_input_mute":
		inputName := getStringParam(req.Params, "input_name")
		err = actions.ToggleInputMute(client, inputName)

	case "mute_input":
		inputName := getStringParam(req.Params, "input_name")
		err = actions.SetInputMute(client, inputName, true)

	case "unmute_input":
		inputName := getStringParam(req.Params, "input_name")
		err = actions.SetInputMute(client, inputName, false)

	case "set_source_visibility":
		sourceName := getStringParam(req.Params, "source_name")
		visible := getBoolParam(req.Params, "visible")
		err = actions.SetSourceVisibility(client, sourceName, visible)

	default:
		respondError(w, fmt.Sprintf("Unknown action: %s", req.Action), http.StatusBadRequest)
		return
	}

	if err != nil {
		h.logger.Errorf("Action failed: %v", err)
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondSuccess(w, "Action executed successfully", data)
}

// HandleGetStatus returns current OBS status
func (h *Handler) HandleGetStatus(w http.ResponseWriter, r *http.Request) {
	client := h.manager.Client()
	if client == nil {
		respondError(w, "Not connected to OBS", http.StatusServiceUnavailable)
		return
	}

	streamStatus, err := actions.GetStreamStatus(client)
	if err != nil {
		h.logger.Errorf("Failed to get stream status: %v", err)
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	recordStatus, err := actions.GetRecordStatus(client)
	if err != nil {
		h.logger.Errorf("Failed to get record status: %v", err)
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	currentScene, err := actions.GetCurrentScene(client)
	if err != nil {
		h.logger.Errorf("Failed to get current scene: %v", err)
		currentScene = "Unknown"
	}

	status := map[string]interface{}{
		"streaming":     streamStatus.Active,
		"recording":     recordStatus.Active,
		"paused":        recordStatus.Paused,
		"current_scene": currentScene,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// HandleGetScenes returns list of available scenes
func (h *Handler) HandleGetScenes(w http.ResponseWriter, r *http.Request) {
	client := h.manager.Client()
	if client == nil {
		respondError(w, "Not connected to OBS", http.StatusServiceUnavailable)
		return
	}

	scenes, err := actions.GetSceneList(client)
	if err != nil {
		h.logger.Errorf("Failed to get scenes: %v", err)
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scenes": scenes,
	})
}

// HandleGetInputs returns list of available audio inputs
func (h *Handler) HandleGetInputs(w http.ResponseWriter, r *http.Request) {
	client := h.manager.Client()
	if client == nil {
		respondError(w, "Not connected to OBS", http.StatusServiceUnavailable)
		return
	}

	inputs, err := actions.GetInputList(client)
	if err != nil {
		h.logger.Errorf("Failed to get inputs: %v", err)
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"inputs": inputs,
	})
}

// Helper functions

func getStringParam(params map[string]interface{}, key string) string {
	if val, ok := params[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getBoolParam(params map[string]interface{}, key string) bool {
	if val, ok := params[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func respondSuccess(w http.ResponseWriter, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ActionResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func respondError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ActionResponse{
		Success: false,
		Message: message,
	})
}
