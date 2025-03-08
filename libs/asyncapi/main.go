package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kaje94/slek-link/asyncapi/asyncapi"

	"github.com/ThreeDotsLabs/watermill/message"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	err := doAMQPPublish(ctx)
	if err != nil {
		log.Fatalf("error publishing to amqp: %s", err)
	}

	router, err := asyncapi.GetRouter()
	if err != nil {
		log.Fatalf("error getting router: %s", err)
	}

	err = startAMQPSubscribers(ctx, router)
	if err != nil {
		log.Fatalf("error starting amqp subscribers: %s", err)
	}

	if err = router.Run(ctx); err != nil {
		log.Fatalf("error running watermill router: %s", err)
	}

}

func doAMQPPublish(ctx context.Context) error {
	amqpPub, err := asyncapi.GetAMQPPublisher(asyncapi.GetAMQPURI())
	if err != nil {
		return err
	}

	var urlvisitedpayload asyncapi.UrlVisitedPayload
	//construct your message here
	err = asyncapi.PublishUrlVisited(ctx, amqpPub, urlvisitedpayload)
	if err != nil {
		return err
	}

	return nil
}

func startAMQPSubscribers(ctx context.Context, router *message.Router) error {
	amqpSubscriber, err := asyncapi.GetAMQPSubscriber(asyncapi.GetAMQPURI())
	if err != nil {
		return err
	}

	asyncapi.ConfigureAMQPSubscriptionHandlers(router, amqpSubscriber)
	return nil
}
