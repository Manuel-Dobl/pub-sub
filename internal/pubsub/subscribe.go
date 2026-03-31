package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int

const (
	Ack AckType = iota
	NackDiscard
	NackRequeue
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) error {

	ch, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		log.Fatal(err)
	}

	messages, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Printf("Consume error: %v", err)
	}

	go func() {
		for msg := range messages {
			var target T
			err := json.Unmarshal(msg.Body, &target)
			if err != nil {
				log.Printf("Unmarshal error: %v", err)
				continue
			}
			ackType := handler(target)

			switch ackType {
			case Ack:
				err := msg.Ack(false)
				if err != nil {
					log.Printf("ack error: %v", err)
					break
				}
				log.Printf("Ack: message processed successfully")

			case NackRequeue:
				err := msg.Nack(false, true)
				if err != nil {
					log.Printf("Nack error: %v", err)
					break
				}
				log.Printf("NackRequeue: message will be retried")

			case NackDiscard:
				err := msg.Nack(false, false)
				if err != nil {
					log.Printf("Nack error: %v", err)
					break
				}
				log.Printf("NackDiscard: message discarded")

			default:
				log.Printf("default ack nack")

			}

		}

	}()
	return nil
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) error {

	ch, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		log.Fatal(err)
	}
	ch.Qos(10, 0, false)
	messages, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Printf("Consume error: %v", err)
	}

	go func() {
		for msg := range messages {
			var target T
			buf := bytes.NewBuffer(msg.Body)
			decoder := gob.NewDecoder(buf)
			err := decoder.Decode(&target)
			if err != nil {
				log.Printf("Decode error: %v", err)
				continue
			}
			ackType := handler(target)

			switch ackType {
			case Ack:
				err := msg.Ack(false)
				if err != nil {
					log.Printf("ack error: %v", err)
					break
				}
				log.Printf("Ack: message processed successfully")

			case NackRequeue:
				err := msg.Nack(false, true)
				if err != nil {
					log.Printf("Nack error: %v", err)
					break
				}
				log.Printf("NackRequeue: message will be retried")

			case NackDiscard:
				err := msg.Nack(false, false)
				if err != nil {
					log.Printf("Nack error: %v", err)
					break
				}
				log.Printf("NackDiscard: message discarded")

			default:
				log.Printf("default ack nack")

			}

		}

	}()
	return nil
}
