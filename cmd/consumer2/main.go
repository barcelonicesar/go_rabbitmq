package main

import (
	"encoding/json"
	"log"

	"rabbitmq/internal/models"

	"github.com/rabbitmq/amqp091-go"
)

func main() {

	conn, _ := amqp091.Dial("amqp://guest:guest@localhost:5672/")
	defer conn.Close()

	ch, _ := conn.Channel()
	defer ch.Close()

	q, _ := ch.QueueDeclare("barcode_work2", false, false, false, false, nil)
	ch.QueueBind(q.Name, "", "barcode_logs", false, nil)

	msgs, _ := ch.Consume(q.Name, "", true, false, false, false, nil)

	log.Println("Consumer 2 aguardando mensagens...")

	forever := make(chan bool)

	go func() {
		for d := range msgs {

			var msg models.Mensagem
			json.Unmarshal(d.Body, &msg)

			msg.Source = "consumer2"

			body, _ := json.Marshal(msg)

			ch.Publish(
				"",
				"writer_queue",
				false,
				false,
				amqp091.Publishing{
					ContentType: "application/json",
					Body:        body,
				},
			)
		}
	}()

	<-forever
}
