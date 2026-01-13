package staff

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

type staffRepository struct {
	db          *sql.DB
	cacheHelper *utils.CacheHelper
}

// NewStaffRepository creates a new staff repository
func NewStaffRepository(db *sql.DB, cacheHelper *utils.CacheHelper) StaffRepository {
	return &staffRepository{
		db:          db,
		cacheHelper: cacheHelper,
	}
}

// Create creates a new staff member
func (r *staffRepository) Create(staff *Staff) error {
	id, _ := uuid.NewV7()
	staff.ID = id.String()
	staff.CreatedAt = time.Now()
	staff.UpdatedAt = time.Now()
	if staff.Shift == "" {
		staff.Shift = ShiftMorning
	}

	query := `
		INSERT INTO staff (id, user_id, employee_id, type, department, shift, on_duty, phone, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.Exec(query, staff.ID, staff.UserID, staff.EmployeeID, staff.Type, staff.Department, staff.Shift, staff.OnDuty, staff.Phone, staff.CreatedAt, staff.UpdatedAt)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
			return errors.New(errors.ResourceAlreadyExists)
		}
		return errors.Wrap(err, errors.DatabaseInsertFailed)
	}

	return nil
}

// GetByID gets a staff member by ID
func (r *staffRepository) GetByID(id string) (*Staff, error) {
	cacheKey := fmt.Sprintf("staff:%s", id)

	cachedData, err := r.cacheHelper.GetOrSet(context.Background(), cacheKey, func() (interface{}, error) {
		staff := &Staff{}
		query := `
			SELECT id, user_id, employee_id, type, department, shift, on_duty, phone, created_at, updated_at
			FROM staff
			WHERE id = $1
		`
		err := r.db.QueryRow(query, id).Scan(
			&staff.ID, &staff.UserID, &staff.EmployeeID, &staff.Type, &staff.Department,
			&staff.Shift, &staff.OnDuty, &staff.Phone, &staff.CreatedAt, &staff.UpdatedAt,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, errors.New(errors.ResourceNotFound)
			}
			return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
		}
		return staff, nil
	}, 5*time.Minute)

	if err != nil {
		return nil, err
	}

	staff, ok := cachedData.(*Staff)
	if !ok {
		return nil, errors.New(errors.InternalServerError)
	}

	return staff, nil
}

// GetByUserID gets a staff member by user ID
func (r *staffRepository) GetByUserID(userID string) (*Staff, error) {
	staff := &Staff{}
	query := `
		SELECT id, user_id, employee_id, type, department, shift, on_duty, phone, created_at, updated_at
		FROM staff
		WHERE user_id = $1
	`
	err := r.db.QueryRow(query, userID).Scan(
		&staff.ID, &staff.UserID, &staff.EmployeeID, &staff.Type, &staff.Department,
		&staff.Shift, &staff.OnDuty, &staff.Phone, &staff.CreatedAt, &staff.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(errors.ResourceNotFound)
		}
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	return staff, nil
}

// GetByEmployeeID gets a staff member by employee ID
func (r *staffRepository) GetByEmployeeID(employeeID string) (*Staff, error) {
	staff := &Staff{}
	query := `
		SELECT id, user_id, employee_id, type, department, shift, on_duty, phone, created_at, updated_at
		FROM staff
		WHERE employee_id = $1
	`
	err := r.db.QueryRow(query, employeeID).Scan(
		&staff.ID, &staff.UserID, &staff.EmployeeID, &staff.Type, &staff.Department,
		&staff.Shift, &staff.OnDuty, &staff.Phone, &staff.CreatedAt, &staff.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(errors.ResourceNotFound)
		}
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	return staff, nil
}

// GetAll gets all staff with filters
func (r *staffRepository) GetAll(filter *StaffFilter) ([]*StaffWithUser, int, error) {
	var staffList []*StaffWithUser
	var total int

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 10
	}

	baseQuery := `FROM staff s JOIN users u ON s.user_id = u.id WHERE 1=1`
	var args []interface{}
	argIndex := 1

	if filter.Type != "" {
		baseQuery += fmt.Sprintf(" AND s.type = $%d", argIndex)
		args = append(args, filter.Type)
		argIndex++
	}
	if filter.Department != "" {
		baseQuery += fmt.Sprintf(" AND s.department = $%d", argIndex)
		args = append(args, filter.Department)
		argIndex++
	}
	if filter.Shift != "" {
		baseQuery += fmt.Sprintf(" AND s.shift = $%d", argIndex)
		args = append(args, filter.Shift)
		argIndex++
	}
	if filter.OnDuty != nil {
		baseQuery += fmt.Sprintf(" AND s.on_duty = $%d", argIndex)
		args = append(args, *filter.OnDuty)
		argIndex++
	}

	countQuery := `SELECT COUNT(*) ` + baseQuery
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, errors.Wrap(err, errors.DatabaseQueryFailed)
	}

	offset := (filter.Page - 1) * filter.Limit
	selectQuery := fmt.Sprintf(`SELECT s.id, s.user_id, s.employee_id, s.type, s.department, s.shift, s.on_duty, s.phone, s.created_at, s.updated_at, u.name, u.email %s ORDER BY s.created_at DESC LIMIT $%d OFFSET $%d`, baseQuery, argIndex, argIndex+1)
	args = append(args, filter.Limit, offset)

	rows, err := r.db.Query(selectQuery, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	defer rows.Close()

	for rows.Next() {
		staff := &StaffWithUser{}
		err := rows.Scan(
			&staff.ID, &staff.UserID, &staff.EmployeeID, &staff.Type, &staff.Department,
			&staff.Shift, &staff.OnDuty, &staff.Phone, &staff.CreatedAt, &staff.UpdatedAt,
			&staff.UserName, &staff.UserEmail,
		)
		if err != nil {
			return nil, 0, errors.Wrap(err, errors.DatabaseQueryFailed)
		}
		staffList = append(staffList, staff)
	}

	return staffList, total, nil
}

// GetOnDutyByType gets on-duty staff by type
func (r *staffRepository) GetOnDutyByType(staffType StaffType) ([]*StaffWithUser, error) {
	var staffList []*StaffWithUser

	query := `
		SELECT s.id, s.user_id, s.employee_id, s.type, s.department, s.shift, s.on_duty, s.phone, s.created_at, s.updated_at, u.name, u.email
		FROM staff s
		JOIN users u ON s.user_id = u.id
		WHERE s.on_duty = true AND s.type = $1
		ORDER BY s.shift, u.name
	`

	rows, err := r.db.Query(query, staffType)
	if err != nil {
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	defer rows.Close()

	for rows.Next() {
		staff := &StaffWithUser{}
		err := rows.Scan(
			&staff.ID, &staff.UserID, &staff.EmployeeID, &staff.Type, &staff.Department,
			&staff.Shift, &staff.OnDuty, &staff.Phone, &staff.CreatedAt, &staff.UpdatedAt,
			&staff.UserName, &staff.UserEmail,
		)
		if err != nil {
			return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
		}
		staffList = append(staffList, staff)
	}

	return staffList, nil
}

// GetOnDutyByRoom gets on-duty staff assigned to a specific room
func (r *staffRepository) GetOnDutyByRoom(roomID string) ([]*StaffWithUser, error) {
	var staffList []*StaffWithUser

	query := `
		SELECT s.id, s.user_id, s.employee_id, s.type, s.department, s.shift, s.on_duty, s.phone, s.created_at, s.updated_at, u.name, u.email
		FROM staff s
		JOIN users u ON s.user_id = u.id
		JOIN staff_room_assignments sra ON s.id = sra.staff_id
		WHERE s.on_duty = true AND sra.room_id = $1
		ORDER BY sra.is_primary DESC, s.type, u.name
	`

	rows, err := r.db.Query(query, roomID)
	if err != nil {
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	defer rows.Close()

	for rows.Next() {
		staff := &StaffWithUser{}
		err := rows.Scan(
			&staff.ID, &staff.UserID, &staff.EmployeeID, &staff.Type, &staff.Department,
			&staff.Shift, &staff.OnDuty, &staff.Phone, &staff.CreatedAt, &staff.UpdatedAt,
			&staff.UserName, &staff.UserEmail,
		)
		if err != nil {
			return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
		}
		staffList = append(staffList, staff)
	}

	return staffList, nil
}

// Update updates a staff member
func (r *staffRepository) Update(staff *Staff) error {
	staff.UpdatedAt = time.Now()

	query := `
		UPDATE staff 
		SET department = $2, shift = $3, phone = $4, updated_at = $5
		WHERE id = $1
	`

	result, err := r.db.Exec(query, staff.ID, staff.Department, staff.Shift, staff.Phone, staff.UpdatedAt)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseUpdateFailed)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(errors.ResourceNotFound)
	}

	cacheKey := fmt.Sprintf("staff:%s", staff.ID)
	_ = r.cacheHelper.Delete(context.Background(), cacheKey)

	return nil
}

// UpdateOnDutyStatus updates the on-duty status
func (r *staffRepository) UpdateOnDutyStatus(id string, onDuty bool) error {
	query := `UPDATE staff SET on_duty = $2, updated_at = $3 WHERE id = $1`

	result, err := r.db.Exec(query, id, onDuty, time.Now())
	if err != nil {
		return errors.Wrap(err, errors.DatabaseUpdateFailed)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(errors.ResourceNotFound)
	}

	cacheKey := fmt.Sprintf("staff:%s", id)
	_ = r.cacheHelper.Delete(context.Background(), cacheKey)

	return nil
}

// Delete deletes a staff member
func (r *staffRepository) Delete(id string) error {
	query := `DELETE FROM staff WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseDeleteFailed)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(errors.ResourceNotFound)
	}

	cacheKey := fmt.Sprintf("staff:%s", id)
	_ = r.cacheHelper.Delete(context.Background(), cacheKey)

	return nil
}

