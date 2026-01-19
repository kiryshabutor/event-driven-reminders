package storage

import (
	"time"

	"github.com/kiribu/jwt-practice/models"
)

// Storage определяет интерфейс для работы с хранилищем данных
type Storage interface {
	// User operations
	CreateUser(username, password string) (*models.User, error)
	GetUser(username string) (*models.User, error)
	GetUserByID(userID int) (*models.User, error)
	ValidatePassword(username, password string) (*models.User, error)

	// Refresh token operations
	SaveRefreshToken(token string, userID int, expiresAt time.Time) error
	ValidateRefreshToken(token string) (int, error) // возвращает userID
	DeleteRefreshToken(token string) error
	DeleteUserRefreshTokens(userID int) error // удалить все токены пользователя
}

// Store - глобальная переменная для хранилища
var Store Storage
