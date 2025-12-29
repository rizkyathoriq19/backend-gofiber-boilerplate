package auth

type AuthRepository interface {
	CreateUser(user *User) error
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id string) (*User, error)
	UpdateUser(user *User) error
}

type AuthUseCase interface {
	Register(email, password, name string) (*User, string, string, error)
	Login(email, password string) (string, string, error)
	RefreshToken(refreshToken string) (string, string, error)
	Logout(userID, tokenID string) error
	GetProfile(userID string) (*User, error)
	UpdateProfile(userID, name string) (*User, error)
}
