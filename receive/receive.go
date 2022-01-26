package main

import (
	"encoding/json"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

type sender struct {
	Text string `json:"text"`
	ID   int    `json:"id"`
}

func main() {
	bot, err := tgbotapi.NewBotAPI("1251601996:AAGXiUfRVsRfjXdqaBlipY9cd8VvuDAKxm0")
	if err != nil {
		log.Panic(err)
	}

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var txt sender
			log.Printf("Received a message: %s", d.Body)

			json.Unmarshal(d.Body, &txt)

			message := tgbotapi.NewMessage(109435993, fmt.Sprintf("%s - %d", txt.Text, txt.ID))
			if _, err := bot.Send(message); err != nil {
				log.Println(err)
			}
		}

	}()
	//------------------------------------------------------------------
	// u := tgbotapi.NewUpdate(0)
	// u.Timeout = 60

	// updates := bot.GetUpdatesChan(u)
	// for update := range updates {
	// 	if update.Message != nil { // If we got a message
	// 		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

	// 		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
	// 		msg.ReplyToMessageID = update.Message.MessageID

	// 		bot.Send(msgs)
	// 	}
	// }

	//-------------------------------------------------------------------
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
