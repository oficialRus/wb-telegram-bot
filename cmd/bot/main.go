package main

import (
	"log"
	"os"

	"postavkinBot/internal/storage" // подключаем наше хранилище

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Загружаем .env файл
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	// Получаем токен бота из переменной окружения
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN не установлен")
	}

	// Инициализация бота
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true // Для дебага

	log.Printf("Бот запущен как: %s", bot.Self.UserName)

	// Подключение к базе данных
	dbPath := "data.db" // Имя файла базы данных
	storage, err := storage.NewStorage(dbPath)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			switch update.Message.Command() {
			case "start":
				handleStart(bot, update, storage)
			case "help":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Доступные команды:\n/start - Начало работы\n/help - Помощь")
				bot.Send(msg)
			default:
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда. Введите /help для списка доступных команд.")
				bot.Send(msg)
			}
		}
	}
}

func handleStart(bot *tgbotapi.BotAPI, update tgbotapi.Update, storage *storage.Storage) {
	telegramID := update.Message.From.ID
	username := update.Message.From.UserName

	exists, err := storage.UserExists(int64(telegramID))
	if err != nil {
		log.Printf("Ошибка проверки пользователя: %v", err)
		return
	}

	if !exists {
		err := storage.CreateUser(int64(telegramID), username)
		if err != nil {
			log.Printf("Ошибка создания пользователя: %v", err)
			return
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы зарегистрированы! 🎉 Добро пожаловать!")
		bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "С возвращением! 👋")
		bot.Send(msg)
	}
}
