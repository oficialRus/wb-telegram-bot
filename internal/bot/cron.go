package bot

import (
	"fmt"
	"log"
	"strings"
	"time"

	"postavkinBot/internal/wb"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	notifiedWarehouses = make(map[int64]map[int]int64) // userID -> warehouseID -> timestamp
	checkInterval      = 15 * time.Second              // Проверка каждые 15 секунд
	repeatNotifyDelay  = time.Minute                   // Повторное уведомление через 1 минуту
	cachedWarehouses   []wb.Warehouse                  // Кэш складов
)

// StartCronJob — запуск задач проверки
func StartCronJob(bot *tgbotapi.BotAPI) {
	var err error
	cachedWarehouses, err = WbClient.GetWarehouses()
	if err != nil {
		log.Fatalf("Ошибка получения списка складов при старте: %v", err)
	}

	go func() {
		for {
			log.Println("[CRON] Проверка складов по кэшу...")
			checkWarehouses(bot)
			time.Sleep(checkInterval)
		}
	}()
}

// SetCheckInterval — установить интервал проверки
func SetCheckInterval(seconds int) {
	if seconds <= 0 {
		seconds = 1
	}
	checkInterval = time.Duration(seconds) * time.Second
	log.Printf("Интервал проверки изменён на каждые %d секунд\n", seconds)
}

// checkWarehouses — проверка лимитов всех пользователей
func checkWarehouses(bot *tgbotapi.BotAPI) {
	users, err := Storage.GetAllUsers()
	if err != nil {
		log.Printf("Ошибка получения пользователей: %v", err)
		return
	}

	coefficients, err := WbClient.GetAcceptanceCoefficients()
	if err != nil {
		if isTooManyRequestsError(err) {
			log.Println("[WARN] Слишком много запросов (429), делаем паузу 1 минуту...")
			time.Sleep(1 * time.Minute)
		} else {
			log.Printf("Ошибка получения коэффициентов приёмки: %v", err)
		}
		return
	}

	for _, user := range users {
		checkUserWarehouses(bot, user.TelegramID, cachedWarehouses, coefficients)
	}
}

// checkUserWarehouses — проверка складов одного пользователя
func checkUserWarehouses(bot *tgbotapi.BotAPI, telegramID int64, allWarehouses []wb.Warehouse, coefficients []wb.Coefficient) {
	warehouseIDs, err := Storage.GetUserWarehouses(telegramID)
	if err != nil {
		log.Printf("Ошибка получения складов пользователя %d: %v", telegramID, err)
		return
	}

	now := time.Now().Unix()

	for _, id := range warehouseIDs {
		coefficient := findCoefficient(coefficients, id)
		if coefficient != nil && (coefficient.Coefficient == 0 || coefficient.Coefficient == 1) && coefficient.AllowUnload {
			lastNotified := getLastNotificationTime(telegramID, id)

			if lastNotified == 0 || now-lastNotified >= int64(repeatNotifyDelay.Seconds()) {
				text := fmt.Sprintf(
					"📦 Лимит на складе: %s (ID: %d)\n📈 Коэффициент: %d\n🗓 Дата: %s",
					coefficient.WarehouseName,
					coefficient.WarehouseID,
					coefficient.Coefficient,
					coefficient.Date,
				)

				msg := tgbotapi.NewMessage(telegramID, text)
				if _, err := bot.Send(msg); err != nil {
					log.Printf("Ошибка отправки сообщения пользователю %d: %v", telegramID, err)
				}
				markAsNotified(telegramID, id, now)
			}
		} else {
			unmarkNotification(telegramID, id)
		}
	}
}

// findCoefficient — найти коэффициент по ID склада
func findCoefficient(coefficients []wb.Coefficient, warehouseID int) *wb.Coefficient {
	for _, c := range coefficients {
		if c.WarehouseID == warehouseID {
			return &c
		}
	}
	return nil
}

// isTooManyRequestsError — проверка на ошибку 429
func isTooManyRequestsError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "429 Too Many Requests")
}
