package bot

import (
	"postavkinBot/internal/wb"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Канал для получения обновлений
var UpdatesChan chan tgbotapi.Update

// waitForUserInput — ожидание ввода пользователя через общий канал UpdatesChan
func waitForUserInput(update tgbotapi.Update, callback func(input string)) {
	for nextUpdate := range UpdatesChan {
		if nextUpdate.Message != nil && nextUpdate.Message.Chat.ID == update.Message.Chat.ID {
			callback(nextUpdate.Message.Text)
			break
		}
	}
}

// findWarehouseName — поиск названия склада по ID
func findWarehouseName(warehouses []wb.Warehouse, id int) string {
	for _, w := range warehouses {
		if w.ID == id {
			return w.Name
		}
	}
	return ""
}

// ===== Логика работы с уведомлениями =====

// getLastNotificationTime — получить время последней отправки уведомления
func getLastNotificationTime(userID int64, warehouseID int) int64 {
	if warehouses, ok := notifiedWarehouses[userID]; ok {
		return warehouses[warehouseID]
	}
	return 0
}

// markAsNotified — отметить время последнего уведомления
func markAsNotified(userID int64, warehouseID int, timestamp int64) {
	if _, ok := notifiedWarehouses[userID]; !ok {
		notifiedWarehouses[userID] = make(map[int]int64)
	}
	notifiedWarehouses[userID][warehouseID] = timestamp
}

// unmarkNotification — удалить запись об уведомлении
func unmarkNotification(userID int64, warehouseID int) {
	if warehouses, ok := notifiedWarehouses[userID]; ok {
		delete(warehouses, warehouseID)
	}
}
