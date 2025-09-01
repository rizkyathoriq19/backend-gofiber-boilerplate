package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"boilerplate-be/internal/infrastructure/errors"
	"boilerplate-be/internal/infrastructure/helper"
	"boilerplate-be/internal/module/auth/entity"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type authRepository struct {
	db          *sql.DB
	cacheHelper *helper.CacheHelper
}

func NewAuthRepository(db *sql.DB, cacheHelper *helper.CacheHelper) *authRepository {
	return &authRepository{
		db:          db,
		cacheHelper: cacheHelper,
	}
}

func (r *authRepository) CreateUser(user *entity.User) error {
	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	if user.Role == "" {
		user.Role = "user"
	}


	query := `
		INSERT INTO users (id, name, email, password, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(query, user.ID, user.Name, user.Email, user.Password, user.Role, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
			return errors.New(errors.EmailExists)
		}
		return errors.Wrap(err, errors.DatabaseInsertFailed)
	}

	return nil
}

func (r *authRepository) GetUserByEmail(email string) (*entity.User, error) {
	user := &entity.User{}
	query := `
		SELECT id, name, email, password, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(errors.AccountNotFound)
		}
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}

	return user, nil
}

func (r *authRepository) GetUserByID(id string) (*entity.User, error) {
	cacheKey := r.cacheHelper.BuildUserCacheKey(id, "profile")

	cachedData, err := r.cacheHelper.GetOrSet(context.Background(), cacheKey, func() (interface{}, error) {
		dbUser := &entity.User{}
		query := `
			SELECT id, name, email, password, role, created_at, updated_at
			FROM users
			WHERE id = $1
		`
		err := r.db.QueryRow(query, id).Scan(
			&dbUser.ID, &dbUser.Name, &dbUser.Email, &dbUser.Password,
			&dbUser.Role, &dbUser.CreatedAt, &dbUser.UpdatedAt,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, errors.New(errors.AccountNotFound)
			}
			return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
		}
		return dbUser, nil
	}, 5*time.Minute)

	if err != nil {
		return nil, err
	}

	user, ok := cachedData.(*entity.User)
	if !ok {
		
		return nil, errors.New(errors.InternalServerError)
	}

	return user, nil
}

func (r *authRepository) UpdateUser(user *entity.User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users
		SET name = $2, updated_at = $3
		WHERE id = $1
	`

	result, err := r.db.Exec(query, user.ID, user.Name, user.UpdatedAt)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseUpdateFailed)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, errors.DatabaseError)
	}

	if rowsAffected == 0 {
		return errors.New(errors.AccountNotFound)
	}

	if err := r.cacheHelper.InvalidateUserCache(context.Background(), user.ID); err != nil {
		return errors.Wrap(err, errors.CacheError)
	}

	return nil
}

func (r *authRepository) StoreRefreshToken(userID, tokenID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("refresh_token:%s:%s", userID, tokenID)
	return r.cacheHelper.SetWithTTL(ctx, key, "1", 168*time.Hour)
}

func (r *authRepository) ValidateRefreshToken(userID, tokenID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("refresh_token:%s:%s", userID, tokenID)
	return r.cacheHelper.Exists(ctx, key)
}

func (r *authRepository) RevokeRefreshToken(userID, tokenID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("refresh_token:%s:%s", userID, tokenID)
	return r.cacheHelper.DeleteKey(ctx, key)
}