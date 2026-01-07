package patient

import "time"

// AdmitPatientRequest represents the request to admit a patient
type AdmitPatientRequest struct {
	RoomID              *string        `json:"room_id" validate:"omitempty,uuid"`
	MedicalRecordNumber string         `json:"medical_record_number" validate:"required,min=1,max=50"`
	Name                string         `json:"name" validate:"required,min=1,max=255"`
	DateOfBirth         *time.Time     `json:"date_of_birth"`
	Gender              string         `json:"gender" validate:"omitempty,oneof=male female other"`
	ConditionLevel      ConditionLevel `json:"condition_level" validate:"omitempty,oneof=critical serious moderate stable good"`
	Diagnosis           string         `json:"diagnosis"`
	Notes               string         `json:"notes"`
}

// UpdatePatientRequest represents the request to update a patient
type UpdatePatientRequest struct {
	RoomID         *string        `json:"room_id" validate:"omitempty,uuid"`
	Name           string         `json:"name" validate:"omitempty,min=1,max=255"`
	ConditionLevel ConditionLevel `json:"condition_level" validate:"omitempty,oneof=critical serious moderate stable good"`
	Diagnosis      string         `json:"diagnosis"`
	Notes          string         `json:"notes"`
}

// UpdateConditionRequest represents the request to update condition level
type UpdateConditionRequest struct {
	ConditionLevel ConditionLevel `json:"condition_level" validate:"required,oneof=critical serious moderate stable good"`
}
