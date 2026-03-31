package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	gamelogic.PrintServerHelp()
	connString := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connString)
	if err != nil {
		log.Fatal(err)
	}
	//close this connection later
	defer conn.Close()

	//open up a channel inside that connection
	publishCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("Could not create channel: %v", err)
	}

	fmt.Println("Connection big success")

	//declare a queue
	routingKey := routing.GameLogSlug + ".*"
	err = pubsub.SubscribeGob(conn, routing.ExchangePerilTopic, routing.GameLogSlug, routingKey, pubsub.SimpleQueueDurable, handlerLogs())
	if err != nil {
		log.Fatal(err)
	}

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		word := words[0]

		if word == "pause" {
			fmt.Println("Sending pause message")
			pubsub.PublishJSON(publishCh, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: true,
			})
		} else if word == "resume" {
			fmt.Println("Sending resume message")
			pubsub.PublishJSON(publishCh, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: false,
			})
		} else if word == "quit" {
			fmt.Println("Quitting")
			return
		} else {
			fmt.Println("Invalid Command")
		}

	}

}
