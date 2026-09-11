package auth

import (
	stderrors "errors"

	"boilerplate-be/internal/shared/enum"
	"boilerplate-be/internal/shared/errors"
	"boilerplate-be/internal/shared/security"
)

type authUseCase struct {
	authRepo       AuthRepository
	sessionManager *security.LoginSessionManager
}

func NewAuthUseCase(
	authRepo AuthRepository,
	sessionManager *security.LoginSessionManager,
) *authUseCase {
	return &authUseCase{
		authRepo:       authRepo,
		sessionManager: sessionManager,
	}
}

func (u *authUseCase) Register(email, password, name string) (*User, string, string, error) {
	_, err := u.authRepo.GetUserByEmail(email)
	if err == nil {
		return nil, "", "", errors.New(errors.EmailExists)
	}

	hashedPassword, err := security.HashPassword(password)
	if err != nil {
		return nil, "", "", errors.Wrap(err, errors.PasswordHashFailed)
	}

	user := &User{
		Name:     name,
		Email:    email,
		Password: hashedPassword,
	}

	if err := u.authRepo.CreateUser(user); err != nil {
		return nil, "", "", err
	}

	pair, err := u.sessionManager.Issue(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, "", "", sessionAppError(err, errors.TokenGenerationFailed)
	}

	return user, pair.AccessToken, pair.RefreshToken, nil
}

func (u *authUseCase) Login(email, password string) (string, string, error) {
	user, err := u.authRepo.GetUserByEmail(email)
	if err != nil {
		return "", "", errors.New(errors.AccountNotFound)
	}

	if err := security.CheckPassword(user.Password, password); err != nil {
		return "", "", errors.New(errors.PasswordMismatch)
	}

	pair, err := u.sessionManager.Issue(user.ID, user.Email, user.Role)
	if err != nil {
		return "", "", sessionAppError(err, errors.TokenGenerationFailed)
	}

	return pair.AccessToken, pair.RefreshToken, nil
}

func (u *authUseCase) RefreshToken(refreshTokenString string) (string, string, error) {
	claims, err := u.sessionManager.ValidateRefresh(refreshTokenString)
	if err != nil {
		return "", "", sessionAppError(err, errors.InvalidToken)
	}

	user, err := u.authRepo.GetUserByID(claims.UserID)
	if err != nil {
		return "", "", errors.Wrap(err, errors.AccountNotFound)
	}

	pair, err := u.sessionManager.Renew(refreshTokenString, user.ID, user.Email, user.Role)
	if err != nil {
		return "", "", sessionAppError(err, errors.TokenGenerationFailed)
	}

	return pair.AccessToken, pair.RefreshToken, nil
}

func (u *authUseCase) Logout(sessionID string) error {
	if err := u.sessionManager.Logout(sessionID); err != nil {
		return sessionAppError(err, errors.CacheError)
	}
	return nil
}

func (u *authUseCase) GetProfile(userID string) (*User, error) {
	return u.authRepo.GetUserByID(userID)
}

func (u *authUseCase) UpdateProfile(userID, name string) (*User, error) {
	user, err := u.authRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	user.Name = name

	if err := u.authRepo.UpdateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

func sessionAppError(err error, fallback enum.ErrorCode) error {
	if stderrors.Is(err, security.ErrSessionStoreUnavailable) {
		return errors.New(errors.AuthServiceUnavailable)
	}
	if stderrors.Is(err, security.ErrInvalidSession) {
		return errors.New(errors.InvalidToken)
	}
	return errors.Wrap(err, fallback)
}
