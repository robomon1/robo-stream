package config

// GridConfig defines the button grid layout
type GridConfig struct {
	Rows int `json:"rows"`
	Cols int `json:"cols"`
}

// ButtonAction defines what happens when a button is pressed
type ButtonAction struct {
	Type   string                 `json:"type"`
	Params map[string]interface{} `json:"params"`
}

// ResolvedButton represents a button with position from server
type ResolvedButton struct {
	ID     string       `json:"id"`
	Row    int          `json:"row"`
	Col    int          `json:"col"`
	Text   string       `json:"text"`
	Icon   string       `json:"icon"`
	Color  string       `json:"color"`
	Action ButtonAction `json:"action"`
}

// Configuration represents a button configuration (for listing)
// This is what the server returns in the list endpoint
type Configuration struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Grid        GridConfig        `json:"grid"`
	Buttons     map[string]string `json:"buttons"` // position -> button_id
	IsDefault   bool              `json:"is_default"`
}

// ResolvedConfiguration represents a configuration with full button details
// This is what the server returns when getting a specific config
type ResolvedConfiguration struct {
	ID      string           `json:"id"`
	Name    string           `json:"name"`
	Grid    GridConfig       `json:"grid"`
	Buttons []ResolvedButton `json:"buttons"` // array of buttons with positions
}

// GetButtonAt returns the button at the given position
func (c *ResolvedConfiguration) GetButtonAt(row, col int) *ResolvedButton {
	for i := range c.Buttons {
		if c.Buttons[i].Row == row && c.Buttons[i].Col == col {
			return &c.Buttons[i]
		}
	}
	return nil
}
