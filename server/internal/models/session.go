package models

import "time"

// ClientSession tracks which client is using which configuration
type ClientSession struct {
	SessionID     string    `json:"session_id"`
	ClientID      string    `json:"client_id"`
	ClientName    string    `json:"client_name"`
	ConfigID      string    `json:"config_id"`
	IPAddress     string    `json:"ip_address"`
	LastConnected time.Time `json:"last_connected"`
	LastActive    time.Time `json:"last_active"`
}
