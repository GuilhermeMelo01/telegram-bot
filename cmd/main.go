package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	telegramBotAPI "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "telegram"
)

var (

	// Menu texts
	firstMenu  = "<b>Menu 1</b>\n\nA beautiful menu with a shiny inline button."
	secondMenu = "<b>Menu 2</b>\n\nA better menu with even more shiny inline buttons."

	// Button texts
	nextButton     = "Next"
	backButton     = "Back"
	tutorialButton = "Tutorial"

	// Store bot screaming status
	screaming = false
	bot       *telegramBotAPI.BotAPI

	// Keyboard layout for the first menu. One button, one row
	firstMenuMarkup = telegramBotAPI.NewInlineKeyboardMarkup(
		telegramBotAPI.NewInlineKeyboardRow(
			telegramBotAPI.NewInlineKeyboardButtonData(nextButton, nextButton),
		),
	)

	// Keyboard layout for the second menu. Two buttons, one per row
	secondMenuMarkup = telegramBotAPI.NewInlineKeyboardMarkup(
		telegramBotAPI.NewInlineKeyboardRow(
			telegramBotAPI.NewInlineKeyboardButtonData(backButton, backButton),
		),
		telegramBotAPI.NewInlineKeyboardRow(
			telegramBotAPI.NewInlineKeyboardButtonURL(tutorialButton, "https://core.telegram.org/bots/api"),
		),
	)
)

func init() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Print("No .env file found")
	}

}

func main() {
	telegramBotToken := os.Getenv("BOT_TOKEN")
	bot, err := telegramBotAPI.NewBotAPI(telegramBotToken)

	if err != nil {
		fmt.Println("error in bot creation")
		log.Panic(err)
	}

	chatID := getChatIdListFromBot(telegramBotToken)

	if chatID == 0 {
		log.Panic("chatID has no value")
	}

	ticker := time.NewTicker(3 * time.Second)
	for range ticker.C {
		msg := telegramBotAPI.NewMessage(chatID, "Seu plano mensal está perto do vencimento!")

		_, err = bot.Send(msg)
		if err != nil {
			fmt.Println("error in bot send")
			log.Panic(err)
		}
	}

	// time.Sleep(5 * time.Second)

	// psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
	// 	host, port, user, password, dbname)

	// db, err := sql.Open("postgres", psqlInfo)

	// if err != nil {
	// 	panic(err)
	// }

	// defer db.Close()

	// err = db.Ping()
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println("Successfully connected")

	// var err error
	// bot, err = tgbotapi.NewBotAPI("6713538459:AAEf__EIUb15ZfOi1Cmir1vV9qtI8VUM3Mw")
	// if err != nil {
	// 	// Abort if something is wrong
	// 	log.Panic(err)
	// }

	// // Set this to true to log all interactions with telegram servers
	// bot.Debug = false

	// u := tgbotapi.NewUpdate(0)
	// u.Timeout = 60

	// // Create a new cancellable background context. Calling `cancel()` leads to the cancellation of the context
	// ctx := context.Background()
	// ctx, cancel := context.WithCancel(ctx)

	// // `updates` is a golang channel which receives telegram updates
	// updates := bot.GetUpdatesChan(u)

	// // Pass cancellable context to goroutine
	// go receiveUpdates(ctx, updates)

	// // Tell the user the bot is online
	// log.Println("Start listening for updates. Press enter to stop")

	// // Wait for a newline symbol, then cancel handling updates
	// bufio.NewReader(os.Stdin).ReadBytes('\n')
	// cancel()

}

