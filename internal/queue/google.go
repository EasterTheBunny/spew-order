package queue

import (
	"context"
	"log"

	"cloud.google.com/go/pubsub"
)

type PubSub interface {
	Publish(context.Context, string, []byte) (string, error)
}

func NewGooglePubSub(projectID string) PubSub {

	// client is initialized with context.Background() because it should
	// persist between function invocations.
	client, err := pubsub.NewClient(context.Background(), projectID)
	if err != nil {
		log.Fatalf("pubsub.NewClient: %v", err)
	}

	return &GooglePubSub{
		client: client}
}

type GooglePubSub struct {
	client *pubsub.Client
}

func (g *GooglePubSub) Publish(ctx context.Context, topic string, data []byte) (id string, err error) {
	m := &pubsub.Message{
		Data: data,
	}

	return g.client.Topic(topic).Publish(ctx, m).Get(ctx)
}
