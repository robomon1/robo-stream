package models

import "time"

// Button represents a reusable button in the library
type Button struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Icon        string       `json:"icon"`
	Color       string       `json:"color"`
	Action      ButtonAction `json:"action"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// ButtonAction defines what the button does
type ButtonAction struct {
	Type   string                 `json:"type"`
	Params map[string]interface{} `json:"params,omitempty"`
}
