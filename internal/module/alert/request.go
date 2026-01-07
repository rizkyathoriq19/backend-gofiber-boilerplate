package alert

// CreateAlertRequest represents the request to create an alert
type CreateAlertRequest struct {
	RoomID    string        `json:"room_id" validate:"required,uuid"`
	PatientID *string       `json:"patient_id" validate:"omitempty,uuid"`
	Type      AlertType     `json:"type" validate:"required,oneof=voice_call button_press emergency system scheduled"`
	Priority  AlertPriority `json:"priority" validate:"omitempty,oneof=critical high medium low"`
	Message   string        `json:"message"`
}

// DeviceAlertRequest represents an alert from a device (AI voice detection)
type DeviceAlertRequest struct {
	Type             AlertType `json:"type" validate:"required,oneof=voice_call button_press emergency"`
	DetectedKeywords []string  `json:"detected_keywords"`
	AudioReference   string    `json:"audio_reference"`
	Message          string    `json:"message"`
}

// UpdateAlertStatusRequest represents a status update request
type UpdateAlertStatusRequest struct {
	Status AlertStatus `json:"status" validate:"required,oneof=acknowledged in_progress resolved cancelled"`
	Notes  string      `json:"notes"`
}

// ResolveAlertRequest represents the request to resolve an alert
type ResolveAlertRequest struct {
	Notes string `json:"notes"`
}

// CancelAlertRequest represents the request to cancel an alert
type CancelAlertRequest struct {
	Reason string `json:"reason" validate:"required"`
}
