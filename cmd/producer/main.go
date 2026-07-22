package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"rabbitmq/internal/models"

	"github.com/rabbitmq/amqp091-go"
)

func gerarNumero8Digitos() string {
	return fmt.Sprintf("%08d", rand.Intn(100000000))
}

func main() {
	rand.Seed(time.Now().UnixNano())

	conn, _ := amqp091.Dial("amqp://guest:guest@localhost:5672/")
	defer conn.Close()

	ch, _ := conn.Channel()
	defer ch.Close()

	ch.ExchangeDeclare("barcode_logs", "fanout", true, false, false, false, nil)

	for i := 0; i < 10000; i++ {

		msg := models.Mensagem{
			Barcode:   gerarNumero8Digitos(),
			DataEnvio: time.Now().Format(time.RFC3339),
			Code:      "GT" + gerarNumero8Digitos(),
		}

		body, _ := json.Marshal(msg)

		ch.Publish("barcode_logs", "", false, false, amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		})

		log.Println("Mensagem enviada:", string(body))
	}
}
