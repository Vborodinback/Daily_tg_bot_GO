package main

import (
	//"fmt"
	"fmt"
	"log"
	"time"

	//"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Структура для хранения данных о чате
type ChatData struct {
	Id   int64
	Name string
}

// статус бота
var botStatus = make(map[int64]string)
var user []string
var dailyTime time.Time

func main() {
	// Замени на свой токен
	bot, err := tgbotapi.NewBotAPI("YOUR_TOKEN")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Авторизован как %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for {
		select {
		case update := <-updates:
			if update.Message != nil {
				handleMessage(bot, update.Message)
				fmt.Println(update.Message.From.UserName)
			} else if update.CallbackQuery != nil {
				handleCallbackQuery(bot, update.CallbackQuery)
			}
		}
	}
}

// Обработка сообщения
func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	fmt.Println(message.From.UserName)

	// Проверка на наличие бота в чате
	if message.NewChatMembers != nil {
		contains := false

		for _, item := range message.NewChatMembers {
			if item.ID == bot.Self.ID {
				contains = true
			}
		}

		if !contains {
			return
		}

		var chat ChatData

		chat.Id = chatID
		chat.Name = message.Chat.Title

		msgText := "Я запомнил Ваш чат, чтобы добавить ведущих напиши [мне](https://t.me/" + "YOUTELEGRAM_USERNAME" + ") в личные сообщения"
		msg := tgbotapi.NewMessage(chatID, msgText)
		msg.ParseMode = "Markdown"
		bot.Send(msg)

		showGroupMenu(bot, chatID)
	} else if message.Chat.Type == "private" && message.From.ID != bot.Self.ID { // Проверка на наличие бота в личных сообщениях
		if message.IsCommand() && message.Command() == "start" {
			showStartMenu(bot, chatID)
			botStatus[chatID] = "start"
			fmt.Println("kommand menu")
		}

		switch botStatus[chatID] { // Проверка статуса бота
		case "":
			msg := tgbotapi.NewMessage(chatID, "Привет! Для начала работы напиши /start")
			bot.Send(msg)
		// case "start":
		// 	showStartMenu(bot, chatID)
		// 	fmt.Println("start menu")
		case "waiting_for_users":
			entities := message.Entities
			if entities == nil || len(entities) > 1 || entities[0].Type != "mention" {
				msg := tgbotapi.NewMessage(chatID, "⚠️ Вы некорректно ввели пользователя\nДля добавления участников пишите их через @username")
				bot.Send(msg)
				showAddUsersMenu(bot, chatID)
			}
			newUser := extractMention(message.Text, message.Entities)
			checkUser := containsUser(user, newUser)
			if newUser != "" && checkUser == false {
				user = append(user, newUser)
				msg := tgbotapi.NewMessage(chatID, "Пользователь "+newUser+" добавлен")
				bot.Send(msg)
				showAddUsersMenu(bot, chatID)
			} else if newUser == "" {
				msg := tgbotapi.NewMessage(chatID, "Для добавления участников пишите их через @username")
				bot.Send(msg)
				showAddUsersMenu(bot, chatID)
			} else if checkUser == true {
				msg := tgbotapi.NewMessage(chatID, "⚠️ Пользователь "+newUser+" уже добавлен")
				bot.Send(msg)	
				showAddUsersMenu(bot, chatID)
			}
		case "users_added":
			// if len(user) != 0 {
			// 	msg := tgbotapi.NewMessage(chatID, "Напишите время проведения дейли в формате HH:MM")
			// 	bot.Send(msg)
			// 	botStatus[chatID] = "waiting_for_time"
			// } else {
			// 	msg := tgbotapi.NewMessage(chatID, "Вы не добавили участников")
			// 	bot.Send(msg)
			// 	showStartMenu(bot, chatID)
			// }
		case "waiting_for_time":
			dailyTime, err := time.Parse("15:04", message.Text)
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "⚠️ Вы некорректно ввели время\nНапишите время проведения дейли в формате HH:MM")
				bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(chatID, "Время проведения дейли: "+dailyTime.Format("15:04"))
				bot.Send(msg)
				showConfirmMenu(bot, chatID, dailyTime, user)
			}
		}
	}
}

