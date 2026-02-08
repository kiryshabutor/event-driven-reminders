package storage

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kiribu/jwt-practice/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Storage interface {
	CreateUser(ctx context.Context, username, password string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	ValidatePassword(ctx context.Context, username, password string) (*models.User, error)
	SaveRefreshToken(ctx context.Context, token string, userID uuid.UUID, expiresAt time.Time) error
	ValidateRefreshToken(ctx context.Context, token string) (uuid.UUID, error)
	DeleteRefreshToken(ctx context.Context, token string) error
}

type PostgresStorage struct {
	db *gorm.DB
}

func NewPostgresStorage(db *gorm.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func (s *PostgresStorage) CreateUser(ctx context.Context, username, password string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:           uuid.Must(uuid.NewV7()),
		Username:     username,
		PasswordHash: string(hashedPassword),
	}

	result := s.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		return nil, errors.New("user with this username already exists")
	}

	return user, nil
}

func (s *PostgresStorage) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	result := s.db.WithContext(ctx).Where("username = ?", username).First(&user)
	if result.Error != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

func (s *PostgresStorage) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	result := s.db.WithContext(ctx).First(&user, "id = ?", id)
	if result.Error != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

func (s *PostgresStorage) ValidatePassword(ctx context.Context, username, password string) (*models.User, error) {
	user, err := s.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid password")
	}

	return user, nil
}

func (s *PostgresStorage) SaveRefreshToken(ctx context.Context, token string, userID uuid.UUID, expiresAt time.Time) error {
	refreshToken := &models.RefreshToken{
		ID:        uuid.Must(uuid.NewV7()),
		Token:     token,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}
	return s.db.WithContext(ctx).Create(refreshToken).Error
}

func (s *PostgresStorage) ValidateRefreshToken(ctx context.Context, token string) (uuid.UUID, error) {
	var rt models.RefreshToken
	result := s.db.WithContext(ctx).Where("token = ?", token).First(&rt)
	if result.Error != nil {
		return uuid.Nil, errors.New("token not found")
	}

	if time.Now().After(rt.ExpiresAt) {
		s.DeleteRefreshToken(ctx, token)
		return uuid.Nil, errors.New("token expired")
	}

	return rt.UserID, nil
}

func (s *PostgresStorage) DeleteRefreshToken(ctx context.Context, token string) error {
	return s.db.WithContext(ctx).Where("token = ?", token).Delete(&models.RefreshToken{}).Error
}
