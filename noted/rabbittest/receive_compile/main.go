package main

import (
	"log"
	"context"
	"os"
	"encoding/json"
	"os/exec"

	amqp "github.com/rabbitmq/amqp091-go"
)

type codeFile struct {
	Name string `json:"filename"`
	Data string `json:"code"`
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@172.26.0.2:5672/")
	if err != nil {
		log.Fatalf("unable to open connect to RabbitMQ server. Error: %s", err)
	}

	defer func() {
		_ = conn.Close() // Закрываем подключение в случае удачной попытки подключения
	}()
	
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open a channel. Error: %s", err)
	}

	defer func() {
		_ = ch.Close() // Закрываем подключение в случае удачной попытки подключения
	}()

	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		log.Fatalf("failed to declare a queue. Error: %s", err)
	}

	q2, err := ch.QueueDeclare(
		"prepare", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)

	messages, err := ch.Consume(
		q2.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("failed to register a consumer. Error: %s", err)
	}

	var forever chan struct{}

	go func() {
		for message := range messages {

			var code codeFile
			err = json.Unmarshal(message.Body, &code)

			if err != nil  {
				log.Fatalf("failed to unmarshal. Error: %s", err)
				continue
			}

			d1 := []byte(code.Data)
			f, err := os.Create("/tmp/kernel228/" + code.Name)

			if err != nil  {
				log.Fatalf("failed to unmarshal. Error: %s", err)
				continue
			}

			_, _ = f.Write(d1)
			f.Close()

			cmd := exec.Command("python3", "/tmp/kernel228/blockparser.py", code.Name)
			err = cmd.Run()

			if err != nil  {
				log.Printf("Python cmd error: %s", err)
				continue
			}

			err = ch.PublishWithContext(context.Background(),
				"",     // exchange
				q.Name, // routing key
				false,  // mandatory
				false,  // immediate
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(code.Name),
				})
			if err != nil {
				log.Fatalf("failed to publish a message. Error: %s", err)
				continue
			}
			log.Printf(" [x] Sent %s\n", code.Name)
		}
	}()

	log.Printf(" [*] Waiting for messages from prepare. To exit press CTRL+C")
	<-forever
}