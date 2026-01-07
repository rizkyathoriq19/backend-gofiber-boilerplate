package patient

import (
	"time"
)

// ConditionLevel represents the patient's condition level
type ConditionLevel string

const (
	ConditionCritical ConditionLevel = "critical"
	ConditionSerious  ConditionLevel = "serious"
	ConditionModerate ConditionLevel = "moderate"
	ConditionStable   ConditionLevel = "stable"
	ConditionGood     ConditionLevel = "good"
)

// Patient represents a patient in the hospital
type Patient struct {
	ID                   string         `json:"id" db:"id"`
	RoomID               *string        `json:"room_id" db:"room_id"`
	MedicalRecordNumber  string         `json:"medical_record_number" db:"medical_record_number"`
	Name                 string         `json:"name" db:"name"`
	DateOfBirth          *time.Time     `json:"date_of_birth" db:"date_of_birth"`
	Gender               string         `json:"gender" db:"gender"`
	ConditionLevel       ConditionLevel `json:"condition_level" db:"condition_level"`
	Diagnosis            string         `json:"diagnosis" db:"diagnosis"`
	Notes                string         `json:"notes" db:"notes"`
	AdmissionDate        time.Time      `json:"admission_date" db:"admission_date"`
	DischargeDate        *time.Time     `json:"discharge_date" db:"discharge_date"`
	CreatedAt            time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at" db:"updated_at"`
}

// PatientWithRoom includes room details
type PatientWithRoom struct {
	Patient
	RoomName string `json:"room_name"`
	RoomType string `json:"room_type"`
}

// PatientFilter represents filters for querying patients
type PatientFilter struct {
	RoomID         string         `query:"room_id"`
	ConditionLevel ConditionLevel `query:"condition_level"`
	IsAdmitted     *bool          `query:"is_admitted"`
	Page           int            `query:"page"`
	Limit          int            `query:"limit"`
}
