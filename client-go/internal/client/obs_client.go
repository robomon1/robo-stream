package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// OBSClient handles communication with the server-go instance
type OBSClient struct {
	serverURL string
	httpClient *http.Client
	logger    *logrus.Logger
}

// NewOBSClient creates a new OBS client
func NewOBSClient(serverURL string, logger *logrus.Logger) *OBSClient {
	return &OBSClient{
		serverURL: serverURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// ActionRequest represents a request to perform an OBS action
type ActionRequest struct {
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params,omitempty"`
}

// ActionResponse represents a response from an OBS action
type ActionResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ExecuteAction sends an action request to the server
func (c *OBSClient) ExecuteAction(action string, params map[string]interface{}) (*ActionResponse, error) {
	req := ActionRequest{
		Action: action,
		Params: params,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	c.logger.Debugf("Sending action: %s with params: %v", action, params)

	resp, err := c.httpClient.Post(
		fmt.Sprintf("%s/api/action", c.serverURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var actionResp ActionResponse
	if err := json.Unmarshal(body, &actionResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &actionResp, nil
}

// GetStatus gets the current OBS status
func (c *OBSClient) GetStatus() (map[string]interface{}, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/api/status", c.serverURL))
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var status map[string]interface{}
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("failed to parse status: %w", err)
	}

	return status, nil
}

// GetScenes gets the list of available scenes
func (c *OBSClient) GetScenes() ([]string, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/api/scenes", c.serverURL))
	if err != nil {
		return nil, fmt.Errorf("failed to get scenes: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Scenes []string `json:"scenes"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse scenes: %w", err)
	}

	return result.Scenes, nil
}

// GetInputs gets the list of available audio inputs
func (c *OBSClient) GetInputs() ([]string, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/api/inputs", c.serverURL))
	if err != nil {
		return nil, fmt.Errorf("failed to get inputs: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Inputs []string `json:"inputs"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse inputs: %w", err)
	}

	return result.Inputs, nil
}
