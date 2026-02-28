// Package sdk provides the types and helpers needed to build Robo-Stream
// controller plugins. Plugin authors import this package, implement the
// PluginImplementation interface, and call RunPlugin() from main().
package sdk

// PluginInfo contains static metadata about the plugin.
type PluginInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Icon        string `json:"icon,omitempty"`
	Author      string `json:"author,omitempty"`
}

// PluginStatus describes the plugin's current connection state.
type PluginStatus struct {
	Connected bool                   `json:"connected"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// ActionTypeDef describes one action that the plugin supports.
// The host uses these definitions to build configuration UIs.
type ActionTypeDef struct {
	Type        string     `json:"type"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Params      []ParamDef `json:"params,omitempty"`
}

// ParamDef describes one parameter of an action.
type ParamDef struct {
	Key      string      `json:"key"`
	Label    string      `json:"label"`
	Type     string      `json:"type"`    // "string" | "number" | "boolean" | "select"
	Options  []string    `json:"options,omitempty"` // for "select" type
	Required bool        `json:"required,omitempty"`
	Default  interface{} `json:"default,omitempty"`
}

// ConfigField describes one setting that the plugin requires.
// The host displays these fields in its settings UI and persists the values.
type ConfigField struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Type     string `json:"type"`    // "string" | "password" | "number" | "boolean"
	Required bool   `json:"required,omitempty"`
	Default  string `json:"default,omitempty"`
	Help     string `json:"help,omitempty"`
}

// ExecuteRequest is the body sent to POST /execute.
// Its fields mirror models.ButtonAction in the host.
type ExecuteRequest struct {
	Controller string                 `json:"controller"`
	Type       string                 `json:"type"`
	Params     map[string]interface{} `json:"params,omitempty"`
}

// Button represents a default button that the plugin provides.
// The host creates these buttons in the library when the plugin is first installed.
type Button struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Icon        string         `json:"icon,omitempty"`
	Color       string         `json:"color,omitempty"`
	Action      ExecuteRequest `json:"action"`
}

// InitializeRequest is sent to POST /initialize on plugin startup.
type InitializeRequest struct {
	Config map[string]interface{} `json:"config"`
}

// ErrorResponse is returned by any endpoint on error.
type ErrorResponse struct {
	Error string `json:"error"`
}
