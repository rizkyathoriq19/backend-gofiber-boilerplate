package patient

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"boilerplate-be/internal/shared/errors"
	"boilerplate-be/internal/shared/utils"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type patientRepository struct {
	db          *sql.DB
	cacheHelper *utils.CacheHelper
}

// NewPatientRepository creates a new patient repository
func NewPatientRepository(db *sql.DB, cacheHelper *utils.CacheHelper) PatientRepository {
	return &patientRepository{
		db:          db,
		cacheHelper: cacheHelper,
	}
}

// Create creates a new patient
func (r *patientRepository) Create(patient *Patient) error {
	id, _ := uuid.NewV7()
	patient.ID = id.String()
	patient.CreatedAt = time.Now()
	patient.UpdatedAt = time.Now()
	patient.AdmissionDate = time.Now()
	if patient.ConditionLevel == "" {
		patient.ConditionLevel = ConditionStable
	}

	query := `
		INSERT INTO patients (id, room_id, medical_record_number, name, date_of_birth, gender, condition_level, diagnosis, notes, admission_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.Exec(query, patient.ID, patient.RoomID, patient.MedicalRecordNumber, patient.Name, patient.DateOfBirth, patient.Gender, patient.ConditionLevel, patient.Diagnosis, patient.Notes, patient.AdmissionDate, patient.CreatedAt, patient.UpdatedAt)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
			return errors.New(errors.ResourceAlreadyExists)
		}
		return errors.Wrap(err, errors.DatabaseInsertFailed)
	}

	return nil
}

// GetByID gets a patient by ID
func (r *patientRepository) GetByID(id string) (*Patient, error) {
	cacheKey := fmt.Sprintf("patient:%s", id)

	cachedData, err := r.cacheHelper.GetOrSet(context.Background(), cacheKey, func() (interface{}, error) {
		patient := &Patient{}
		query := `
			SELECT id, room_id, medical_record_number, name, date_of_birth, gender, condition_level, diagnosis, notes, admission_date, discharge_date, created_at, updated_at
			FROM patients
			WHERE id = $1
		`
		err := r.db.QueryRow(query, id).Scan(
			&patient.ID, &patient.RoomID, &patient.MedicalRecordNumber, &patient.Name,
			&patient.DateOfBirth, &patient.Gender, &patient.ConditionLevel, &patient.Diagnosis,
			&patient.Notes, &patient.AdmissionDate, &patient.DischargeDate, &patient.CreatedAt, &patient.UpdatedAt,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, errors.New(errors.ResourceNotFound)
			}
			return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
		}
		return patient, nil
	}, 5*time.Minute)

	if err != nil {
		return nil, err
	}

	patient, ok := cachedData.(*Patient)
	if !ok {
		return nil, errors.New(errors.InternalServerError)
	}

	return patient, nil
}

// GetByMedicalRecordNumber gets a patient by MRN
func (r *patientRepository) GetByMedicalRecordNumber(mrn string) (*Patient, error) {
	patient := &Patient{}
	query := `
		SELECT id, room_id, medical_record_number, name, date_of_birth, gender, condition_level, diagnosis, notes, admission_date, discharge_date, created_at, updated_at
		FROM patients
		WHERE medical_record_number = $1
	`
	err := r.db.QueryRow(query, mrn).Scan(
		&patient.ID, &patient.RoomID, &patient.MedicalRecordNumber, &patient.Name,
		&patient.DateOfBirth, &patient.Gender, &patient.ConditionLevel, &patient.Diagnosis,
		&patient.Notes, &patient.AdmissionDate, &patient.DischargeDate, &patient.CreatedAt, &patient.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(errors.ResourceNotFound)
		}
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	return patient, nil
}

// GetByRoomID gets all patients in a room
func (r *patientRepository) GetByRoomID(roomID string) ([]*Patient, error) {
	var patients []*Patient

	query := `
		SELECT id, room_id, medical_record_number, name, date_of_birth, gender, condition_level, diagnosis, notes, admission_date, discharge_date, created_at, updated_at
		FROM patients
		WHERE room_id = $1 AND discharge_date IS NULL
		ORDER BY condition_level, name
	`

	rows, err := r.db.Query(query, roomID)
	if err != nil {
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	defer rows.Close()

	for rows.Next() {
		patient := &Patient{}
		err := rows.Scan(
			&patient.ID, &patient.RoomID, &patient.MedicalRecordNumber, &patient.Name,
			&patient.DateOfBirth, &patient.Gender, &patient.ConditionLevel, &patient.Diagnosis,
			&patient.Notes, &patient.AdmissionDate, &patient.DischargeDate, &patient.CreatedAt, &patient.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
		}
		patients = append(patients, patient)
	}

	return patients, nil
}

// GetAll gets all patients with filters
func (r *patientRepository) GetAll(filter *PatientFilter) ([]*PatientWithRoom, int, error) {
	var patients []*PatientWithRoom
	var total int

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 10
	}

	baseQuery := `FROM patients p LEFT JOIN rooms rm ON p.room_id = rm.id WHERE 1=1`
	var args []interface{}
	argIndex := 1

	if filter.RoomID != "" {
		baseQuery += fmt.Sprintf(" AND p.room_id = $%d", argIndex)
		args = append(args, filter.RoomID)
		argIndex++
	}
	if filter.ConditionLevel != "" {
		baseQuery += fmt.Sprintf(" AND p.condition_level = $%d", argIndex)
		args = append(args, filter.ConditionLevel)
		argIndex++
	}
	if filter.IsAdmitted != nil {
		if *filter.IsAdmitted {
			baseQuery += " AND p.discharge_date IS NULL"
		} else {
			baseQuery += " AND p.discharge_date IS NOT NULL"
		}
	}

	countQuery := `SELECT COUNT(*) ` + baseQuery
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, errors.Wrap(err, errors.DatabaseQueryFailed)
	}

	offset := (filter.Page - 1) * filter.Limit
	selectQuery := fmt.Sprintf(`SELECT p.id, p.room_id, p.medical_record_number, p.name, p.date_of_birth, p.gender, p.condition_level, p.diagnosis, p.notes, p.admission_date, p.discharge_date, p.created_at, p.updated_at, COALESCE(rm.name, ''), COALESCE(rm.type::text, '') %s ORDER BY p.condition_level, p.admission_date DESC LIMIT $%d OFFSET $%d`, baseQuery, argIndex, argIndex+1)
	args = append(args, filter.Limit, offset)

	rows, err := r.db.Query(selectQuery, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	defer rows.Close()

	for rows.Next() {
		patient := &PatientWithRoom{}
		err := rows.Scan(
			&patient.ID, &patient.RoomID, &patient.MedicalRecordNumber, &patient.Name,
			&patient.DateOfBirth, &patient.Gender, &patient.ConditionLevel, &patient.Diagnosis,
			&patient.Notes, &patient.AdmissionDate, &patient.DischargeDate, &patient.CreatedAt, &patient.UpdatedAt,
			&patient.RoomName, &patient.RoomType,
		)
		if err != nil {
			return nil, 0, errors.Wrap(err, errors.DatabaseQueryFailed)
		}
		patients = append(patients, patient)
	}

	return patients, total, nil
}

// Update updates a patient
func (r *patientRepository) Update(patient *Patient) error {
	patient.UpdatedAt = time.Now()

	query := `
		UPDATE patients 
		SET room_id = $2, name = $3, condition_level = $4, diagnosis = $5, notes = $6, updated_at = $7
		WHERE id = $1
	`

	result, err := r.db.Exec(query, patient.ID, patient.RoomID, patient.Name, patient.ConditionLevel, patient.Diagnosis, patient.Notes, patient.UpdatedAt)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseUpdateFailed)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(errors.ResourceNotFound)
	}

	cacheKey := fmt.Sprintf("patient:%s", patient.ID)
	_ = r.cacheHelper.Delete(context.Background(), cacheKey)

	return nil
}

// UpdateConditionLevel updates the condition level
func (r *patientRepository) UpdateConditionLevel(id string, level ConditionLevel) error {
	query := `UPDATE patients SET condition_level = $2, updated_at = $3 WHERE id = $1`

	result, err := r.db.Exec(query, id, level, time.Now())
	if err != nil {
		return errors.Wrap(err, errors.DatabaseUpdateFailed)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(errors.ResourceNotFound)
	}

	cacheKey := fmt.Sprintf("patient:%s", id)
	_ = r.cacheHelper.Delete(context.Background(), cacheKey)

	return nil
}

// Discharge discharges a patient
func (r *patientRepository) Discharge(id string) error {
	query := `UPDATE patients SET discharge_date = $2, updated_at = $3 WHERE id = $1`

	now := time.Now()
	result, err := r.db.Exec(query, id, now, now)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseUpdateFailed)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(errors.ResourceNotFound)
	}

	cacheKey := fmt.Sprintf("patient:%s", id)
	_ = r.cacheHelper.Delete(context.Background(), cacheKey)

	return nil
}

// Delete deletes a patient
func (r *patientRepository) Delete(id string) error {
	query := `DELETE FROM patients WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseDeleteFailed)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(errors.ResourceNotFound)
	}

	cacheKey := fmt.Sprintf("patient:%s", id)
	_ = r.cacheHelper.Delete(context.Background(), cacheKey)

	return nil
}

