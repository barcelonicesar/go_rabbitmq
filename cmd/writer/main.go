package main

import (
	"database/sql"
	"encoding/json"
	"log"

	"rabbitmq/internal/models"

	"github.com/rabbitmq/amqp091-go"
	_ "modernc.org/sqlite"
)

func main() {

	// Caminho correto quando rodando: go run ./cmd/writer
	db, err := sql.Open("sqlite", "internal/sqlite/barcodes.db")
	if err != nil {
		log.Fatal("Erro ao abrir banco:", err)
	}
	defer db.Close()

	// Cria as tabelas necessárias
	db.Exec(`CREATE TABLE IF NOT EXISTS barcodes_work1 (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        barcode TEXT,
        data_envio TEXT,
        code TEXT
    )`)

	db.Exec(`CREATE TABLE IF NOT EXISTS barcodes_work2 (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        barcode TEXT,
        data_envio TEXT,
        code TEXT
    )`)

	conn, err := amqp091.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal("Erro ao conectar ao RabbitMQ:", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Erro ao abrir canal:", err)
	}
	defer ch.Close()

	// Fila exclusiva do writer
	q, err := ch.QueueDeclare(
		"writer_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Erro ao declarar fila:", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Erro ao consumir fila:", err)
	}

	log.Println("Writer aguardando mensagens...")

	forever := make(chan bool)

	go func() {
		for d := range msgs {

			var msg models.Mensagem
			json.Unmarshal(d.Body, &msg)

			// Seleciona a tabela correta
			table := ""
			switch msg.Source {
			case "consumer1":
				table = "barcodes_work1"
			case "consumer2":
				table = "barcodes_work2"
			default:
				log.Println("Mensagem recebida sem Source válido:", msg)
				continue
			}

			_, err := db.Exec(
				"INSERT INTO "+table+" (barcode, data_envio, code) VALUES (?, ?, ?)",
				msg.Barcode, msg.DataEnvio, msg.Code,
			)
			if err != nil {
				log.Println("Erro ao inserir:", err)
			}
		}
	}()

	<-forever
}
