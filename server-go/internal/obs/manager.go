package obs

import (
	"fmt"
	"sync"
	"time"

	"github.com/andreykaipov/goobs"
	"github.com/sirupsen/logrus"
)

// Manager manages the connection to OBS Studio via WebSocket 5.x
type Manager struct {
	client          *goobs.Client
	config          *Config
	mu              sync.RWMutex
	connected       bool
	logger          *logrus.Logger
	connectCallback func()
	errorCallback   func(error)
}

// Config contains OBS connection configuration
type Config struct {
	Host              string        `yaml:"host" json:"host"`
	Port              int           `yaml:"port" json:"port"`
	Password          string        `yaml:"password" json:"password"`
	AutoConnect       bool          `yaml:"autoConnect" json:"autoConnect"`
	ReconnectInterval time.Duration `yaml:"reconnectInterval" json:"reconnectInterval"`
}

// NewManager creates a new OBS connection manager
func NewManager(config *Config, logger *logrus.Logger) *Manager {
	if logger == nil {
		logger = logrus.New()
	}

	if config.ReconnectInterval == 0 {
		config.ReconnectInterval = 5 * time.Second
	}

	return &Manager{
		config: config,
		logger: logger,
	}
}

// Connect establishes connection to OBS Studio
func (m *Manager) Connect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.connected {
		return fmt.Errorf("already connected to OBS")
	}

	m.logger.Infof("Connecting to OBS at %s:%d", m.config.Host, m.config.Port)

	address := fmt.Sprintf("%s:%d", m.config.Host, m.config.Port)

	var client *goobs.Client
	var err error

	if m.config.Password != "" {
		client, err = goobs.New(address, goobs.WithPassword(m.config.Password))
	} else {
		client, err = goobs.New(address)
	}

	if err != nil {
		m.logger.Errorf("Failed to connect to OBS: %v", err)
		return fmt.Errorf("failed to connect to OBS: %w", err)
	}

	m.client = client
	m.connected = true
	m.logger.Info("Successfully connected to OBS")

	if m.connectCallback != nil {
		go m.connectCallback()
	}

	return nil
}

// Disconnect closes the connection to OBS
func (m *Manager) Disconnect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.connected {
		return nil // Not an error if already disconnected
	}

	if m.client != nil {
		m.client.Disconnect()
	}

	m.connected = false
	m.client = nil
	m.logger.Info("Disconnected from OBS")

	return nil
}

// IsConnected returns whether we're currently connected to OBS
func (m *Manager) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connected
}

// Client returns the underlying goobs client
func (m *Manager) Client() *goobs.Client {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.client
}

// OnConnect sets a callback to be called when connection is established
func (m *Manager) OnConnect(callback func()) {
	m.connectCallback = callback
}

// OnError sets a callback to be called when an error occurs
func (m *Manager) OnError(callback func(error)) {
	m.errorCallback = callback
}

// GetVersion gets the OBS version information
func (m *Manager) GetVersion() (string, error) {
	m.mu.RLock()
	client := m.client
	m.mu.RUnlock()

	if client == nil {
		return "", fmt.Errorf("not connected to OBS")
	}

	version, err := client.General.GetVersion()
	if err != nil {
		return "", fmt.Errorf("failed to get OBS version: %w", err)
	}

	return version.ObsVersion, nil
}
