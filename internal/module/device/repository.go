package device

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"boilerplate-be/internal/pkg/errors"
	"boilerplate-be/internal/pkg/utils"

	"github.com/google/uuid"
)

type deviceRepository struct {
	db          *sql.DB
	cacheHelper *utils.CacheHelper
}

// NewDeviceRepository creates a new device repository
func NewDeviceRepository(db *sql.DB, cacheHelper *utils.CacheHelper) DeviceRepository {
	return &deviceRepository{
		db:          db,
		cacheHelper: cacheHelper,
	}
}

// Create creates a new device
func (r *deviceRepository) Create(device *Device) error {
	id, _ := uuid.NewV7()
	device.ID = id.String()
	device.CreatedAt = time.Now()
	device.UpdatedAt = time.Now()
	device.Status = DeviceStatusOffline

	configJSON, err := json.Marshal(device.Config)
	if err != nil {
		return errors.Wrap(err, errors.InternalServerError)
	}

	query := `
		INSERT INTO devices (id, room_id, type, serial_number, name, status, api_key_hash, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = r.db.Exec(query, device.ID, device.RoomID, device.Type, device.SerialNumber, device.Name, device.Status, device.APIKeyHash, configJSON, device.CreatedAt, device.UpdatedAt)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseInsertFailed)
	}

	return nil
}

// GetByID gets a device by ID
func (r *deviceRepository) GetByID(id string) (*Device, error) {
	cacheKey := fmt.Sprintf("device:%s", id)

	cachedData, err := r.cacheHelper.GetOrSet(context.Background(), cacheKey, func() (interface{}, error) {
		return r.queryDevice("id", id)
	}, 5*time.Minute)

	if err != nil {
		return nil, err
	}

	device, ok := cachedData.(*Device)
	if !ok {
		return nil, errors.New(errors.InternalServerError)
	}

	return device, nil
}

// GetBySerialNumber gets a device by serial number
func (r *deviceRepository) GetBySerialNumber(serialNumber string) (*Device, error) {
	return r.queryDevice("serial_number", serialNumber)
}

// GetByAPIKey gets a device by API key hash
func (r *deviceRepository) GetByAPIKey(apiKeyHash string) (*Device, error) {
	return r.queryDevice("api_key_hash", apiKeyHash)
}

// queryDevice is a helper function to query device by a field
func (r *deviceRepository) queryDevice(field, value string) (*Device, error) {
	device := &Device{}
	var configJSON []byte

	query := fmt.Sprintf(`
		SELECT id, room_id, type, serial_number, name, status, api_key_hash, config, last_heartbeat, created_at, updated_at
		FROM devices
		WHERE %s = $1
	`, field)

	err := r.db.QueryRow(query, value).Scan(
		&device.ID, &device.RoomID, &device.Type, &device.SerialNumber, &device.Name,
		&device.Status, &device.APIKeyHash, &configJSON, &device.LastHeartbeat,
		&device.CreatedAt, &device.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(errors.ResourceNotFound)
		}
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}

	if len(configJSON) > 0 {
		if err := json.Unmarshal(configJSON, &device.Config); err != nil {
			return nil, errors.Wrap(err, errors.InternalServerError)
		}
	}

	return device, nil
}

// GetAll gets all devices with filters
func (r *deviceRepository) GetAll(filter *DeviceFilter) ([]*Device, int, error) {
	var devices []*Device
	var total int

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 10
	}

	baseQuery := `FROM devices WHERE 1=1`
	var args []interface{}
	argIndex := 1

	if filter.RoomID != "" {
		baseQuery += fmt.Sprintf(" AND room_id = $%d", argIndex)
		args = append(args, filter.RoomID)
		argIndex++
	}
	if filter.Type != "" {
		baseQuery += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, filter.Type)
		argIndex++
	}
	if filter.Status != "" {
		baseQuery += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, filter.Status)
		argIndex++
	}

	countQuery := `SELECT COUNT(*) ` + baseQuery
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, errors.Wrap(err, errors.DatabaseQueryFailed)
	}

	offset := (filter.Page - 1) * filter.Limit
	selectQuery := fmt.Sprintf(`SELECT id, room_id, type, serial_number, name, status, config, last_heartbeat, created_at, updated_at %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, baseQuery, argIndex, argIndex+1)
	args = append(args, filter.Limit, offset)

	rows, err := r.db.Query(selectQuery, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	defer rows.Close()

	for rows.Next() {
		device := &Device{}
		var configJSON []byte
		err := rows.Scan(
			&device.ID, &device.RoomID, &device.Type, &device.SerialNumber, &device.Name,
			&device.Status, &configJSON, &device.LastHeartbeat, &device.CreatedAt, &device.UpdatedAt,
		)
		if err != nil {
			return nil, 0, errors.Wrap(err, errors.DatabaseQueryFailed)
		}
		if len(configJSON) > 0 {
			if err := json.Unmarshal(configJSON, &device.Config); err != nil {
				continue
			}
		}
		devices = append(devices, device)
	}

	return devices, total, nil
}

