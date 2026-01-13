package device

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"boilerplate-be/internal/shared/errors"
)

type deviceUseCase struct {
	deviceRepo DeviceRepository
}

// NewDeviceUseCase creates a new device use case
func NewDeviceUseCase(deviceRepo DeviceRepository) DeviceUseCase {
	return &deviceUseCase{
		deviceRepo: deviceRepo,
	}
}

// generateAPIKey generates a random API key
func generateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("mp_%s", hex.EncodeToString(bytes)), nil
}

// hashAPIKey hashes an API key
func hashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}

// RegisterDevice registers a new device and returns API key
func (u *deviceUseCase) RegisterDevice(req *RegisterDeviceRequest) (*Device, string, error) {
	// Check if serial number already exists
	existing, err := u.deviceRepo.GetBySerialNumber(req.SerialNumber)
	if err == nil && existing != nil {
		return nil, "", errors.New(errors.ResourceAlreadyExists)
	}

	// Generate API key
	apiKey, err := generateAPIKey()
	if err != nil {
		return nil, "", errors.Wrap(err, errors.InternalServerError)
	}

	device := &Device{
		RoomID:       req.RoomID,
		Type:         req.Type,
		SerialNumber: req.SerialNumber,
		Name:         req.Name,
		Config:       req.Config,
		APIKeyHash:   hashAPIKey(apiKey),
	}

	if device.Config == nil {
		device.Config = make(map[string]interface{})
	}

	if err := u.deviceRepo.Create(device); err != nil {
		return nil, "", err
	}

	return device, apiKey, nil
}

// GetDevice gets a device by ID
func (u *deviceUseCase) GetDevice(id string) (*Device, error) {
	return u.deviceRepo.GetByID(id)
}

// GetDevices gets all devices with filters
func (u *deviceUseCase) GetDevices(filter *DeviceFilter) ([]*Device, int, error) {
	return u.deviceRepo.GetAll(filter)
}

// GetDevicesByRoom gets all devices in a room
func (u *deviceUseCase) GetDevicesByRoom(roomID string) ([]*Device, error) {
	return u.deviceRepo.GetByRoomID(roomID)
}

// UpdateDevice updates a device
func (u *deviceUseCase) UpdateDevice(id string, req *UpdateDeviceRequest) (*Device, error) {
	device, err := u.deviceRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.RoomID != nil {
		device.RoomID = req.RoomID
	}
	if req.Name != "" {
		device.Name = req.Name
	}
	if req.Config != nil {
		device.Config = req.Config
	}

	if err := u.deviceRepo.Update(device); err != nil {
		return nil, err
	}

	return device, nil
}

// UpdateDeviceStatus updates a device status
func (u *deviceUseCase) UpdateDeviceStatus(id string, status DeviceStatus) error {
	return u.deviceRepo.UpdateStatus(id, status)
}

// Heartbeat updates device heartbeat
func (u *deviceUseCase) Heartbeat(id string) error {
	return u.deviceRepo.UpdateHeartbeat(id)
}

// RegenerateAPIKey generates a new API key for a device
func (u *deviceUseCase) RegenerateAPIKey(id string) (string, error) {
	device, err := u.deviceRepo.GetByID(id)
	if err != nil {
		return "", err
	}

	apiKey, err := generateAPIKey()
	if err != nil {
		return "", errors.Wrap(err, errors.InternalServerError)
	}

	device.APIKeyHash = hashAPIKey(apiKey)
	if err := u.deviceRepo.Update(device); err != nil {
		return "", err
	}

	return apiKey, nil
}

// DeleteDevice deletes a device
func (u *deviceUseCase) DeleteDevice(id string) error {
	_, err := u.deviceRepo.GetByID(id)
	if err != nil {
		return err
	}

	return u.deviceRepo.Delete(id)
}

// ValidateAPIKey validates an API key and returns the device
func (u *deviceUseCase) ValidateAPIKey(apiKey string) (*Device, error) {
	hash := hashAPIKey(apiKey)
	device, err := u.deviceRepo.GetByAPIKey(hash)
	if err != nil {
		return nil, errors.New(errors.InvalidAPIKey)
	}
	return device, nil
}
