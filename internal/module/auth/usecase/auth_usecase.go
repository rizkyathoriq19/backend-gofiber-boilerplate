package usecase

import (
	"context"
	"time"

	"boilerplate-be/internal/infrastructure/errors"
	"boilerplate-be/internal/infrastructure/helper"
	"boilerplate-be/internal/infrastructure/redis"
	"boilerplate-be/internal/infrastructure/token"
	"boilerplate-be/internal/module/auth/domain"
	"boilerplate-be/internal/module/auth/entity"
)

type authUseCase struct {
	authRepo    domain.AuthRepository
	jwtManager  *token.JWTManager
	redisClient *redis.Client
}

func NewAuthUseCase(authRepo domain.AuthRepository, jwtManager *token.JWTManager, redisClient *redis.Client) *authUseCase {
	return &authUseCase{
		authRepo:    authRepo,
		jwtManager:  jwtManager,
		redisClient: redisClient,
	}
}

func (u *authUseCase) Register(email, password, name string) (*entity.User, string, string, error) {
	// Check if user already exists
	_, err := u.authRepo.GetUserByEmail(email)
	if err == nil {
		return nil, "", "", errors.ErrEmailHasBeenUsed
	}

	// Hash password
	hashedPassword, err := helper.HashPassword(password)
	if err != nil {
		return nil, "", "", errors.Wrap(err, errors.ServerErrorCantGeneratePassword)
	}

	// Create user
	user := &entity.User{
		Name:     name,
		Email:    email,
		Password: hashedPassword,
	}

	if err := u.authRepo.CreateUser(user); err != nil {
		return nil, "", "", err
	}

	// Generate token pair
	accessToken, refreshToken, err := u.jwtManager.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, "", "", errors.Wrap(err, errors.ServerErrorFailedGenerateToken)
	}

	// Validate refresh token to get token ID for storage
	refreshClaims, err := u.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, "", "", errors.Wrap(err, errors.ServerErrorFailedGenerateToken)
	}

	// Store refresh token
	if err := u.authRepo.StoreRefreshToken(user.ID, refreshClaims.ID); err != nil {
		return nil, "", "", errors.Wrap(err, errors.ServerErrorRedisCantStore)
	}

	return user, accessToken, refreshToken, nil
}

func (u *authUseCase) Login(email, password string) (string, string, error) {
    // ambil user
    user, err := u.authRepo.GetUserByEmail(email)
    if err != nil {
        if appErr, ok := errors.IsAppError(err); ok {
            return "", "", appErr
        }
        return "", "", err
    }

    // cek password
    if err := helper.CheckPassword(user.Password, password); err != nil {
        return "", "", errors.ErrInvalidCredential
    }

    // generate token
    accessToken, refreshToken, err := u.jwtManager.GenerateTokenPair(user.ID, user.Email, user.Role)
    if err != nil {
        return "", "", errors.Wrap(err, errors.ServerErrorFailedGenerateToken)
    }

    // validate & store refresh token
    refreshClaims, err := u.jwtManager.ValidateToken(refreshToken)
    if err != nil {
        return "", "", errors.Wrap(err, errors.ServerErrorFailedGenerateToken)
    }
    if err := u.authRepo.StoreRefreshToken(user.ID, refreshClaims.ID); err != nil {
        return "", "", errors.Wrap(err, errors.ServerErrorRedisCantStore)
    }

    return accessToken, refreshToken, nil
}

func (u *authUseCase) RefreshToken(refreshTokenString string) (string, string, error) {
	// Validate refresh token
	claims, err := u.jwtManager.ValidateToken(refreshTokenString)
	if err != nil {
		return "", "", errors.ErrInvalidToken
	}

	// Check if token type is refresh
	if claims.TokenType != "refresh" {
		return "", "", errors.ErrInvalidToken
	}

	// Check if refresh token exists in Redis
	exists, err := u.authRepo.ValidateRefreshToken(claims.UserID, claims.ID)
	if err != nil {
		return "", "", errors.Wrap(err, errors.ServerErrorRedis)
	}
	if !exists {
		return "", "", errors.ErrInvalidToken
	}

	// Get user details
	user, err := u.authRepo.GetUserByID(claims.UserID)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return "", "", appErr
		}
		return "", "", errors.Wrap(err, errors.ServerCantScanUserData)
	}

	// Generate new token pair
	newAccessToken, newRefreshToken, err := u.jwtManager.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		return "", "", errors.Wrap(err, errors.ServerErrorFailedGenerateToken)
	}

	// Revoke old refresh token
	if err := u.authRepo.RevokeRefreshToken(claims.UserID, claims.ID); err != nil {
		return "", "", errors.Wrap(err, errors.ServerErrorRedis)
	}

	// Store new refresh token
	newRefreshClaims, err := u.jwtManager.ValidateToken(newRefreshToken)
	if err != nil {
		return "", "", errors.Wrap(err, errors.ServerErrorFailedGenerateToken)
	}

	if err := u.authRepo.StoreRefreshToken(user.ID, newRefreshClaims.ID); err != nil {
		return "", "", errors.Wrap(err, errors.ServerErrorRedisCantStore)
	}

	return newAccessToken, newRefreshToken, nil
}

func (u *authUseCase) Logout(userID, tokenID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Blacklist the current access token
	if err := u.jwtManager.BlacklistToken(ctx, u.redisClient, tokenID, 24*time.Hour); err != nil {
		return errors.Wrap(err, errors.ServerErrorRedisCantStore)
	}

	// Revoke all refresh tokens for the user
	// Note: This is a simple implementation. In production, you might want to be more selective
	if err := u.authRepo.RevokeRefreshToken(userID, "*"); err != nil {
		return errors.Wrap(err, errors.ServerErrorRedis)
	}

	return nil
}

func (u *authUseCase) GetProfile(userID string) (*entity.User, error) {
	return u.authRepo.GetUserByID(userID)
}

func (u *authUseCase) UpdateProfile(userID, name string) (*entity.User, error) {
	// Get current user
	user, err := u.authRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Update user data
	user.Name = name

	// Update in repository
	if err := u.authRepo.UpdateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}