func containsUser(user []string, username string) bool {
	for _, item := range user {
		if item == username {
			return true
		}
	}
	return false
}

func showConfirmMenu(bot *tgbotapi.BotAPI, chatID int64, dailyTime time.Time, user []string) {
	msg := tgbotapi.NewMessage(chatID, "Все верно?\nВремя проведения дейли: "+dailyTime.Format("15:04")+"\nУчастники: "+strings.Join(user, ", "))
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Да", "confirm_daily"),
			tgbotapi.NewInlineKeyboardButtonData("Нет, начать добавление заново", "create_daily"),
		),
	)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

// Извлечение упоминания
func extractMention(text string, entities []tgbotapi.MessageEntity) string {
	if entities[0].Type == "mention" {
		return text[entities[0].Offset : entities[0].Offset+entities[0].Length]
	}
	return ""
}

// Обработка callback-запроса
func handleCallbackQuery(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {
	chatID := callbackQuery.Message.Chat.ID
	switch callbackQuery.Data {
	case "get_host_daily":
		//chatID := callbackQuery.Message.Chat.ID
		//select from users where chat_id = chatID
		//if users != nil{...}
		//else{...}
		msgText := "Это все что мне удалось найти..."
		msg := tgbotapi.NewMessage(chatID, msgText)
		bot.Send(msg)
		showGroupMenu(bot, chatID)
	case "create_daily":
		botStatus[chatID] = "waiting_for_users"
		showAddUsersMenu(bot, chatID)
	case "add_more_users":
		botStatus[chatID] = "waiting_for_users"
		showAddUsersMenu(bot, chatID)
	case "cancel_add_users":
		if len(user) != 0 {
			botStatus[chatID] = "waiting_for_time"
			showAddTimeMenu(bot, chatID)
		} else {
			msg := tgbotapi.NewMessage(chatID, "⚠️ Вы не добавили участников")
			bot.Send(msg)
			showStartMenu(bot, chatID)
		}
	//botStatus[chatID] = "users_added"
	case "confirm_daily":
		msg := tgbotapi.NewMessage(chatID, "Спасибо за добавление дейли!")
		bot.Send(msg)
		user = []string{}
		dailyTime = time.Time{}
		botStatus[chatID] = "start"
	case "cancel_add_time":
		botStatus[chatID] = "start"
		user = []string{}
		showStartMenu(bot, chatID)
	}
}

// Обработка стартового меню
func showStartMenu(bot *tgbotapi.BotAPI, chatID int64) {
	botStatus[chatID] = "start"
	msg := tgbotapi.NewMessage(chatID, "Что хочешь сделать?")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Создать график проведения дейли", "create_daily"),
			tgbotapi.NewInlineKeyboardButtonData("Удалить график дейли", "delete_daily"),
		),
	)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

// Обработка меню добавления участников
// func showMoreUsersMenu(bot *tgbotapi.BotAPI, chatID int64) {
// 	msg := tgbotapi.NewMessage(chatID, "Хотите добавить еще пользователя?")
// 	keyboard := tgbotapi.NewInlineKeyboardMarkup(
// 		tgbotapi.NewInlineKeyboardRow(
// 			tgbotapi.NewInlineKeyboardButtonData("Да", "add_more_users"),
// 			tgbotapi.NewInlineKeyboardButtonData("Нет, завершить добавление", "cancel_add_users"),
// 		),
// 	)
// 	msg.ReplyMarkup = keyboard
// 	bot.Send(msg)
// }

// Обработка группового меню
func showGroupMenu(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Что хочешь сделать?")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Узнать порядок ведущих", "get_host_daily"),
		),
	)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func showAddUsersMenu(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Добавьте участника, упомянув его (@username), или нажмите 'Завершить' для завершения добавления. Участники добавляются по одному")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Завершить", "cancel_add_users"),
		),
	)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func showAddTimeMenu(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Напишите время проведения дейли в формате HH:MM")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Отменить создание дейли", "cancel_add_time"),
		),
	)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}