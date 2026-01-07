package room

import (
	"time"
)

// RoomType represents the type of room
type RoomType string

const (
	RoomTypePatientRoom   RoomType = "patient_room"
	RoomTypeNurseStation  RoomType = "nurse_station"
	RoomTypeICU           RoomType = "icu"
	RoomTypeEmergency     RoomType = "emergency"
	RoomTypeOperatingRoom RoomType = "operating_room"
)

// Room represents a room in the hospital
type Room struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Type      RoomType  `json:"type" db:"type"`
	Floor     string    `json:"floor" db:"floor"`
	Building  string    `json:"building" db:"building"`
	Capacity  int       `json:"capacity" db:"capacity"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// RoomFilter represents filters for querying rooms
type RoomFilter struct {
	Type     RoomType `query:"type"`
	Floor    string   `query:"floor"`
	Building string   `query:"building"`
	IsActive *bool    `query:"is_active"`
	Page     int      `query:"page"`
	Limit    int      `query:"limit"`
}
