package api

import (
	"encoding/json"
	"net/http"

	"client/internal/client"
	"client/internal/config"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler handles API requests
type Handler struct {
	obsClient *client.OBSClient
	config    *config.ButtonConfig
	logger    *logrus.Logger
	hub       *Hub
}

// NewHandler creates a new API handler
func NewHandler(obsClient *client.OBSClient, buttonConfig *config.ButtonConfig, logger *logrus.Logger) *Handler {
	return &Handler{
		obsClient: obsClient,
		config:    buttonConfig,
		logger:    logger,
		hub:       NewHub(),
	}
}

// GetButtons returns all buttons
func (h *Handler) GetButtons(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.config)
}

// UpdateButtons updates all buttons
func (h *Handler) UpdateButtons(w http.ResponseWriter, r *http.Request) {
	var newConfig config.ButtonConfig
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.config.Grid = newConfig.Grid
	h.config.Buttons = newConfig.Buttons

	// Save to file
	if err := config.SaveConfig("configs/buttons.json", h.config); err != nil {
		h.logger.Errorf("Failed to save config: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// UpdateButton updates a single button
func (h *Handler) UpdateButton(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	buttonID := vars["id"]

	var button config.Button
	if err := json.NewDecoder(r.Body).Decode(&button); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	button.ID = buttonID
	h.config.UpdateButton(button)

	// Save to file
	if err := config.SaveConfig("configs/buttons.json", h.config); err != nil {
		h.logger.Errorf("Failed to save config: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// DeleteButton deletes a button
func (h *Handler) DeleteButton(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	buttonID := vars["id"]

	h.config.DeleteButton(buttonID)

	// Save to file
	if err := config.SaveConfig("configs/buttons.json", h.config); err != nil {
		h.logger.Errorf("Failed to save config: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// PressButton handles a button press
func (h *Handler) PressButton(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	buttonID := vars["id"]

	button := h.config.GetButton(buttonID)
	if button == nil {
		http.Error(w, "Button not found", http.StatusNotFound)
		return
	}

	h.logger.Infof("Button pressed: %s (action: %s)", button.Text, button.Action.Type)

	// Execute the action
	resp, err := h.obsClient.ExecuteAction(button.Action.Type, button.Action.Params)
	if err != nil {
		h.logger.Errorf("Failed to execute action: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast status update to all connected WebSocket clients
	go h.broadcastStatus()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetScenes returns available OBS scenes
func (h *Handler) GetScenes(w http.ResponseWriter, r *http.Request) {
	scenes, err := h.obsClient.GetScenes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scenes": scenes,
	})
}

// GetInputs returns available OBS audio inputs
func (h *Handler) GetInputs(w http.ResponseWriter, r *http.Request) {
	inputs, err := h.obsClient.GetInputs()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"inputs": inputs,
	})
}

// GetStatus returns current OBS status
func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.obsClient.GetStatus()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// broadcastStatus sends status update to all WebSocket clients
func (h *Handler) broadcastStatus() {
	status, err := h.obsClient.GetStatus()
	if err != nil {
		h.logger.Errorf("Failed to get status for broadcast: %v", err)
		return
	}

	data, err := json.Marshal(map[string]interface{}{
		"type": "status_update",
		"data": status,
	})
	if err != nil {
		h.logger.Errorf("Failed to marshal status: %v", err)
		return
	}

	h.hub.broadcast <- data
}
