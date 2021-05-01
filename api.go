// Package contexttip is an example of how to use Pub/Sub and context.Context in
// a Cloud Function.
package contexttip

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/easterthebunny/spew-order/pkg/book"
	"github.com/easterthebunny/spew-order/pkg/queue"
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
	orderTopic = getEnvVar(envOrderTopic)

	GS book.OrderBook
	GQ queue.OrderQueue
)

func init() {
	// GOOGLE_CLOUD_PROJECT is a user-set environment variable.
	var projectID = getEnvVar(envProjectID)

	conf := []interface{}{
		getEnvVar(envAppName),
		"book",
		strings.ToLower(getEnvVar(envRuntimeEnv)),
		strings.ToLower(getEnvVar(envLocation))}

	GS = book.NewGoogleOrderBook(fmt.Sprintf("%s-%s-%s-%s", conf...))
	GQ = queue.NewGoogleOrderQueue(projectID)

	// register concrete types for the gob encoder/decoder
	//gob.Register(types.LimitOrderType{})
	//gob.Register(types.MarketOrderType{})
}

// PublishOrder publishes a message to Pub/Sub. PublishMessage only works
// with topics that already exist.
func PublishOrder(w http.ResponseWriter, r *http.Request) {
	var or types.OrderRequest

	if err := json.NewDecoder(r.Body).Decode(&or); err != nil {
		log.Printf("json.NewDecoder: %v", err)
		http.Error(w, "Error parsing request", http.StatusBadRequest)
		return
	}

	id, err := GQ.PublishOrderRequest(r.Context(), or)
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
