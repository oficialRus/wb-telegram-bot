package storage

import (
	"fmt"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// User — структура таблицы пользователей
type User struct {
	ID            uint   `gorm:"primaryKey"`  // Автоинкремент ID в базе
	TelegramID    int64  `gorm:"uniqueIndex"` // Уникальный Telegram ID
	Username      string // Никнейм пользователя
	Warehouses    string // Список складов в виде строки (например "123,456")
	CheckInterval int    // Интервал проверки лимитов в минутах
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
		TelegramID:    telegramID,
		Username:      username,
		Warehouses:    "",
		CheckInterval: 5, // По умолчанию 5 минут интервал
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

// AddWarehouseToUser — добавить ID склада пользователю
func (s *Storage) AddWarehouseToUser(telegramID int64, warehouseID int) error {
	var user User
	if err := s.db.Where("telegram_id = ?", telegramID).First(&user).Error; err != nil {
		return err
	}

	warehouses := strings.Split(user.Warehouses, ",")
	for _, w := range warehouses {
		if w == fmt.Sprint(warehouseID) {
			// Уже есть
			return nil
		}
	}

	if user.Warehouses != "" {
		user.Warehouses += ","
	}
	user.Warehouses += fmt.Sprint(warehouseID)

	return s.db.Save(&user).Error
}

// GetUserWarehouses — получить список складов пользователя
func (s *Storage) GetUserWarehouses(telegramID int64) ([]int, error) {
	var user User
	if err := s.db.Where("telegram_id = ?", telegramID).First(&user).Error; err != nil {
		return nil, err
	}

	var warehouseIDs []int
	if user.Warehouses == "" {
		return warehouseIDs, nil
	}

	ids := strings.Split(user.Warehouses, ",")
	for _, idStr := range ids {
		var id int
		fmt.Sscanf(idStr, "%d", &id)
		warehouseIDs = append(warehouseIDs, id)
	}

	return warehouseIDs, nil
}

// RemoveWarehouseFromUser — удалить ID склада у пользователя
func (s *Storage) RemoveWarehouseFromUser(telegramID int64, warehouseID int) error {
	var user User
	if err := s.db.Where("telegram_id = ?", telegramID).First(&user).Error; err != nil {
		return err
	}

	warehouses := strings.Split(user.Warehouses, ",")
	newWarehouses := []string{}

	for _, w := range warehouses {
		if w != fmt.Sprint(warehouseID) && w != "" {
			newWarehouses = append(newWarehouses, w)
		}
	}

	user.Warehouses = strings.Join(newWarehouses, ",")

	return s.db.Save(&user).Error
}

// GetAllUsers — получить всех пользователей
func (s *Storage) GetAllUsers() ([]User, error) {
	var users []User
	if err := s.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// UpdateCheckInterval — изменить интервал проверки для пользователя
func (s *Storage) UpdateCheckInterval(telegramID int64, interval int) error {
	return s.db.Model(&User{}).
		Where("telegram_id = ?", telegramID).
		Update("check_interval", interval).Error
}
