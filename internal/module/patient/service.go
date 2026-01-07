package patient

import (
	"boilerplate-be/internal/pkg/errors"
)

type patientUseCase struct {
	patientRepo PatientRepository
}

// NewPatientUseCase creates a new patient use case
func NewPatientUseCase(patientRepo PatientRepository) PatientUseCase {
	return &patientUseCase{
		patientRepo: patientRepo,
	}
}

// AdmitPatient admits a new patient
func (u *patientUseCase) AdmitPatient(req *AdmitPatientRequest) (*Patient, error) {
	// Check if MRN already exists
	existing, err := u.patientRepo.GetByMedicalRecordNumber(req.MedicalRecordNumber)
	if err == nil && existing != nil {
		return nil, errors.New(errors.ResourceAlreadyExists)
	}

	patient := &Patient{
		RoomID:              req.RoomID,
		MedicalRecordNumber: req.MedicalRecordNumber,
		Name:                req.Name,
		DateOfBirth:         req.DateOfBirth,
		Gender:              req.Gender,
		ConditionLevel:      req.ConditionLevel,
		Diagnosis:           req.Diagnosis,
		Notes:               req.Notes,
	}

	if patient.ConditionLevel == "" {
		patient.ConditionLevel = ConditionStable
	}

	if err := u.patientRepo.Create(patient); err != nil {
		return nil, err
	}

	return patient, nil
}

// GetPatient gets a patient by ID
func (u *patientUseCase) GetPatient(id string) (*Patient, error) {
	return u.patientRepo.GetByID(id)
}

// GetPatientsByRoom gets all patients in a room
func (u *patientUseCase) GetPatientsByRoom(roomID string) ([]*Patient, error) {
	return u.patientRepo.GetByRoomID(roomID)
}

// GetAllPatients gets all patients with filters
func (u *patientUseCase) GetAllPatients(filter *PatientFilter) ([]*PatientWithRoom, int, error) {
	return u.patientRepo.GetAll(filter)
}

// UpdatePatient updates a patient
func (u *patientUseCase) UpdatePatient(id string, req *UpdatePatientRequest) (*Patient, error) {
	patient, err := u.patientRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.RoomID != nil {
		patient.RoomID = req.RoomID
	}
	if req.Name != "" {
		patient.Name = req.Name
	}
	if req.ConditionLevel != "" {
		patient.ConditionLevel = req.ConditionLevel
	}
	if req.Diagnosis != "" {
		patient.Diagnosis = req.Diagnosis
	}
	if req.Notes != "" {
		patient.Notes = req.Notes
	}

	if err := u.patientRepo.Update(patient); err != nil {
		return nil, err
	}

	return patient, nil
}

// UpdateConditionLevel updates a patient's condition level
func (u *patientUseCase) UpdateConditionLevel(id string, level ConditionLevel) error {
	_, err := u.patientRepo.GetByID(id)
	if err != nil {
		return err
	}

	return u.patientRepo.UpdateConditionLevel(id, level)
}

// DischargePatient discharges a patient
func (u *patientUseCase) DischargePatient(id string) error {
	_, err := u.patientRepo.GetByID(id)
	if err != nil {
		return err
	}

	return u.patientRepo.Discharge(id)
}

// DeletePatient deletes a patient
func (u *patientUseCase) DeletePatient(id string) error {
	_, err := u.patientRepo.GetByID(id)
	if err != nil {
		return err
	}

	return u.patientRepo.Delete(id)
}