// GetByRoomID gets all devices in a room
func (r *deviceRepository) GetByRoomID(roomID string) ([]*Device, error) {
	var devices []*Device

	query := `
		SELECT id, room_id, type, serial_number, name, status, config, last_heartbeat, created_at, updated_at
		FROM devices
		WHERE room_id = $1
		ORDER BY type, name
	`

	rows, err := r.db.Query(query, roomID)
	if err != nil {
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	defer rows.Close()

	for rows.Next() {
		device := &Device{}
		var configJSON []byte
		err := rows.Scan(
			&device.ID, &device.RoomID, &device.Type, &device.SerialNumber, &device.Name,
			&device.Status, &configJSON, &device.LastHeartbeat, &device.CreatedAt, &device.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
		}
		if len(configJSON) > 0 {
			if err := json.Unmarshal(configJSON, &device.Config); err != nil {
				continue
			}
		}
		devices = append(devices, device)
	}

	return devices, nil
}

// Update updates a device
func (r *deviceRepository) Update(device *Device) error {
	device.UpdatedAt = time.Now()

	configJSON, err := json.Marshal(device.Config)
	if err != nil {
		return errors.Wrap(err, errors.InternalServerError)
	}

	query := `
		UPDATE devices 
		SET room_id = $2, name = $3, config = $4, updated_at = $5
		WHERE id = $1
	`

	result, err := r.db.Exec(query, device.ID, device.RoomID, device.Name, configJSON, device.UpdatedAt)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseUpdateFailed)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(errors.ResourceNotFound)
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("device:%s", device.ID)
	_ = r.cacheHelper.Delete(context.Background(), cacheKey)

	return nil
}

// UpdateStatus updates a device status
func (r *deviceRepository) UpdateStatus(id string, status DeviceStatus) error {
	query := `UPDATE devices SET status = $2, updated_at = $3 WHERE id = $1`

	result, err := r.db.Exec(query, id, status, time.Now())
	if err != nil {
		return errors.Wrap(err, errors.DatabaseUpdateFailed)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(errors.ResourceNotFound)
	}

	cacheKey := fmt.Sprintf("device:%s", id)
	_ = r.cacheHelper.Delete(context.Background(), cacheKey)

	return nil
}

// UpdateHeartbeat updates the last heartbeat time
func (r *deviceRepository) UpdateHeartbeat(id string) error {
	query := `UPDATE devices SET last_heartbeat = $2, status = $3, updated_at = $4 WHERE id = $1`

	now := time.Now()
	result, err := r.db.Exec(query, id, now, DeviceStatusOnline, now)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseUpdateFailed)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(errors.ResourceNotFound)
	}

	cacheKey := fmt.Sprintf("device:%s", id)
	_ = r.cacheHelper.Delete(context.Background(), cacheKey)

	return nil
}

// Delete deletes a device
func (r *deviceRepository) Delete(id string) error {
	query := `DELETE FROM devices WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseDeleteFailed)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(errors.ResourceNotFound)
	}

	cacheKey := fmt.Sprintf("device:%s", id)
	_ = r.cacheHelper.Delete(context.Background(), cacheKey)

	return nil
}
