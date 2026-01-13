package alert

import (
	"boilerplate-be/internal/module/device"
	"boilerplate-be/internal/module/patient"
	"boilerplate-be/internal/module/staff"
	"boilerplate-be/internal/shared/errors"
)

type alertUseCase struct {
	alertRepo    AlertRepository
	staffRepo    staff.StaffRepository
	patientRepo  patient.PatientRepository
	deviceRepo   device.DeviceRepository
}

// NewAlertUseCase creates a new alert use case
func NewAlertUseCase(alertRepo AlertRepository, staffRepo staff.StaffRepository, patientRepo patient.PatientRepository, deviceRepo device.DeviceRepository) AlertUseCase {
	return &alertUseCase{
		alertRepo:   alertRepo,
		staffRepo:   staffRepo,
		patientRepo: patientRepo,
		deviceRepo:  deviceRepo,
	}
}

// determinePriority determines alert priority based on patient condition
func (u *alertUseCase) determinePriority(alertType AlertType, patientID *string) AlertPriority {
	// Emergency is always critical
	if alertType == AlertTypeEmergency {
		return PriorityCritical
	}

	// If no patient associated, use medium priority
	if patientID == nil {
		return PriorityMedium
	}

	// Get patient condition level
	p, err := u.patientRepo.GetByID(*patientID)
	if err != nil {
		return PriorityMedium
	}

	// Map patient condition to alert priority
	switch p.ConditionLevel {
	case patient.ConditionCritical:
		return PriorityCritical
	case patient.ConditionSerious:
		return PriorityHigh
	case patient.ConditionModerate:
		return PriorityMedium
	default:
		return PriorityLow
	}
}

// CreateAlert creates a new alert
func (u *alertUseCase) CreateAlert(req *CreateAlertRequest) (*Alert, error) {
	priority := req.Priority
	if priority == "" {
		priority = u.determinePriority(req.Type, req.PatientID)
	}

	alert := &Alert{
		RoomID:    req.RoomID,
		PatientID: req.PatientID,
		Type:      req.Type,
		Priority:  priority,
		Message:   req.Message,
	}

	if err := u.alertRepo.Create(alert); err != nil {
		return nil, err
	}

	// Create history entry
	history := &AlertHistory{
		AlertID:   alert.ID,
		Action:    "created",
		NewStatus: StatusPending,
		Notes:     "Alert created",
	}
	_ = u.alertRepo.CreateHistory(history)

	return alert, nil
}

// CreateAlertFromDevice creates an alert from a device (AI voice detection)
func (u *alertUseCase) CreateAlertFromDevice(deviceID string, req *DeviceAlertRequest) (*Alert, error) {
	// Get device and room
	d, err := u.deviceRepo.GetByID(deviceID)
	if err != nil {
		return nil, err
	}

	if d.RoomID == nil {
		return nil, errors.New(errors.InvalidRequestBody)
	}

	// Get patients in room for priority determination
	patients, _ := u.patientRepo.GetByRoomID(*d.RoomID)
	var patientID *string
	var priority AlertPriority

	if len(patients) > 0 {
		// Use the most critical patient for priority
		mostCritical := patients[0]
		for _, p := range patients {
			if getPriorityWeight(u.determinePriority(req.Type, &p.ID)) < getPriorityWeight(u.determinePriority(req.Type, &mostCritical.ID)) {
				mostCritical = p
			}
		}
		patientID = &mostCritical.ID
		priority = u.determinePriority(req.Type, patientID)
	} else {
		priority = u.determinePriority(req.Type, nil)
	}

	alert := &Alert{
		RoomID:           *d.RoomID,
		PatientID:        patientID,
		DeviceID:         &deviceID,
		Type:             req.Type,
		Priority:         priority,
		Message:          req.Message,
		DetectedKeywords: req.DetectedKeywords,
		AudioReference:   req.AudioReference,
	}

	if err := u.alertRepo.Create(alert); err != nil {
		return nil, err
	}

	// Create history entry
	history := &AlertHistory{
		AlertID:   alert.ID,
		Action:    "created_by_device",
		NewStatus: StatusPending,
		Notes:     "Alert created by device voice detection",
	}
	_ = u.alertRepo.CreateHistory(history)

	return alert, nil
}

func getPriorityWeight(p AlertPriority) int {
	switch p {
	case PriorityCritical:
		return 1
	case PriorityHigh:
		return 2
	case PriorityMedium:
		return 3
	default:
		return 4
	}
}

// GetAlert gets an alert by ID
func (u *alertUseCase) GetAlert(id string) (*Alert, error) {
	return u.alertRepo.GetByID(id)
}

// GetAlerts gets all alerts with filters
func (u *alertUseCase) GetAlerts(filter *AlertFilter) ([]*AlertWithDetails, int, error) {
	return u.alertRepo.GetAll(filter)
}

// GetActiveAlerts gets all active alerts
func (u *alertUseCase) GetActiveAlerts() ([]*AlertWithDetails, error) {
	return u.alertRepo.GetActiveAlerts()
}

