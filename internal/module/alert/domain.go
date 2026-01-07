package alert

// AlertRepository defines the interface for alert data operations
type AlertRepository interface {
	Create(alert *Alert) error
	GetByID(id string) (*Alert, error)
	GetAll(filter *AlertFilter) ([]*AlertWithDetails, int, error)
	GetActiveAlerts() ([]*AlertWithDetails, error)
	GetActiveAlertsByRoom(roomID string) ([]*AlertWithDetails, error)
	GetPendingAlerts() ([]*Alert, error)
	GetPendingAlertsForEscalation(timeoutMinutes int) ([]*Alert, error)
	Update(alert *Alert) error
	UpdateStatus(id string, status AlertStatus, staffID *string) error
	Acknowledge(id, staffID string) error
	Resolve(id, staffID string) error
	Escalate(id string, newStaffID *string) error
	Delete(id string) error

	// History
	CreateHistory(history *AlertHistory) error
	GetHistory(alertID string) ([]*AlertHistory, error)
}

// AlertUseCase defines the interface for alert business logic
type AlertUseCase interface {
	// Alert CRUD
	CreateAlert(req *CreateAlertRequest) (*Alert, error)
	CreateAlertFromDevice(deviceID string, req *DeviceAlertRequest) (*Alert, error)
	GetAlert(id string) (*Alert, error)
	GetAlerts(filter *AlertFilter) ([]*AlertWithDetails, int, error)
	GetActiveAlerts() ([]*AlertWithDetails, error)
	GetActiveAlertsByRoom(roomID string) ([]*AlertWithDetails, error)

	// Alert actions
	AcknowledgeAlert(id, staffID string) error
	StartProgress(id, staffID string) error
	ResolveAlert(id, staffID string, notes string) error
	CancelAlert(id, staffID string, reason string) error
	
	// Escalation
	EscalateAlert(id string) error
	GetPendingAlertsForEscalation() ([]*Alert, error)

	// History
	GetAlertHistory(alertID string) ([]*AlertHistory, error)

	// Permission check
	CanStaffHandleAlert(staffID, alertID string) (bool, error)
}
