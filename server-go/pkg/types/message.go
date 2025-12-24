package types

import "encoding/json"

// MessageType represents the type of message being sent
type MessageType string

const (
	// Client-Server Messages
	MsgTypeConnect       MessageType = "CONNECT"
	MsgTypeDisconnect    MessageType = "DISCONNECT"
	MsgTypeHeartbeat     MessageType = "HEARTBEAT"
	MsgTypeActionTrigger MessageType = "ACTION_TRIGGER"
	
	// Profile Messages
	MsgTypeProfileLoad   MessageType = "PROFILE_LOAD"
	MsgTypeProfileUpdate MessageType = "PROFILE_UPDATE"
	MsgTypeProfileSwitch MessageType = "PROFILE_SWITCH"
	
	// Configuration Messages
	MsgTypeConfigGet    MessageType = "CONFIG_GET"
	MsgTypeConfigUpdate MessageType = "CONFIG_UPDATE"
	
	// Theme Messages
	MsgTypeThemeUpdate MessageType = "THEME_UPDATE"
	
	// Response Messages
	MsgTypeSuccess MessageType = "SUCCESS"
	MsgTypeError   MessageType = "ERROR"
)

// Message represents a message exchanged between server and client
type Message struct {
	Type    MessageType     `json:"type"`
	Header  string          `json:"header,omitempty"`
	Body    string          `json:"body,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// ConnectPayload contains information when a client connects
type ConnectPayload struct {
	ClientID      string `json:"clientId"`
	ClientName    string `json:"clientName"`
	ClientVersion string `json:"clientVersion"`
	Platform      string `json:"platform"`
	ProfileID     string `json:"profileId,omitempty"`
}

// ActionTriggerPayload contains information when an action is triggered
type ActionTriggerPayload struct {
	ActionID   string            `json:"actionId"`
	ProfileID  string            `json:"profileId"`
	Properties map[string]string `json:"properties,omitempty"`
}

// ProfilePayload contains profile information
type ProfilePayload struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Rows        int                    `json:"rows"`
	Cols        int                    `json:"cols"`
	Actions     []Action               `json:"actions"`
	Orientation string                 `json:"orientation"` // portrait or landscape
	Properties  map[string]interface{} `json:"properties,omitempty"`
}

// Action represents a button action on the stream deck
type Action struct {
	ID                   string                 `json:"id"`
	Type                 ActionType             `json:"type"`
	Name                 string                 `json:"name"`
	Row                  int                    `json:"row"`
	Col                  int                    `json:"col"`
	RowSpan              int                    `json:"rowSpan"`
	ColSpan              int                    `json:"colSpan"`
	DisplayText          string                 `json:"displayText,omitempty"`
	DisplayTextAlignment string                 `json:"displayTextAlignment,omitempty"`
	IconPath             string                 `json:"iconPath,omitempty"`
	BackgroundColor      string                 `json:"backgroundColor,omitempty"`
	Properties           map[string]interface{} `json:"properties,omitempty"`
	PluginID             string                 `json:"pluginId,omitempty"`
	Version              string                 `json:"version,omitempty"`
}

// ActionType represents the type of action
type ActionType string

const (
	ActionTypeNormal  ActionType = "NORMAL"
	ActionTypeFolder  ActionType = "FOLDER"
	ActionTypeCombine ActionType = "COMBINE"
	ActionTypeGauge   ActionType = "GAUGE"
)

// ErrorPayload contains error information
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// SuccessPayload contains success response information
type SuccessPayload struct {
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
}
