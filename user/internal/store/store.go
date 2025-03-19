package store

import (
	"context"
	"errors"
	"fmt"
	"github.com/HJyup/mlt-user/internal/service"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type Store struct {
	dbConn *pgx.Conn
}

func NewStore(dbConn *pgx.Conn) *Store {
	return &Store{dbConn: dbConn}
}

func (s *Store) CreateUser(ctx context.Context, username, email, password string) (string, error) {
	var existingID string
	err := s.dbConn.QueryRow(ctx, "SELECT id FROM users WHERE email = $1", email).Scan(&existingID)
	if err == nil {
		return "", errors.New("user with this email already exists")
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("error checking existing user: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	var userID string
	err = s.dbConn.QueryRow(ctx,
		"INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id",
		username, email, string(hashedPassword)).Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	return userID, nil
}

func (s *Store) AuthUser(ctx context.Context, email, password string) (*service.User, error) {
	user := &service.User{Email: email}
	var hashedPassword string

	err := s.dbConn.QueryRow(ctx, "SELECT id, username, password FROM users WHERE email = $1", email).Scan(&user.ID, &user.Username, &hashedPassword)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return nil, errors.New("invalid password")
	}

	return user, nil
}

func (s *Store) GetUser(ctx context.Context, userID string) (*service.User, error) {
	user := &service.User{}

	err := s.dbConn.QueryRow(ctx,
		"SELECT id, username, email FROM users WHERE id = $1",
		userID).Scan(&user.ID, &user.Username, &user.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (s *Store) DeleteUser(ctx context.Context, userID string) error {
	result, err := s.dbConn.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}
