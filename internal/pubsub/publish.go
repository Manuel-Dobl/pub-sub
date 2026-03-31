package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	SimpleQueueDurable SimpleQueueType = iota
	SimpleQueueTransient
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	//Marshal the val to JSON bytes
	jsonBytes, err := json.Marshal(val)
	if err != nil {
		fmt.Printf("error marshalling json: %v", err)
		return err
	}

	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        jsonBytes,
	})
	if err != nil {
		fmt.Printf("error publish: %v", err)
		return err
	}
	return nil
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	//encode to gob
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(val)
	if err != nil {
		fmt.Printf("error encoding gob: %v", err)
		return err
	}

	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/gob",
		Body:        buf.Bytes(),
	})
	if err != nil {
		fmt.Printf("error publish: %v", err)
		return err
	}
	return nil
}

func DeclareAndBind(conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // SimpleQueueType is an "enum" type I made to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {

	// create a channel
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not create channel: %v", err)
	}
	//declare a queue

	q, err := ch.QueueDeclare(
		queueName,                       //queue Name
		queueType == SimpleQueueDurable, // durable
		queueType != SimpleQueueDurable, // auto-delete
		queueType != SimpleQueueDurable, // exclusive
		false,                           // noWait
		amqp.Table{
			"x-dead-letter-exchange": "peril_dlx",
		},
		//args
	)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not declare queue: %v", err)
	}
	//bind queue

	err = ch.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		log.Fatalf("Could not bind queue: %v", err)
	}
	return ch, q, nil
}
