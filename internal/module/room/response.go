package room

import (
	"time"
)

// RoomResponse represents the room response
type RoomResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      RoomType  `json:"type"`
	Floor     string    `json:"floor"`
	Building  string    `json:"building"`
	Capacity  int       `json:"capacity"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RoomListResponse represents paginated room list response
type RoomListResponse struct {
	Rooms      []*RoomResponse `json:"rooms"`
	Total      int             `json:"total"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	TotalPages int             `json:"total_pages"`
}

// ToResponse converts Room entity to RoomResponse
func (r *Room) ToResponse() *RoomResponse {
	return &RoomResponse{
		ID:        r.ID,
		Name:      r.Name,
		Type:      r.Type,
		Floor:     r.Floor,
		Building:  r.Building,
		Capacity:  r.Capacity,
		IsActive:  r.IsActive,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
