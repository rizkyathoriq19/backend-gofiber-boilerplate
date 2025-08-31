package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"boilerplate-be/internal/infrastructure/enum"
	"boilerplate-be/internal/infrastructure/errors"
	"boilerplate-be/internal/infrastructure/redis"
	"boilerplate-be/internal/module/auth/entity"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type authRepository struct {
	db          *sql.DB
	redisClient *redis.Client
}

func NewAuthRepository(db *sql.DB, redisClient *redis.Client) *authRepository {
	return &authRepository{
		db:          db,
		redisClient: redisClient,
	}
}

func (r *authRepository) CreateUser(user *entity.User) error {
	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Role = enum.UserRoleUser

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
	user := &entity.User{}
	query := `
		SELECT id, name, email, password, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
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

	return nil
}

func (r *authRepository) StoreRefreshToken(userID, tokenID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("refresh_token:%s:%s", userID, tokenID)
	return r.redisClient.SetWithTTL(ctx, key, "1", 168*time.Hour)
}

func (r *authRepository) ValidateRefreshToken(userID, tokenID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("refresh_token:%s:%s", userID, tokenID)
	return r.redisClient.Exists(ctx, key)
}

func (r *authRepository) RevokeRefreshToken(userID, tokenID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("refresh_token:%s:%s", userID, tokenID)
	return r.redisClient.DeleteKey(ctx, key)
}