// GetActiveAlertsByRoom gets active alerts for a room
func (u *alertUseCase) GetActiveAlertsByRoom(roomID string) ([]*AlertWithDetails, error) {
	return u.alertRepo.GetActiveAlertsByRoom(roomID)
}

// AcknowledgeAlert acknowledges an alert
func (u *alertUseCase) AcknowledgeAlert(id, staffID string) error {
	alert, err := u.alertRepo.GetByID(id)
	if err != nil {
		return err
	}

	if alert.Status != StatusPending && alert.Status != StatusEscalated {
		return errors.New(errors.InvalidRequestBody)
	}

	previousStatus := alert.Status
	if err := u.alertRepo.Acknowledge(id, staffID); err != nil {
		return err
	}

	// Create history entry
	history := &AlertHistory{
		AlertID:        id,
		StaffID:        &staffID,
		Action:         "acknowledged",
		PreviousStatus: previousStatus,
		NewStatus:      StatusAcknowledged,
	}
	_ = u.alertRepo.CreateHistory(history)

	return nil
}

// StartProgress moves alert to in-progress status
func (u *alertUseCase) StartProgress(id, staffID string) error {
	alert, err := u.alertRepo.GetByID(id)
	if err != nil {
		return err
	}

	if alert.Status != StatusAcknowledged {
		return errors.New(errors.InvalidRequestBody)
	}

	previousStatus := alert.Status
	if err := u.alertRepo.UpdateStatus(id, StatusInProgress, &staffID); err != nil {
		return err
	}

	history := &AlertHistory{
		AlertID:        id,
		StaffID:        &staffID,
		Action:         "in_progress",
		PreviousStatus: previousStatus,
		NewStatus:      StatusInProgress,
	}
	_ = u.alertRepo.CreateHistory(history)

	return nil
}

// ResolveAlert resolves an alert
func (u *alertUseCase) ResolveAlert(id, staffID string, notes string) error {
	alert, err := u.alertRepo.GetByID(id)
	if err != nil {
		return err
	}

	if alert.Status == StatusResolved || alert.Status == StatusCancelled {
		return errors.New(errors.InvalidRequestBody)
	}

	previousStatus := alert.Status
	if err := u.alertRepo.Resolve(id, staffID); err != nil {
		return err
	}

	history := &AlertHistory{
		AlertID:        id,
		StaffID:        &staffID,
		Action:         "resolved",
		PreviousStatus: previousStatus,
		NewStatus:      StatusResolved,
		Notes:          notes,
	}
	_ = u.alertRepo.CreateHistory(history)

	return nil
}

// CancelAlert cancels an alert
func (u *alertUseCase) CancelAlert(id, staffID string, reason string) error {
	alert, err := u.alertRepo.GetByID(id)
	if err != nil {
		return err
	}

	if alert.Status == StatusResolved || alert.Status == StatusCancelled {
		return errors.New(errors.InvalidRequestBody)
	}

	previousStatus := alert.Status
	if err := u.alertRepo.UpdateStatus(id, StatusCancelled, &staffID); err != nil {
		return err
	}

	history := &AlertHistory{
		AlertID:        id,
		StaffID:        &staffID,
		Action:         "cancelled",
		PreviousStatus: previousStatus,
		NewStatus:      StatusCancelled,
		Notes:          reason,
	}
	_ = u.alertRepo.CreateHistory(history)

	return nil
}

// EscalateAlert escalates an alert to another nurse
func (u *alertUseCase) EscalateAlert(id string) error {
	alert, err := u.alertRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Get on-duty staff for the room (excluding current assigned)
	onDutyStaff, err := u.staffRepo.GetOnDutyByRoom(alert.RoomID)
	if err != nil {
		return err
	}

	var newStaffID *string
	for _, s := range onDutyStaff {
		if alert.AssignedStaffID == nil || s.ID != *alert.AssignedStaffID {
			newStaffID = &s.ID
			break
		}
	}

	previousStatus := alert.Status
	if err := u.alertRepo.Escalate(id, newStaffID); err != nil {
		return err
	}

	history := &AlertHistory{
		AlertID:        id,
		Action:         "escalated",
		PreviousStatus: previousStatus,
		NewStatus:      StatusEscalated,
		Notes:          "Alert escalated due to timeout",
	}
	_ = u.alertRepo.CreateHistory(history)

	return nil
}

// GetPendingAlertsForEscalation gets alerts that need escalation
func (u *alertUseCase) GetPendingAlertsForEscalation() ([]*Alert, error) {
	return u.alertRepo.GetPendingAlertsForEscalation(5) // 5 minutes default
}

// GetAlertHistory gets the history for an alert
func (u *alertUseCase) GetAlertHistory(alertID string) ([]*AlertHistory, error) {
	return u.alertRepo.GetHistory(alertID)
}

// CanStaffHandleAlert checks if staff is assigned to the alert's room
func (u *alertUseCase) CanStaffHandleAlert(staffID, alertID string) (bool, error) {
	alert, err := u.alertRepo.GetByID(alertID)
	if err != nil {
		return false, err
	}

	return u.staffRepo.IsAssignedToRoom(staffID, alert.RoomID)
}
