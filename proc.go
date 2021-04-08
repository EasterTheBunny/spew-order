package contexttip

import (
	"bytes"
	"context"
	"encoding/gob"
	"log"

	"github.com/easterthebunny/spew-order/pkg/types"
)

// PubSubMessage is the payload of a Pub/Sub event.
// See the documentation for more details:
// https://cloud.google.com/pubsub/docs/reference/rest/v1/PubsubMessage
type PubSubMessage struct {
	Data []byte `json:"data"`
}

// OrderPubSub consumes a Pub/Sub message.
func OrderPubSub(ctx context.Context, m PubSubMessage) error {
	buf := bytes.NewBuffer(m.Data)

	dec := gob.NewDecoder(buf)
	var req types.OrderRequest
	err := dec.Decode(&req)
	if err != nil {
		return err
	}

	/*
		if err := order.UnmarshalJSON(m.Data); err != nil {
			return err
		}
	*/
	order := types.NewOrderFromRequest(req)

	log.Printf("%s", order.ID)

	/*
		if err := GS.ExecuteOrInsertOrder(order); err != nil {
			return err
		}
	*/

	return nil
}
