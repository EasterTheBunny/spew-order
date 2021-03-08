// Package contexttip is an example of how to use Pub/Sub and context.Context in
// a Cloud Function.
package contexttip

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/pubsub"
	"github.com/easterthebunny/spew-order/pkg/types"
)

const (
	envProjectID  = "GOOGLE_CLOUD_PROJECT"
	envOrderTopic = "ORDER_TOPIC"
)

var (
	// client is a global Pub/Sub client, initialized once per instance.
	client     *pubsub.Client
	orderTopic = getEnvVar(envOrderTopic)
)

func init() {
	// GOOGLE_CLOUD_PROJECT is a user-set environment variable.
	var projectID = getEnvVar(envProjectID)

	// err is pre-declared to avoid shadowing client.
	var err error

	// client is initialized with context.Background() because it should
	// persist between function invocations.
	client, err = pubsub.NewClient(context.Background(), projectID)
	if err != nil {
		log.Fatalf("pubsub.NewClient: %v", err)
	}
}

// PublishOrder publishes a message to Pub/Sub. PublishMessage only works
// with topics that already exist.
func PublishOrder(w http.ResponseWriter, r *http.Request) {
	// Parse the request body to get the topic name and message.
	var or types.OrderRequest

	if err := json.NewDecoder(r.Body).Decode(&or); err != nil {
		log.Printf("json.NewDecoder: %v", err)
		http.Error(w, "Error parsing request", http.StatusBadRequest)
		return
	}

	if or.Quantity <= 0 {
		http.Error(w, fmt.Sprintf("incorrect quantity: %f", or.Quantity), http.StatusBadRequest)
		return
	}

	b, err := json.Marshal(or)
	if err != nil {
		log.Printf("json.Marshal: %v", err)
		http.Error(w, "Error encoding request", http.StatusBadRequest)
		return
	}

	m := &pubsub.Message{
		Data: b,
	}

	// Publish and Get use r.Context() because they are only needed for this
	// function invocation. If this were a background function, they would use
	// the ctx passed as an argument.
	id, err := client.Topic(orderTopic).Publish(r.Context(), m).Get(r.Context())
	if err != nil {
		log.Printf("topic(%s).Publish.Get: %v", orderTopic, err)
		http.Error(w, "Error publishing message", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Message published: %v", id)
}

func getEnvVar(key string) string {
	keyEnv, ok := os.LookupEnv(key)
	if !ok {
		err := fmt.Errorf("%s environment variable not available", key)
		panic(err)
	}

	return strings.Trim(keyEnv, "\n")
}
