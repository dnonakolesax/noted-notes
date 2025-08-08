package main

import (
	"fmt"
	"log"
	"plugin"

	amqp "github.com/rabbitmq/amqp091-go"
)


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

	messages, err := ch.Consume(
		q.Name, // queue
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

	vars := make(map[string]any, 10)
	funcs := make(map[string]any, 10)

	go func() {
		for message := range messages {
			log.Printf("received a message: %s", message.Body)
			pluginName := string(message.Body) + ".so"
			p, err := plugin.Open(pluginName)
			if err != nil {
				fmt.Println("pizdec 1")
				fmt.Println(err.Error())
				continue
			}

			expectedName := "Export_" + string(message.Body)
			v, err := p.Lookup(expectedName)
			if err != nil {
				fmt.Println("pizdec 2")
				fmt.Println(err.Error())
				continue
			}
			v.(func(*map[string]any, *map[string]any))(&vars, &funcs)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}