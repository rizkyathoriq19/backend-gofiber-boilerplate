package domain

import (
	"boilerplate-be/internal/module/auth/entity"
)

type AuthRepository interface {
	CreateUser(user *entity.User) error
	GetUserByEmail(email string) (*entity.User, error)
	GetUserByID(id string) (*entity.User, error)
	UpdateUser(user *entity.User) error
	StoreRefreshToken(userID, tokenID string) error
	ValidateRefreshToken(userID, tokenID string) (bool, error)
	RevokeRefreshToken(userID, tokenID string) error
}

type AuthUseCase interface {
	Register(email, password, name string) (*entity.User, string, string, error)
	Login(email, password string) (string, string, error)
	RefreshToken(refreshToken string) (string, string, error)
	Logout(userID, tokenID string) error
	GetProfile(userID string) (*entity.User, error)
	UpdateProfile(userID, name string) (*entity.User, error)
}