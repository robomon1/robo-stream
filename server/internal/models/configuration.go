package models

import "time"

// Configuration represents a button layout for a specific role/client
type Configuration struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Grid        GridConfig        `json:"grid"`
	Buttons     map[string]string `json:"buttons"` // position (btn-0-0) -> button ID
	IsDefault   bool              `json:"is_default"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// GridConfig defines the button grid size
type GridConfig struct {
	Rows int `json:"rows"`
	Cols int `json:"cols"`
}

// ConfigurationSummary is the lightweight version returned by the list endpoint.
// It omits the full Buttons map so clients can display a config picker without
// fetching all button data upfront.
type ConfigurationSummary struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Grid        GridConfig `json:"grid"`
	IsDefault   bool       `json:"is_default"`
	ButtonCount int        `json:"button_count"`
}

// ResolvedConfiguration is what gets sent to clients with full button details
type ResolvedConfiguration struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Grid    GridConfig        `json:"grid"`
	Buttons []ResolvedButton  `json:"buttons"`
}

// ResolvedButton is a button with position information for the client
type ResolvedButton struct {
	ID     string       `json:"id"`
	Row    int          `json:"row"`
	Col    int          `json:"col"`
	Text   string       `json:"text"`
	Icon   string       `json:"icon"`
	Color  string       `json:"color"`
	Action ButtonAction `json:"action"`
}
