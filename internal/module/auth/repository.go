package auth

import (
	"context"
	"database/sql"
	"time"

	"boilerplate-be/internal/shared/errors"
	"boilerplate-be/internal/shared/utils"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type authRepository struct {
	db          *sql.DB
	cacheHelper *utils.CacheHelper
}

func NewAuthRepository(db *sql.DB, cacheHelper *utils.CacheHelper) *authRepository {
	return &authRepository{
		db:          db,
		cacheHelper: cacheHelper,
	}
}

func (r *authRepository) CreateUser(user *User) error {
	id, _ := uuid.NewV7()
	user.ID = id.String()
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

func (r *authRepository) GetUserByEmail(email string) (*User, error) {
	user := &User{}
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

func (r *authRepository) GetUserByID(id string) (*User, error) {
	cacheKey := r.cacheHelper.BuildUserCacheKey(id, "profile")

	cachedData, err := r.cacheHelper.GetOrSet(context.Background(), cacheKey, func() (interface{}, error) {
		dbUser := &User{}
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

	user, ok := cachedData.(*User)
	if !ok {
		
		return nil, errors.New(errors.InternalServerError)
	}

	return user, nil
}

func (r *authRepository) UpdateUser(user *User) error {
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
