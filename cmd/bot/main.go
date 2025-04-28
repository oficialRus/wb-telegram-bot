package main

import (
	"log"
	"os"

	"postavkinBot/internal/bot"
	"postavkinBot/internal/storage"
	"postavkinBot/internal/wb"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

// Глобальный канал апдейтов
var updatesChan = make(chan tgbotapi.Update)

func main() {
	// Загружаем .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	// Инициализация токена бота
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN не установлен")
	}

	// Инициализация бота
	tgBot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}
	tgBot.Debug = true

	log.Printf("Бот запущен как: %s", tgBot.Self.UserName)

	// Инициализация БД
	dbPath := "data.db"
	storageInstance, err := storage.NewStorage(dbPath)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}

	// Инициализация WB клиента
	wbClient := wb.NewClient()

	// Связываем пакеты bot -> storage и wb
	bot.Storage = storageInstance
	bot.WbClient = wbClient
	bot.UpdatesChan = updatesChan // << добавляем канал апдейтов в пакет bot

	// Старт планировщика
	bot.StartCronJob(tgBot)

	// Получение апдейтов
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := tgBot.GetUpdatesChan(u)

	// Параллельно закидываем апдейты в канал
	go func() {
		for update := range updates {
			updatesChan <- update
		}
	}()

	// Обработка апдейтов
	for update := range updatesChan {
		if update.Message == nil {
			continue
		}

		switch update.Message.Command() {
		case "start":
			bot.HandleStart(tgBot, update)
		case "help":
			bot.HandleHelp(tgBot, update)
		case "warehouses":
			bot.HandleWarehouses(tgBot, update)
		case "addwarehouse":
			bot.HandleAddWarehouse(tgBot, update)
		case "mywarehouses":
			bot.HandleMyWarehouses(tgBot, update)
		case "removewarehouse":
			bot.HandleRemoveWarehouse(tgBot, update)
		case "setinterval":
			bot.HandleSetInterval(tgBot, update)
		default:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда. Введите /help для списка доступных команд.")
			tgBot.Send(msg)
		}
	}
}
