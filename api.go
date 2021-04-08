// Package contexttip is an example of how to use Pub/Sub and context.Context in
// a Cloud Function.
package contexttip

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
)

const (
	envProjectID  = "GOOGLE_CLOUD_PROJECT"
	envOrderTopic = "ORDER_TOPIC"
	envAppName    = "APP_NAME"       // application name used as prefix for named resources
	envRuntimeEnv = "DEPLOYMENT_ENV" // deployment environment; CI, QA, PROD
	envLocation   = "LOCATION"       // resources location for this function instanc
)

var (
	// client is a global Pub/Sub client, initialized once per instance.
	client        *pubsub.Client
	storageClient *storage.Client
	orderTopic    = getEnvVar(envOrderTopic)

	GS *persist.GoogleStorage
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

	storageClient, err := storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("storage.NewClient: %v", err)
	}

	GS = persist.NewGoogleStorage(persist.NewGoogleStorageAPI(storageClient))

	conf := []interface{}{
		getEnvVar(envAppName),
		persist.StorageBucket,
		strings.ToLower(getEnvVar(envRuntimeEnv)),
		strings.ToLower(getEnvVar(envLocation))}

	// get the primary storage bucket
	persist.StorageBucket = fmt.Sprintf("%s-%s-%s-%s", conf...)

	// register concrete types for the gob encoder/decoder
	gob.Register(types.LimitOrderType{})
	gob.Register(types.MarketOrderType{})
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

	/*
		if or.Quantity <= 0 {
			http.Error(w, fmt.Sprintf("incorrect quantity: %f", or.Quantity), http.StatusBadRequest)
			return
		}
	*/

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(or)
	if err != nil {
		log.Printf("gob.Encode: %v", err)
		http.Error(w, "Error encoding request", http.StatusBadRequest)
		return
	}

	/*
		b, err := json.Marshal(or)
		if err != nil {
			log.Printf("json.Marshal: %v", err)
			http.Error(w, "Error encoding request", http.StatusBadRequest)
			return
		}
	*/

	m := &pubsub.Message{
		Data: buf.Bytes(),
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
