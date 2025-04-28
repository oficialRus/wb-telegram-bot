package bot

import (
	"fmt"
	"log"

	"postavkinBot/internal/storage"
	"postavkinBot/internal/wb"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	WbClient *wb.Client
	Storage  *storage.Storage
)

func HandleStart(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	telegramID := update.Message.From.ID
	username := update.Message.From.UserName

	exists, err := Storage.UserExists(int64(telegramID))
	if err != nil {
		log.Printf("Ошибка проверки пользователя: %v", err)
		return
	}

	var text string
	if !exists {
		err := Storage.CreateUser(int64(telegramID), username)
		if err != nil {
			log.Printf("Ошибка создания пользователя: %v", err)
			return
		}
		text = "Вы зарегистрированы! 🎉 Добро пожаловать!"
	} else {
		text = "С возвращением! 👋"
	}

	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, text))
}

func HandleHelp(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	helpText := "📋 Доступные команды:\n" +
		"/start - Начало работы\n" +
		"/help - Помощь\n" +
		"/warehouses - Список всех складов\n" +
		"/addwarehouse - Добавить склад в отслеживание\n" +
		"/mywarehouses - Показать мои склады\n" +
		"/removewarehouse - Удалить склад из отслеживания\n" +
		"/setinterval - Установить интервал проверки лимитов"
	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, helpText))
}

func HandleWarehouses(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	warehouses, err := WbClient.GetWarehouses()
	if err != nil {
		log.Printf("Ошибка получения складов: %v", err)
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении складов. Попробуйте позже."))
		return
	}

	if len(warehouses) == 0 {
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Нет доступных складов."))
		return
	}

	const maxMessageSize = 4000
	text := "📦 Список доступных складов:\n"
	for _, w := range warehouses {
		line := fmt.Sprintf("- %s (ID: %d)\n", w.Name, w.ID)

		if len(text)+len(line) > maxMessageSize {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, text))
			text = ""
		}
		text += line
	}

	if text != "" {
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, text))
	}
}

func HandleAddWarehouse(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Введите ID склада, который хотите добавить в отслеживание:"))

	waitForUserInput(update, func(input string) {
		var warehouseID int
		if _, err := fmt.Sscanf(input, "%d", &warehouseID); err != nil {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка: введите корректный числовой ID склада."))
			return
		}

		if err := Storage.AddWarehouseToUser(update.Message.From.ID, warehouseID); err != nil {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при добавлении склада."))
			return
		}

		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("✅ Склад с ID %d успешно добавлен!", warehouseID)))
	})
}

func HandleMyWarehouses(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	telegramID := update.Message.From.ID

	warehouseIDs, err := Storage.GetUserWarehouses(int64(telegramID))
	if err != nil {
		log.Printf("Ошибка получения складов пользователя: %v", err)
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении ваших складов."))
		return
	}

	if len(warehouseIDs) == 0 {
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "У вас пока нет складов в отслеживании. Добавьте их через /addwarehouse."))
		return
	}

	allWarehouses, err := WbClient.GetWarehouses()
	if err != nil {
		log.Printf("Ошибка получения всех складов WB: %v", err)
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении списка складов. Попробуйте позже."))
		return
	}

	text := "📦 Ваши склады для отслеживания:\n"
	for _, id := range warehouseIDs {
		name := findWarehouseName(allWarehouses, id)
		if name == "" {
			name = fmt.Sprintf("Неизвестный склад (ID: %d)", id)
		}
		text += fmt.Sprintf("- %s (ID: %d)\n", name, id)
	}

	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, text))
}

func HandleRemoveWarehouse(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Введите ID склада, который хотите удалить из отслеживания:"))

	waitForUserInput(update, func(input string) {
		var warehouseID int
		if _, err := fmt.Sscanf(input, "%d", &warehouseID); err != nil {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка: введите корректный числовой ID склада."))
			return
		}

		if err := Storage.RemoveWarehouseFromUser(update.Message.From.ID, warehouseID); err != nil {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при удалении склада."))
			return
		}

		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("✅ Склад с ID %d успешно удалён из отслеживания!", warehouseID)))
	})
}

func HandleSetInterval(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Введите интервал проверки в минутах (например 5, 10, 15):"))

	waitForUserInput(update, func(input string) {
		var interval int
		if _, err := fmt.Sscanf(input, "%d", &interval); err != nil || interval <= 0 {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка: пожалуйста, введите положительное число."))
			return
		}

		if err := Storage.UpdateCheckInterval(update.Message.From.ID, interval); err != nil {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при сохранении интервала."))
			return
		}

		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("✅ Интервал обновлён! Теперь лимиты будут проверяться каждые %d минут.", interval)))
	})
}
