package auth

import (
	"time"

	"boilerplate-be/internal/pkg/enum"
)

type User struct {
	ID        string          `json:"id" db:"id"`
	Name      string          `json:"name" db:"name"`
	Email     string          `json:"email" db:"email"`
	Password  string          `json:"-" db:"password"`
	Role      enum.UserRole   `json:"role" db:"role"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt time.Time       `json:"updated_at" db:"updated_at"`
}
