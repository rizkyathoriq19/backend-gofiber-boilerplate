package device

// DeviceRepository defines the interface for device data operations
type DeviceRepository interface {
	Create(device *Device) error
	GetByID(id string) (*Device, error)
	GetBySerialNumber(serialNumber string) (*Device, error)
	GetByAPIKey(apiKeyHash string) (*Device, error)
	GetAll(filter *DeviceFilter) ([]*Device, int, error)
	GetByRoomID(roomID string) ([]*Device, error)
	Update(device *Device) error
	UpdateStatus(id string, status DeviceStatus) error
	UpdateHeartbeat(id string) error
	Delete(id string) error
}

// DeviceUseCase defines the interface for device business logic
type DeviceUseCase interface {
	RegisterDevice(req *RegisterDeviceRequest) (*Device, string, error)
	GetDevice(id string) (*Device, error)
	GetDevices(filter *DeviceFilter) ([]*Device, int, error)
	GetDevicesByRoom(roomID string) ([]*Device, error)
	UpdateDevice(id string, req *UpdateDeviceRequest) (*Device, error)
	UpdateDeviceStatus(id string, status DeviceStatus) error
	Heartbeat(id string) error
	RegenerateAPIKey(id string) (string, error)
	DeleteDevice(id string) error
	ValidateAPIKey(apiKey string) (*Device, error)
}
