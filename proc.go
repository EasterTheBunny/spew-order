package contexttip

import (
	"context"
	"log"
)

// PubSubMessage is the payload of a Pub/Sub event.
// See the documentation for more details:
// https://cloud.google.com/pubsub/docs/reference/rest/v1/PubsubMessage
type PubSubMessage struct {
	Data []byte `json:"data"`
}

// OrderPubSub consumes a Pub/Sub message.
func OrderPubSub(ctx context.Context, m PubSubMessage) error {
	data := string(m.Data) // Automatically decoded from base64.
	log.Printf("%s", data)
	return nil
}
