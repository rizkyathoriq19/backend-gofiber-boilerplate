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
	_, err := u.authRepo.GetUserByEmail(email)
	if err == nil {
		return nil, "", "", errors.New(errors.EmailExists)
	}

	hashedPassword, err := helper.HashPassword(password)
	if err != nil {
		return nil, "", "", errors.Wrap(err, errors.PasswordHashFailed)
	}

	user := &entity.User{
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

	if err := u.authRepo.StoreRefreshToken(user.ID, refreshClaims.ID); err != nil {
		return nil, "", "", errors.Wrap(err, errors.CacheStoreFailed)
	}

	return user, accessToken, refreshToken, nil
}

func (u *authUseCase) Login(email, password string) (string, string, error) {
    user, err := u.authRepo.GetUserByEmail(email)
    if err != nil {
        if appErr, ok := errors.IsAppError(err); ok {
            return "", "", appErr
        }
        return "", "", err
    }

    if err := helper.CheckPassword(user.Password, password); err != nil {
        return "", "", errors.New(errors.InvalidCredentials)
    }

    accessToken, refreshToken, err := u.jwtManager.GenerateTokenPair(user.ID, user.Email, user.Role)
    if err != nil {
        return "", "", errors.Wrap(err, errors.TokenGenerationFailed)
    }

    refreshClaims, err := u.jwtManager.ValidateToken(refreshToken)
    if err != nil {
        return "", "", errors.Wrap(err, errors.TokenGenerationFailed)
    }
    if err := u.authRepo.StoreRefreshToken(user.ID, refreshClaims.ID); err != nil {
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

	exists, err := u.authRepo.ValidateRefreshToken(claims.UserID, claims.ID)
	if err != nil {
		return "", "", errors.Wrap(err, errors.CacheError)
	}
	if !exists {
		return "", "", errors.New(errors.InvalidToken)
	}

	user, err := u.authRepo.GetUserByID(claims.UserID)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return "", "", appErr
		}
		return "", "", errors.Wrap(err, errors.DatabaseScanFailed)
	}

	newAccessToken, newRefreshToken, err := u.jwtManager.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		return "", "", errors.Wrap(err, errors.TokenGenerationFailed)
	}

	if err := u.authRepo.RevokeRefreshToken(claims.UserID, claims.ID); err != nil {
		return "", "", errors.Wrap(err, errors.CacheError)
	}

	newRefreshClaims, err := u.jwtManager.ValidateToken(newRefreshToken)
	if err != nil {
		return "", "", errors.Wrap(err, errors.TokenGenerationFailed)
	}

	if err := u.authRepo.StoreRefreshToken(user.ID, newRefreshClaims.ID); err != nil {
		return "", "", errors.Wrap(err, errors.CacheStoreFailed)
	}

	return newAccessToken, newRefreshToken, nil
}

func (u *authUseCase) Logout(userID, tokenID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := u.jwtManager.BlacklistToken(ctx, u.redisClient, tokenID, 24*time.Hour); err != nil {
		return errors.Wrap(err, errors.CacheStoreFailed)
	}

	// Revoke all refresh tokens for the user
	// Note: This is a simple implementation. In production, you might want to be more selective
	if err := u.authRepo.RevokeRefreshToken(userID, "*"); err != nil {
		return errors.Wrap(err, errors.CacheError)
	}

	return nil
}

func (u *authUseCase) GetProfile(userID string) (*entity.User, error) {
	return u.authRepo.GetUserByID(userID)
}

func (u *authUseCase) UpdateProfile(userID, name string) (*entity.User, error) {
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