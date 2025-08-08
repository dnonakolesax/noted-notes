package mq

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MQChan struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

type MQProducer struct {
	queue amqp.Queue
	*MQChan
}

type MQConsumer struct {
	consumeChan <-chan amqp.Delivery
	*MQChan
}

func NewMqChan() (*MQChan, error) {
	conn, err := amqp.Dial("amqp://guest:guest@172.26.0.2:5672/")
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()

	if err != nil {
		return nil, err
	}
	return &MQChan{conn: conn, channel: ch}, nil
}

func (mc *MQChan) Close() error {
	err := mc.conn.Close()
	if err != nil {
		return err
	}
	err = mc.channel.Close()
	if err != nil {
		return err
	}
	return nil
}

func (mc *MQChan) NewProducer(queueName string) (*MQProducer, error) {
	q, err := mc.channel.QueueDeclare(
		queueName, // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)

	if err != nil {
		return nil, err
	}

	return &MQProducer{
		queue: q,
		MQChan: mc,
	}, nil
}

func (mp *MQProducer) Publish(msg []byte) error {
	err := mp.channel.PublishWithContext(context.TODO(),
		"",     // exchange
		mp.queue.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msg,
		})
	if err != nil {
		return err
	}	
	return nil
}

func (mc *MQChan) NewConsumer(queueName string) (*MQConsumer, error) {
	consumeChan, err := mc.channel.Consume(
		queueName, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return nil, err
	}
	return &MQConsumer{
		consumeChan: consumeChan,
		MQChan: mc,
	}, nil
}

func (mc *MQConsumer) Consume() (<-chan []byte) {
	byteChan := make(chan []byte)
	go func() {	
		for message := range mc.consumeChan {
			byteChan <- message.Body
		}
	}()
	return byteChan
}
