package models

// OBSConfig stores OBS WebSocket connection settings
type OBSConfig struct {
	URL      string `json:"url"`
	Password string `json:"password"`
}
