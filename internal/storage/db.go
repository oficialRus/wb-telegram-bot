package storage

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// User — структура таблицы пользователей
type User struct {
	ID         uint   `gorm:"primaryKey"`  // Автоинкремент ID в базе
	TelegramID int64  `gorm:"uniqueIndex"` // Уникальный Telegram ID
	Username   string // Никнейм пользователя
	Warehouses string // Список складов в виде строки (например "123,456")
}

// Storage — обёртка для базы данных
type Storage struct {
	db *gorm.DB
}

// NewStorage — инициализация базы данных
func NewStorage(dbPath string) (*Storage, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Миграция таблицы пользователей
	err = db.AutoMigrate(&User{})
	if err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

// CreateUser — создать нового пользователя
func (s *Storage) CreateUser(telegramID int64, username string) error {
	user := User{
		TelegramID: telegramID,
		Username:   username,
		Warehouses: "",
	}
	return s.db.Create(&user).Error
}

// GetUserByTelegramID — получить пользователя по Telegram ID
func (s *Storage) GetUserByTelegramID(telegramID int64) (*User, error) {
	var user User
	result := s.db.Where("telegram_id = ?", telegramID).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// UserExists — проверка существует ли пользователь
func (s *Storage) UserExists(telegramID int64) (bool, error) {
	var count int64
	err := s.db.Model(&User{}).Where("telegram_id = ?", telegramID).Count(&count).Error
	return count > 0, err
}
