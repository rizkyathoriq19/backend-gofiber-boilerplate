package patient

import "time"

// VitalSign represents vital measurements of a patient
type VitalSign struct {
	ID                string     `json:"id" db:"id"`
	PatientID         string     `json:"patient_id" db:"patient_id"`
	RecordedByStaffID *string    `json:"recorded_by_staff_id,omitempty" db:"recorded_by_staff_id"`
	HeartRate         *int       `json:"heart_rate,omitempty" db:"heart_rate"`
	BloodPressureSys  *int       `json:"blood_pressure_sys,omitempty" db:"blood_pressure_sys"`
	BloodPressureDia  *int       `json:"blood_pressure_dia,omitempty" db:"blood_pressure_dia"`
	Temperature       *float64   `json:"temperature,omitempty" db:"temperature"`
	OxygenSaturation  *int       `json:"oxygen_saturation,omitempty" db:"oxygen_saturation"`
	RespiratoryRate   *int       `json:"respiratory_rate,omitempty" db:"respiratory_rate"`
	PainLevel         *int       `json:"pain_level,omitempty" db:"pain_level"`
	Notes             *string    `json:"notes,omitempty" db:"notes"`
	RecordedAt        time.Time  `json:"recorded_at" db:"recorded_at"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

// CreateVitalSignRequest represents the request to record vital signs
type CreateVitalSignRequest struct {
	HeartRate        *int     `json:"heart_rate" validate:"omitempty,min=30,max=250"`
	BloodPressureSys *int     `json:"blood_pressure_sys" validate:"omitempty,min=50,max=300"`
	BloodPressureDia *int     `json:"blood_pressure_dia" validate:"omitempty,min=30,max=200"`
	Temperature      *float64 `json:"temperature" validate:"omitempty,min=30,max=45"`
	OxygenSaturation *int     `json:"oxygen_saturation" validate:"omitempty,min=0,max=100"`
	RespiratoryRate  *int     `json:"respiratory_rate" validate:"omitempty,min=5,max=60"`
	PainLevel        *int     `json:"pain_level" validate:"omitempty,min=0,max=10"`
	Notes            *string  `json:"notes"`
}
