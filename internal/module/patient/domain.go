package patient

// PatientRepository defines the interface for patient data operations
type PatientRepository interface {
	Create(patient *Patient) error
	GetByID(id string) (*Patient, error)
	GetByMedicalRecordNumber(mrn string) (*Patient, error)
	GetByRoomID(roomID string) ([]*Patient, error)
	GetAll(filter *PatientFilter) ([]*PatientWithRoom, int, error)
	Update(patient *Patient) error
	UpdateConditionLevel(id string, level ConditionLevel) error
	Discharge(id string) error
	Delete(id string) error
}

// PatientUseCase defines the interface for patient business logic
type PatientUseCase interface {
	AdmitPatient(req *AdmitPatientRequest) (*Patient, error)
	GetPatient(id string) (*Patient, error)
	GetPatientsByRoom(roomID string) ([]*Patient, error)
	GetAllPatients(filter *PatientFilter) ([]*PatientWithRoom, int, error)
	UpdatePatient(id string, req *UpdatePatientRequest) (*Patient, error)
	UpdateConditionLevel(id string, level ConditionLevel) error
	DischargePatient(id string) error
	DeletePatient(id string) error
}
