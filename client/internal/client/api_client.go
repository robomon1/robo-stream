package client

import (
	"bytes"
	"client/internal/config"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

// APIClient handles communication with YOUR actual robo-stream server
type APIClient struct {
	serverURL string
	sessionID string
	clientID  string
	httpClient *http.Client
	logger     *logrus.Logger
}

// loadClientID loads or creates a persistent client ID
func loadClientID(configDir string) string {
	idFile := filepath.Join(configDir, "client_id.txt")
	
	// Try to load existing ID
	data, err := os.ReadFile(idFile)
	if err == nil && len(data) > 0 {
		return string(data)
	}
	
	// Generate new ID
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}
	clientID := fmt.Sprintf("client-%s-%d", hostname, time.Now().Unix())
	
	// Save for next time
	os.WriteFile(idFile, []byte(clientID), 0644)
	
	return clientID
}

// NewAPIClient creates a new API client
func NewAPIClient(serverURL string, logger *logrus.Logger, configDir string) *APIClient {
	return &APIClient{
		serverURL: serverURL,
		clientID:  loadClientID(configDir),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// RegisterResponse from server
type RegisterResponse struct {
	SessionID string                        `json:"session_id"`
	ConfigID  string                        `json:"config_id"`
	Config    config.ResolvedConfiguration  `json:"config"`
}

// GetServerInfo gets server information
func (c *APIClient) GetServerInfo() (map[string]interface{}, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/api/health", c.serverURL))
	if err != nil {
		return nil, fmt.Errorf("failed to get server info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var info map[string]interface{}
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("failed to parse server info: %w", err)
	}

	return info, nil
}

// Register registers this client with the server and gets a session
func (c *APIClient) Register() (*config.ResolvedConfiguration, error) {
	reqBody := map[string]string{
		"client_id":   c.clientID,
		"client_name": "Wails Desktop Client",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		fmt.Sprintf("%s/api/client/register", c.serverURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var regResp RegisterResponse
	if err := json.Unmarshal(body, &regResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Store session ID for future requests
	c.sessionID = regResp.SessionID

	c.logger.Infof("Registered with server - Session: %s, Config: %s", 
		regResp.SessionID, regResp.ConfigID)

	return &regResp.Config, nil
}

// GetConfigurations gets all available configurations
func (c *APIClient) GetConfigurations() ([]config.Configuration, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/api/configurations", c.serverURL))
	if err != nil {
		return nil, fmt.Errorf("failed to get configurations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var configs []config.Configuration
	if err := json.Unmarshal(body, &configs); err != nil {
		return nil, fmt.Errorf("failed to parse configurations: %w", err)
	}

	return configs, nil
}

// GetConfiguration gets a specific configuration
func (c *APIClient) GetConfiguration(configID string) (*config.ResolvedConfiguration, error) {
	// First switch to this config in our session
	if c.sessionID != "" {
		req, err := http.NewRequest("PUT",
			fmt.Sprintf("%s/api/client/config/%s", c.serverURL, configID),
			nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("X-Session-ID", c.sessionID)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to switch config: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		var resolved config.ResolvedConfiguration
		if err := json.Unmarshal(body, &resolved); err != nil {
			return nil, fmt.Errorf("failed to parse configuration: %w", err)
		}

		return &resolved, nil
	}

	// No session yet, just fetch it
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/api/configurations/%s", c.serverURL, configID))
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var resolved config.ResolvedConfiguration
	if err := json.Unmarshal(body, &resolved); err != nil {
		return nil, fmt.Errorf("failed to parse configuration: %w", err)
	}

	return &resolved, nil
}

// GetDefaultConfiguration gets the default configuration
func (c *APIClient) GetDefaultConfiguration() (*config.ResolvedConfiguration, error) {
	// If we don't have a session, register first
	if c.sessionID == "" {
		return c.Register()
	}

	// Otherwise fetch default
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/api/configurations/default", c.serverURL))
	if err != nil {
		return nil, fmt.Errorf("failed to get default configuration: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var resolved config.ResolvedConfiguration
	if err := json.Unmarshal(body, &resolved); err != nil {
		return nil, fmt.Errorf("failed to parse configuration: %w", err)
	}

	return &resolved, nil
}

// ExecuteAction sends an action request to the server with session header
func (c *APIClient) ExecuteAction(action config.ButtonAction) error {
	if c.sessionID == "" {
		return fmt.Errorf("not registered - no session ID")
	}

	jsonData, err := json.Marshal(action)
	if err != nil {
		return fmt.Errorf("failed to marshal action: %w", err)
	}

	c.logger.Debugf("Executing action: %s with params: %v", action.Type, action.Params)

	req, err := http.NewRequest("POST",
		fmt.Sprintf("%s/api/action", c.serverURL),
		bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Session-ID", c.sessionID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetOBSStatus gets the current OBS status
func (c *APIClient) GetOBSStatus() (map[string]interface{}, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/api/obs/status", c.serverURL))
	if err != nil {
		return nil, fmt.Errorf("failed to get OBS status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

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
