package storage

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kiribu/jwt-practice/models"
	"golang.org/x/crypto/bcrypt"
)

type PostgresStorage struct {
	db *sqlx.DB
}

func NewPostgresStorage(db *sqlx.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func (s *PostgresStorage) CreateUser(username, password string) (*models.User, error) {
	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// SQL запрос с RETURNING для получения созданной записи
	query := `
		INSERT INTO users (username, password_hash, created_at)
		VALUES ($1, $2, $3)
		RETURNING id, username, password_hash, created_at
	`

	user := &models.User{}
	err = s.db.QueryRowx(
		query,
		username,
		string(hashedPassword),
		time.Now(),
	).StructScan(user)

	if err != nil {
		// Проверка на нарушение уникальности
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, errors.New("пользователь уже существует")
		}
		return nil, err
	}

	return user, nil
}

func (s *PostgresStorage) GetUser(username string) (*models.User, error) {
	query := `SELECT id, username, password_hash, created_at FROM users WHERE username = $1`

	user := &models.User{}
	err := s.db.Get(user, query, username)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("пользователь не найден")
		}
		return nil, err
	}

	return user, nil
}

func (s *PostgresStorage) GetUserByID(userID int) (*models.User, error) {
	query := `SELECT id, username, password_hash, created_at FROM users WHERE id = $1`

	user := &models.User{}
	err := s.db.Get(user, query, userID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("пользователь не найден")
		}
		return nil, err
	}

	return user, nil
}

func (s *PostgresStorage) ValidatePassword(username, password string) (*models.User, error) {
	user, err := s.GetUser(username)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, errors.New("неверный пароль")
	}

	return user, nil
}

func (s *PostgresStorage) SaveRefreshToken(token string, userID int, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (token, user_id, created_at, expires_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := s.db.Exec(query, token, userID, time.Now(), expiresAt)
	return err
}

func (s *PostgresStorage) ValidateRefreshToken(token string) (int, error) {
	query := `
		SELECT user_id 
		FROM refresh_tokens 
		WHERE token = $1 AND expires_at > $2
	`

	var userID int
	err := s.db.Get(&userID, query, token, time.Now())

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("невалидный refresh token")
		}
		return 0, err
	}

	return userID, nil
}

func (s *PostgresStorage) DeleteRefreshToken(token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`
	_, err := s.db.Exec(query, token)
	return err
}

func (s *PostgresStorage) DeleteUserRefreshTokens(userID int) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := s.db.Exec(query, userID)
	return err
}

// CleanupExpiredTokens удаляет все истекшие токены
func (s *PostgresStorage) CleanupExpiredTokens() error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < $1`
	_, err := s.db.Exec(query, time.Now())
	return err
}
