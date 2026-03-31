package main

import (
	"fmt"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func publishGameLog(ch *amqp.Channel, username, message string) error {
	routeKey := routing.GameLogSlug + "." + username
	gl := routing.GameLog{
		CurrentTime: time.Now(),
		Username:    username,
		Message:     message,
	}
	return pubsub.PublishGob(ch, routing.ExchangePerilTopic, routeKey, gl)
}

func handlerWar(gs *gamelogic.GameState, ch *amqp.Channel) func(dw gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(dw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(dw)

		switch outcome {

		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue

		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard

		case gamelogic.WarOutcomeOpponentWon:
			msg := fmt.Sprintf("%v won war against %v", winner, loser)
			err := publishGameLog(ch, gs.GetUsername(), msg)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack

		case gamelogic.WarOutcomeYouWon:
			msg := fmt.Sprintf("%v won war against %v", winner, loser)
			err := publishGameLog(ch, gs.GetUsername(), msg)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack

		case gamelogic.WarOutcomeDraw:
			msg := fmt.Sprintf("A war between %v and %v resulted in a draw", winner, loser)
			err := publishGameLog(ch, gs.GetUsername(), msg)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack

		default:
			fmt.Printf("erro")
			return pubsub.NackDiscard
		}
	}
}