// ==================== Vital Signs ====================

// CreateVitalSign creates a new vital sign record
func (r *patientRepository) CreateVitalSign(vs *VitalSign) error {
	id, _ := uuid.NewV7()
	vs.ID = id.String()
	vs.RecordedAt = time.Now()
	vs.CreatedAt = time.Now()

	query := `
		INSERT INTO vital_signs (id, patient_id, recorded_by_staff_id, heart_rate, blood_pressure_sys, blood_pressure_dia, temperature, oxygen_saturation, respiratory_rate, pain_level, notes, recorded_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.Exec(query, vs.ID, vs.PatientID, vs.RecordedByStaffID, vs.HeartRate, vs.BloodPressureSys, vs.BloodPressureDia, vs.Temperature, vs.OxygenSaturation, vs.RespiratoryRate, vs.PainLevel, vs.Notes, vs.RecordedAt, vs.CreatedAt)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseInsertFailed)
	}

	return nil
}

// GetVitalSignsByPatientID gets vital signs history for a patient
func (r *patientRepository) GetVitalSignsByPatientID(patientID string, limit int) ([]*VitalSign, error) {
	if limit <= 0 {
		limit = 20
	}

	query := `
		SELECT id, patient_id, recorded_by_staff_id, heart_rate, blood_pressure_sys, blood_pressure_dia, temperature, oxygen_saturation, respiratory_rate, pain_level, notes, recorded_at, created_at
		FROM vital_signs
		WHERE patient_id = $1
		ORDER BY recorded_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(query, patientID, limit)
	if err != nil {
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	defer rows.Close()

	var vitals []*VitalSign
	for rows.Next() {
		vs := &VitalSign{}
		err := rows.Scan(&vs.ID, &vs.PatientID, &vs.RecordedByStaffID, &vs.HeartRate, &vs.BloodPressureSys, &vs.BloodPressureDia, &vs.Temperature, &vs.OxygenSaturation, &vs.RespiratoryRate, &vs.PainLevel, &vs.Notes, &vs.RecordedAt, &vs.CreatedAt)
		if err != nil {
			return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
		}
		vitals = append(vitals, vs)
	}

	return vitals, nil
}

// GetLatestVitalSign gets the most recent vital sign for a patient
func (r *patientRepository) GetLatestVitalSign(patientID string) (*VitalSign, error) {
	vs := &VitalSign{}

	query := `
		SELECT id, patient_id, recorded_by_staff_id, heart_rate, blood_pressure_sys, blood_pressure_dia, temperature, oxygen_saturation, respiratory_rate, pain_level, notes, recorded_at, created_at
		FROM vital_signs
		WHERE patient_id = $1
		ORDER BY recorded_at DESC
		LIMIT 1
	`

	err := r.db.QueryRow(query, patientID).Scan(&vs.ID, &vs.PatientID, &vs.RecordedByStaffID, &vs.HeartRate, &vs.BloodPressureSys, &vs.BloodPressureDia, &vs.Temperature, &vs.OxygenSaturation, &vs.RespiratoryRate, &vs.PainLevel, &vs.Notes, &vs.RecordedAt, &vs.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No vital signs recorded yet
		}
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}

	return vs, nil
}

