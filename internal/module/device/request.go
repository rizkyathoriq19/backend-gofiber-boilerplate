package device

// RegisterDeviceRequest represents the request to register a device
type RegisterDeviceRequest struct {
	RoomID       *string                `json:"room_id" validate:"omitempty,uuid"`
	Type         DeviceType             `json:"type" validate:"required,oneof=microphone teleprompter button sensor"`
	SerialNumber string                 `json:"serial_number" validate:"required,min=1,max=100"`
	Name         string                 `json:"name" validate:"omitempty,max=100"`
	Config       map[string]interface{} `json:"config"`
}

// UpdateDeviceRequest represents the request to update a device
type UpdateDeviceRequest struct {
	RoomID *string                `json:"room_id" validate:"omitempty,uuid"`
	Name   string                 `json:"name" validate:"omitempty,max=100"`
	Config map[string]interface{} `json:"config"`
}

// UpdateDeviceStatusRequest represents the request to update device status
type UpdateDeviceStatusRequest struct {
	Status DeviceStatus `json:"status" validate:"required,oneof=online offline maintenance error"`
}
