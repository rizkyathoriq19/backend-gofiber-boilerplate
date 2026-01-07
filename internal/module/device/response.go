package device

import (
	"time"
)

// DeviceResponse represents the device response
type DeviceResponse struct {
	ID            string                 `json:"id"`
	RoomID        *string                `json:"room_id"`
	Type          DeviceType             `json:"type"`
	SerialNumber  string                 `json:"serial_number"`
	Name          string                 `json:"name"`
	Status        DeviceStatus           `json:"status"`
	Config        map[string]interface{} `json:"config"`
	LastHeartbeat *time.Time             `json:"last_heartbeat"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// DeviceWithAPIKeyResponse includes the API key (only on registration)
type DeviceWithAPIKeyResponse struct {
	DeviceResponse
	APIKey string `json:"api_key"`
}

// DeviceListResponse represents paginated device list response
type DeviceListResponse struct {
	Devices    []*DeviceResponse `json:"devices"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	TotalPages int               `json:"total_pages"`
}

// ToResponse converts Device entity to DeviceResponse
func (d *Device) ToResponse() *DeviceResponse {
	return &DeviceResponse{
		ID:            d.ID,
		RoomID:        d.RoomID,
		Type:          d.Type,
		SerialNumber:  d.SerialNumber,
		Name:          d.Name,
		Status:        d.Status,
		Config:        d.Config,
		LastHeartbeat: d.LastHeartbeat,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}
}
