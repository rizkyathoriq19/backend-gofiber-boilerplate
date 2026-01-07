package device

import (
	"time"
)

// DeviceType represents the type of device
type DeviceType string

const (
	DeviceTypeMicrophone   DeviceType = "microphone"
	DeviceTypeTeleprompter DeviceType = "teleprompter"
	DeviceTypeButton       DeviceType = "button"
	DeviceTypeSensor       DeviceType = "sensor"
)

// DeviceStatus represents the status of a device
type DeviceStatus string

const (
	DeviceStatusOnline      DeviceStatus = "online"
	DeviceStatusOffline     DeviceStatus = "offline"
	DeviceStatusMaintenance DeviceStatus = "maintenance"
	DeviceStatusError       DeviceStatus = "error"
)

// Device represents a device in the hospital
type Device struct {
	ID            string            `json:"id" db:"id"`
	RoomID        *string           `json:"room_id" db:"room_id"`
	Type          DeviceType        `json:"type" db:"type"`
	SerialNumber  string            `json:"serial_number" db:"serial_number"`
	Name          string            `json:"name" db:"name"`
	Status        DeviceStatus      `json:"status" db:"status"`
	APIKeyHash    string            `json:"-" db:"api_key_hash"`
	Config        map[string]interface{} `json:"config" db:"config"`
	LastHeartbeat *time.Time        `json:"last_heartbeat" db:"last_heartbeat"`
	CreatedAt     time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at" db:"updated_at"`
}

// DeviceFilter represents filters for querying devices
type DeviceFilter struct {
	RoomID string       `query:"room_id"`
	Type   DeviceType   `query:"type"`
	Status DeviceStatus `query:"status"`
	Page   int          `query:"page"`
	Limit  int          `query:"limit"`
}