func getChatIdListFromBot(telegramBotToken string) int64 {

	apiUrl := fmt.Sprintf("%s%s%s", "https://api.telegram.org/bot", telegramBotToken, "/getUpdates")

	request, err := http.NewRequest("GET", apiUrl, nil)

	if err != nil {
		fmt.Println(err)
	}

	request.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		fmt.Println(err)
	}

	responseBody, err := io.ReadAll(response.Body)

	if err != nil {
		fmt.Println(err)
	}

	var data map[string]interface{}

	if err := json.Unmarshal(responseBody, &data); err != nil {
		fmt.Println("Error unmarshalling JSON:", err)

	}

	// Check if "result" key exists in the response
	if updates, ok := data["result"].([]interface{}); ok {
		// Iterate through updates
		for _, update := range updates {
			if message, ok := update.(map[string]interface{})["message"].(map[string]interface{}); ok {
				if chat, ok := message["chat"].(map[string]interface{}); ok {
					if chatID, ok := chat["id"].(float64); ok {
						return int64(chatID)
					}
				}
			}
		}
	}

	return 0
}

func formatJSON(data []byte) string {
	var out bytes.Buffer
	err := json.Indent(&out, data, "", " ")
	if err != nil {
		fmt.Println(err)
	}

	d := out.Bytes()
	return string(d)
}

func receiveUpdates(ctx context.Context, updates telegramBotAPI.UpdatesChannel) {
	// `for {` means the loop is infinite until we manually stop it
	for {
		select {
		// stop looping if ctx is cancelled
		case <-ctx.Done():
			return
		// receive update from channel and then handle it
		case update := <-updates:
			handleUpdate(update)
		}
	}
}

func handleUpdate(update telegramBotAPI.Update) {
	switch {
	// Handle messages
	case update.Message != nil:
		handleMessage(update.Message)
		break

	// Handle button clicks
	case update.CallbackQuery != nil:
		handleButton(update.CallbackQuery)
		break
	}
}

func handleMessage(message *telegramBotAPI.Message) {
	user := message.From
	text := message.Text

	if user == nil {
		return
	}

	// Print to console
	log.Printf("%s wrote %s", user.FirstName, text)

	var err error
	if strings.HasPrefix(text, "/") {
		err = handleCommand(message.Chat.ID, text)
	} else if screaming && len(text) > 0 {
		msg := telegramBotAPI.NewMessage(message.Chat.ID, strings.ToUpper(text))
		// To preserve markdown, we attach entities (bold, italic..)
		msg.Entities = message.Entities
		_, err = bot.Send(msg)
	} else {
		// This is equivalent to forwarding, without the sender's name
		copyMsg := telegramBotAPI.NewCopyMessage(message.Chat.ID, message.Chat.ID, message.MessageID)
		_, err = bot.CopyMessage(copyMsg)
	}

	if err != nil {
		log.Printf("An error occured: %s", err.Error())
	}
}

// When we get a command, we react accordingly
func handleCommand(chatId int64, command string) error {
	var err error

	switch command {
	case "/scream":
		screaming = true
		break

	case "/whisper":
		screaming = false
		break

	case "/menu":
		err = sendMenu(chatId)
		break
	}

	return err
}

func handleButton(query *telegramBotAPI.CallbackQuery) {
	var text string

	markup := telegramBotAPI.NewInlineKeyboardMarkup()
	message := query.Message

	if query.Data == nextButton {
		text = secondMenu
		markup = secondMenuMarkup
	} else if query.Data == backButton {
		text = firstMenu
		markup = firstMenuMarkup
	}

	callbackCfg := telegramBotAPI.NewCallback(query.ID, "")
	bot.Send(callbackCfg)

	// Replace menu text and keyboard
	msg := telegramBotAPI.NewEditMessageTextAndMarkup(message.Chat.ID, message.MessageID, text, markup)
	msg.ParseMode = telegramBotAPI.ModeHTML
	bot.Send(msg)
}

func sendMenu(chatId int64) error {
	msg := telegramBotAPI.NewMessage(chatId, firstMenu)
	msg.ParseMode = telegramBotAPI.ModeHTML
	msg.ReplyMarkup = firstMenuMarkup
	_, err := bot.Send(msg)
	return err
}
