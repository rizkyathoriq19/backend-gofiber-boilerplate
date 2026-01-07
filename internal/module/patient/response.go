package patient

import (
	"time"
)

// PatientResponse represents the patient response
type PatientResponse struct {
	ID                  string         `json:"id"`
	RoomID              *string        `json:"room_id"`
	MedicalRecordNumber string         `json:"medical_record_number"`
	Name                string         `json:"name"`
	DateOfBirth         *time.Time     `json:"date_of_birth"`
	Gender              string         `json:"gender"`
	ConditionLevel      ConditionLevel `json:"condition_level"`
	Diagnosis           string         `json:"diagnosis"`
	Notes               string         `json:"notes"`
	AdmissionDate       time.Time      `json:"admission_date"`
	DischargeDate       *time.Time     `json:"discharge_date"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
}

// PatientWithRoomResponse includes room details
type PatientWithRoomResponse struct {
	PatientResponse
	RoomName string `json:"room_name,omitempty"`
	RoomType string `json:"room_type,omitempty"`
}

// PatientListResponse represents paginated patient list
type PatientListResponse struct {
	Patients   []*PatientWithRoomResponse `json:"patients"`
	Total      int                        `json:"total"`
	Page       int                        `json:"page"`
	Limit      int                        `json:"limit"`
	TotalPages int                        `json:"total_pages"`
}

// ToResponse converts Patient entity to PatientResponse
func (p *Patient) ToResponse() *PatientResponse {
	return &PatientResponse{
		ID:                  p.ID,
		RoomID:              p.RoomID,
		MedicalRecordNumber: p.MedicalRecordNumber,
		Name:                p.Name,
		DateOfBirth:         p.DateOfBirth,
		Gender:              p.Gender,
		ConditionLevel:      p.ConditionLevel,
		Diagnosis:           p.Diagnosis,
		Notes:               p.Notes,
		AdmissionDate:       p.AdmissionDate,
		DischargeDate:       p.DischargeDate,
		CreatedAt:           p.CreatedAt,
		UpdatedAt:           p.UpdatedAt,
	}
}

// ToResponse converts PatientWithRoom to PatientWithRoomResponse
func (p *PatientWithRoom) ToResponse() *PatientWithRoomResponse {
	return &PatientWithRoomResponse{
		PatientResponse: *p.Patient.ToResponse(),
		RoomName:        p.RoomName,
		RoomType:        p.RoomType,
	}
}
