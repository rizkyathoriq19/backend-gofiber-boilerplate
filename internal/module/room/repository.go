package room

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"boilerplate-be/internal/pkg/errors"
	"boilerplate-be/internal/pkg/utils"

	"github.com/google/uuid"
)

type roomRepository struct {
	db          *sql.DB
	cacheHelper *utils.CacheHelper
}

// NewRoomRepository creates a new room repository
func NewRoomRepository(db *sql.DB, cacheHelper *utils.CacheHelper) RoomRepository {
	return &roomRepository{
		db:          db,
		cacheHelper: cacheHelper,
	}
}

// Create creates a new room
func (r *roomRepository) Create(room *Room) error {
	id, _ := uuid.NewV7()
	room.ID = id.String()
	room.CreatedAt = time.Now()
	room.UpdatedAt = time.Now()
	if room.Capacity == 0 {
		room.Capacity = 1
	}
	room.IsActive = true

	query := `
		INSERT INTO rooms (id, name, type, floor, building, capacity, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Exec(query, room.ID, room.Name, room.Type, room.Floor, room.Building, room.Capacity, room.IsActive, room.CreatedAt, room.UpdatedAt)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseInsertFailed)
	}

	return nil
}

// GetByID gets a room by ID
func (r *roomRepository) GetByID(id string) (*Room, error) {
	cacheKey := fmt.Sprintf("room:%s", id)

	cachedData, err := r.cacheHelper.GetOrSet(context.Background(), cacheKey, func() (interface{}, error) {
		room := &Room{}
		query := `
			SELECT id, name, type, floor, building, capacity, is_active, created_at, updated_at
			FROM rooms
			WHERE id = $1
		`
		err := r.db.QueryRow(query, id).Scan(
			&room.ID, &room.Name, &room.Type, &room.Floor, &room.Building,
			&room.Capacity, &room.IsActive, &room.CreatedAt, &room.UpdatedAt,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, errors.New(errors.ResourceNotFound)
			}
			return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
		}
		return room, nil
	}, 5*time.Minute)

	if err != nil {
		return nil, err
	}

	room, ok := cachedData.(*Room)
	if !ok {
		return nil, errors.New(errors.InternalServerError)
	}

	return room, nil
}

// GetAll gets all rooms with filters
func (r *roomRepository) GetAll(filter *RoomFilter) ([]*Room, int, error) {
	var rooms []*Room
	var total int

	// Default pagination
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 10
	}

	// Build query with filters
	baseQuery := `FROM rooms WHERE 1=1`
	var args []interface{}
	argIndex := 1

	if filter.Type != "" {
		baseQuery += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, filter.Type)
		argIndex++
	}
	if filter.Floor != "" {
		baseQuery += fmt.Sprintf(" AND floor = $%d", argIndex)
		args = append(args, filter.Floor)
		argIndex++
	}
	if filter.Building != "" {
		baseQuery += fmt.Sprintf(" AND building = $%d", argIndex)
		args = append(args, filter.Building)
		argIndex++
	}
	if filter.IsActive != nil {
		baseQuery += fmt.Sprintf(" AND is_active = $%d", argIndex)
		args = append(args, *filter.IsActive)
		argIndex++
	}

	// Count total
	countQuery := `SELECT COUNT(*) ` + baseQuery
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, errors.Wrap(err, errors.DatabaseQueryFailed)
	}

	// Get paginated results
	offset := (filter.Page - 1) * filter.Limit
	selectQuery := fmt.Sprintf(`SELECT id, name, type, floor, building, capacity, is_active, created_at, updated_at %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, baseQuery, argIndex, argIndex+1)
	args = append(args, filter.Limit, offset)

	rows, err := r.db.Query(selectQuery, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	defer rows.Close()

	for rows.Next() {
		room := &Room{}
		err := rows.Scan(
			&room.ID, &room.Name, &room.Type, &room.Floor, &room.Building,
			&room.Capacity, &room.IsActive, &room.CreatedAt, &room.UpdatedAt,
		)
		if err != nil {
			return nil, 0, errors.Wrap(err, errors.DatabaseQueryFailed)
		}
		rooms = append(rooms, room)
	}

	return rooms, total, nil
}

// Update updates a room
func (r *roomRepository) Update(room *Room) error {
	room.UpdatedAt = time.Now()

	// Build dynamic update query
	var setClauses []string
	var args []interface{}
	argIndex := 1

	setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIndex))
	args = append(args, room.Name)
	argIndex++

	setClauses = append(setClauses, fmt.Sprintf("type = $%d", argIndex))
	args = append(args, room.Type)
	argIndex++

	setClauses = append(setClauses, fmt.Sprintf("floor = $%d", argIndex))
	args = append(args, room.Floor)
	argIndex++

	setClauses = append(setClauses, fmt.Sprintf("building = $%d", argIndex))
	args = append(args, room.Building)
	argIndex++

	setClauses = append(setClauses, fmt.Sprintf("capacity = $%d", argIndex))
	args = append(args, room.Capacity)
	argIndex++

	setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argIndex))
	args = append(args, room.IsActive)
	argIndex++

	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, room.UpdatedAt)
	argIndex++

	query := fmt.Sprintf(`UPDATE rooms SET %s WHERE id = $%d`, strings.Join(setClauses, ", "), argIndex)
	args = append(args, room.ID)

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseUpdateFailed)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, errors.DatabaseError)
	}

	if rowsAffected == 0 {
		return errors.New(errors.ResourceNotFound)
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("room:%s", room.ID)
	if err := r.cacheHelper.Delete(context.Background(), cacheKey); err != nil {
		return errors.Wrap(err, errors.CacheError)
	}

	return nil
}

// Delete deletes a room
func (r *roomRepository) Delete(id string) error {
	query := `DELETE FROM rooms WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseDeleteFailed)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, errors.DatabaseError)
	}

	if rowsAffected == 0 {
		return errors.New(errors.ResourceNotFound)
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("room:%s", id)
	if err := r.cacheHelper.Delete(context.Background(), cacheKey); err != nil {
		return errors.Wrap(err, errors.CacheError)
	}

	return nil
}
