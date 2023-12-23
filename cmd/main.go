package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	// "os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	// telegramBotAPI "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

type Usuario struct {
	ID        uint
	Nome      string
	Sobrenome string
	Email     string
	Cpf       string
	Idade     uint
	CreateAt  time.Time
	UpdateAt  time.Time
	PlanoID   uint
	Plano     *Plano
}

type Plano struct {
	ID         uint
	Tipo       string
	Nome       string
	Descricao  string
	Valor      float64
	CreateAt   time.Time
	UpdateAt   time.Time
	Datainicio time.Time
	Datafim    time.Time
	Usuarios   []*Usuario
}

type Users struct {
	ID   uint
	nome string
}

func init() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Print("No .env file found")
	}

}

func main() {
	dsn := "host=localhost user=postgres password=postgres dbname=telegram port=5432 sslmode=disable TimeZone=America/Sao_Paulo"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Panic(err)
	}

	plano := Plano{Tipo: "Básio", Nome: "Plano Básico", Descricao: "Plano para pobres", Valor: 29.99,
		CreateAt: time.Now(), UpdateAt: time.Now(), Datainicio: time.Now(), Datafim: time.Now()}

	result := db.Omit(clause.Associations).Create(&plano)

	if result.Error != nil {
		log.Panic(result.Error)
	}

	fmt.Println(result.RowsAffected)
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
