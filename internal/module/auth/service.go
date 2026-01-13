package auth

import (
	"boilerplate-be/internal/shared/errors"
	"boilerplate-be/internal/shared/security"
)

type authUseCase struct {
	authRepo     AuthRepository
	jwtManager   *security.JWTManager
	tokenManager *security.TokenManager
}

func NewAuthUseCase(
	authRepo AuthRepository,
	jwtManager *security.JWTManager,
	tokenManager *security.TokenManager,
) *authUseCase {
	return &authUseCase{
		authRepo:     authRepo,
		jwtManager:   jwtManager,
		tokenManager: tokenManager,
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

	accessToken, refreshToken, err := u.jwtManager.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, "", "", errors.Wrap(err, errors.TokenGenerationFailed)
	}

	refreshClaims, err := u.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, "", "", errors.Wrap(err, errors.TokenGenerationFailed)
	}

	if err := u.tokenManager.StoreToken(user.ID, refreshClaims.ID); err != nil {
		return nil, "", "", errors.Wrap(err, errors.CacheStoreFailed)
	}

	return user, accessToken, refreshToken, nil
}

func (u *authUseCase) Login(email, password string) (string, string, error) {
	user, err := u.authRepo.GetUserByEmail(email)
	if err != nil {
		return "", "", errors.New(errors.AccountNotFound)
	}

	if err := security.CheckPassword(user.Password, password); err != nil {
		return "", "", errors.New(errors.PasswordMismatch)
	}

	accessToken, refreshToken, err := u.jwtManager.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		return "", "", errors.Wrap(err, errors.TokenGenerationFailed)
	}

	refreshClaims, err := u.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return "", "", errors.Wrap(err, errors.TokenGenerationFailed)
	}

	if err := u.tokenManager.StoreToken(user.ID, refreshClaims.ID); err != nil {
		return "", "", errors.Wrap(err, errors.CacheStoreFailed)
	}

	return accessToken, refreshToken, nil
}

func (u *authUseCase) RefreshToken(refreshTokenString string) (string, string, error) {
	claims, err := u.jwtManager.ValidateToken(refreshTokenString)
	if err != nil {
		return "", "", errors.New(errors.InvalidToken)
	}

	if claims.TokenType != "refresh" {
		return "", "", errors.New(errors.InvalidToken)
	}

	exists, err := u.tokenManager.ValidateToken(claims.UserID, claims.ID)
	if err != nil {
		return "", "", errors.Wrap(err, errors.CacheError)
	}
	if !exists {
		return "", "", errors.New(errors.InvalidToken)
	}

	user, err := u.authRepo.GetUserByID(claims.UserID)
	if err != nil {
		return "", "", errors.Wrap(err, errors.AccountNotFound)
	}

	newAccessToken, newRefreshToken, err := u.jwtManager.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		return "", "", errors.Wrap(err, errors.TokenGenerationFailed)
	}

	if err := u.tokenManager.RevokeToken(claims.UserID, claims.ID); err != nil {
		return "", "", errors.Wrap(err, errors.CacheError)
	}

	newRefreshClaims, err := u.jwtManager.ValidateToken(newRefreshToken)
	if err != nil {
		return "", "", errors.Wrap(err, errors.TokenGenerationFailed)
	}

	if err := u.tokenManager.StoreToken(user.ID, newRefreshClaims.ID); err != nil {
		return "", "", errors.Wrap(err, errors.CacheStoreFailed)
	}

	return newAccessToken, newRefreshToken, nil
}

func (u *authUseCase) Logout(userID, tokenID string) error {
	if err := u.tokenManager.BlacklistToken(userID, tokenID); err != nil {
		return errors.Wrap(err, errors.CacheError)
	}

	if err := u.tokenManager.RevokeAllUserTokens(userID); err != nil {
		return errors.Wrap(err, errors.CacheError)
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