// AssignToRoom assigns a staff member to a room
func (r *staffRepository) AssignToRoom(staffID, roomID string, isPrimary bool) error {
	id, _ := uuid.NewV7()

	query := `
		INSERT INTO staff_room_assignments (id, staff_id, room_id, is_primary, assigned_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (staff_id, room_id) DO UPDATE SET is_primary = $4
	`

	_, err := r.db.Exec(query, id.String(), staffID, roomID, isPrimary, time.Now())
	if err != nil {
		return errors.Wrap(err, errors.DatabaseInsertFailed)
	}

	return nil
}

// RemoveFromRoom removes a staff member from a room
func (r *staffRepository) RemoveFromRoom(staffID, roomID string) error {
	query := `DELETE FROM staff_room_assignments WHERE staff_id = $1 AND room_id = $2`

	result, err := r.db.Exec(query, staffID, roomID)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseDeleteFailed)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(errors.ResourceNotFound)
	}

	return nil
}

// GetRoomAssignments gets all room assignments for a staff member
func (r *staffRepository) GetRoomAssignments(staffID string) ([]*StaffRoomAssignment, error) {
	var assignments []*StaffRoomAssignment

	query := `
		SELECT id, staff_id, room_id, is_primary, assigned_at
		FROM staff_room_assignments
		WHERE staff_id = $1
		ORDER BY is_primary DESC, assigned_at
	`

	rows, err := r.db.Query(query, staffID)
	if err != nil {
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	defer rows.Close()

	for rows.Next() {
		assignment := &StaffRoomAssignment{}
		err := rows.Scan(&assignment.ID, &assignment.StaffID, &assignment.RoomID, &assignment.IsPrimary, &assignment.AssignedAt)
		if err != nil {
			return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
		}
		assignments = append(assignments, assignment)
	}

	return assignments, nil
}

// GetStaffByRoom gets all staff assigned to a room
func (r *staffRepository) GetStaffByRoom(roomID string) ([]*StaffWithUser, error) {
	var staffList []*StaffWithUser

	query := `
		SELECT s.id, s.user_id, s.employee_id, s.type, s.department, s.shift, s.on_duty, s.phone, s.created_at, s.updated_at, u.name, u.email
		FROM staff s
		JOIN users u ON s.user_id = u.id
		JOIN staff_room_assignments sra ON s.id = sra.staff_id
		WHERE sra.room_id = $1
		ORDER BY sra.is_primary DESC, s.type, u.name
	`

	rows, err := r.db.Query(query, roomID)
	if err != nil {
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	defer rows.Close()

	for rows.Next() {
		staff := &StaffWithUser{}
		err := rows.Scan(
			&staff.ID, &staff.UserID, &staff.EmployeeID, &staff.Type, &staff.Department,
			&staff.Shift, &staff.OnDuty, &staff.Phone, &staff.CreatedAt, &staff.UpdatedAt,
			&staff.UserName, &staff.UserEmail,
		)
		if err != nil {
			return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
		}
		staffList = append(staffList, staff)
	}

	return staffList, nil
}

// IsAssignedToRoom checks if a staff member is assigned to a room
func (r *staffRepository) IsAssignedToRoom(staffID, roomID string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM staff_room_assignments WHERE staff_id = $1 AND room_id = $2`

	err := r.db.QueryRow(query, staffID, roomID).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, errors.DatabaseQueryFailed)
	}

	return count > 0, nil
}
