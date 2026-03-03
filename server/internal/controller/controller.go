// Package controller defines the interface that all Robo-Stream controllers
// must implement, plus the registry that routes button actions to the correct
// controller at runtime.
//
// Built-in controllers (e.g. OBS) live in sub-packages like controller/obs.
// External plugin controllers are wrapped by the plugin package, which makes
// a subprocess binary appear as a Controller via HTTP.
package controller

import "github.com/robomon1/robo-stream/server/internal/models"

// ActionTypeDefinition describes one action a controller supports.
// The host uses these definitions to populate button editor UIs.
type ActionTypeDefinition struct {
	Type        string     `json:"type"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Params      []ParamDef `json:"params,omitempty"`

	// Indicator rules — see sdk.ActionTypeDef for full documentation.
	IndicatorField  string `json:"indicator_field,omitempty"`
	IndicatorInvert bool   `json:"indicator_invert,omitempty"`
}

// ParamDef describes one parameter of an action type.
type ParamDef struct {
	Key      string      `json:"key"`
	Label    string      `json:"label"`
	Type     string      `json:"type"`    // "string" | "number" | "boolean" | "select"
	Options  []string    `json:"options,omitempty"`
	Required bool        `json:"required,omitempty"`
	Default  interface{} `json:"default,omitempty"`
}

// ConfigField describes one configuration setting for a controller.
// Built-in controllers (OBS) use these for their settings panels.
// Plugin controllers return these from /config/schema.
type ConfigField struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Type     string `json:"type"`    // "string" | "password" | "number" | "boolean"
	Required bool   `json:"required,omitempty"`
	Default  string `json:"default,omitempty"`
	Help     string `json:"help,omitempty"`
}

// Controller is the interface all controllers must implement, whether they are
// compiled into the server (like OBS) or run as plugin subprocesses.
type Controller interface {
	// Identity
	ID() string
	Name() string
	Description() string
	Version() string

	// Shutdown gracefully stops the controller and releases resources.
	Shutdown() error

	// ExecuteAction runs a button action.
	ExecuteAction(action models.ButtonAction) error

	// SupportedActionTypes returns the list of action types this controller
	// can execute. Used to populate the button editor.
	SupportedActionTypes() []ActionTypeDefinition

	// IsConnected reports whether the controller is currently connected to
	// its target service.
	IsConnected() bool

	// GetStatus returns a free-form map of status information.
	// Must always return without error; return an empty or disconnected map
	// if the underlying service is not reachable.
	GetStatus() map[string]interface{}

	// ComputeIndicator returns the CSS indicator class for a single button
	// action, using the controller's last-known state. Returns "active", "warn",
	// or "" (no indicator). Must be fast — do not block on network I/O for
	// the simple boolean cases; use cached state from the most recent
	// GetStatus() call instead.
	ComputeIndicator(action models.ButtonAction) string

	// GetConfigSchema returns the fields the host should display in the
	// controller's settings panel.
	GetConfigSchema() []ConfigField

	// GetCurrentConfig returns the current persisted configuration.
	GetCurrentConfig() map[string]interface{}

	// SaveConfig persists new configuration values. Implementations should
	// apply the new config (e.g. reconnect) after saving.
	SaveConfig(cfg map[string]interface{}) error

	// GetDefaultButtons returns buttons to add to the library when the
	// controller is first used. Only called during initial setup.
	GetDefaultButtons() []*models.Button
}
