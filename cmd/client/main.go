package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	connString := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connString)
	if err != nil {
		log.Fatal(err)
	}
	//close this connection later
	defer conn.Close()

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}

	//gamestate
	gameState := gamelogic.NewGameState(username)

	//names
	queueName := routing.PauseKey + "." + username
	armyMovesQueueName := routing.ArmyMovesPrefix + "." + gameState.GetUsername()
	armyMovesKeyName := routing.ArmyMovesPrefix + ".*"

	// pause subscription consume
	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, queueName, routing.PauseKey, pubsub.SimpleQueueTransient, handlerPause(gameState))
	if err != nil {
		log.Fatal(err)
	}
	//create a new channel for publishing
	publishCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}
	//subscribe to army moves
	pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, armyMovesQueueName, armyMovesKeyName, pubsub.SimpleQueueTransient, handlerMove(gameState, publishCh))
	//SUBSCRIBNE TO WAR
	pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix, routing.WarRecognitionsPrefix+".*", pubsub.SimpleQueueDurable, handlerWar(gameState, publishCh))

	//repl

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		word := words[0]
		switch word {
		case "spawn":
			err := gameState.CommandSpawn(words)
			if err != nil {
				fmt.Println(err)
			}
		case "move":
			move, err := gameState.CommandMove(words)
			if err != nil {
				fmt.Println(err)
				continue
			}
			armyMovesPublishName := routing.ArmyMovesPrefix + "." + move.Player.Username
			err = pubsub.PublishJSON(publishCh, routing.ExchangePerilTopic, armyMovesPublishName, move)
			if err != nil {
				fmt.Printf("error: %s\n", err)
				continue
			}
			fmt.Println("Move succesfull")

		case "status":
			gameState.CommandStatus()

		case "help":
			gamelogic.PrintClientHelp()

		case "spam":
			if len(words) < 2 {
				fmt.Printf("need a second word")
				continue
			}

			n, err := strconv.Atoi(words[1])
			if err != nil {
				fmt.Printf("Invalid Number")
			}
			for i := 0; i < n; i++ {
				msg := gamelogic.GetMaliciousLog()
				err := publishGameLog(publishCh, username, msg)
				if err != nil {
					fmt.Println(err)
					continue
				}

			}

		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("Command not recognized")

		}
	}
}
