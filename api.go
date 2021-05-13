// Package contexttip is an example of how to use Pub/Sub and context.Context in
// a Cloud Function.
package contexttip

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/easterthebunny/spew-order/internal/queue"
	"github.com/easterthebunny/spew-order/pkg/book"
	"github.com/easterthebunny/spew-order/pkg/handlers"
	"github.com/easterthebunny/spew-order/pkg/types"
)

const (
	envIdentityURI = "IDENTITY_PROVIDER" // identity provider as a URI path
	envProjectID   = "GOOGLE_CLOUD_PROJECT"
	envOrderTopic  = "ORDER_TOPIC"
	envAppName     = "APP_NAME"       // application name used as prefix for named resources
	envRuntimeEnv  = "DEPLOYMENT_ENV" // deployment environment; CI, QA, PROD
	envLocation    = "LOCATION"       // resources location for this function instanc
)

var (
	// client is a global Pub/Sub client, initialized once per instance.
	orderTopic = getEnvVar(envOrderTopic)

	GS     book.OrderBook
	Router http.Handler
)

func init() {
	// GOOGLE_CLOUD_PROJECT is a user-set environment variable.
	var projectID = getEnvVar(envProjectID)

	conf := []interface{}{
		getEnvVar(envAppName),
		"book",
		strings.ToLower(getEnvVar(envRuntimeEnv)),
		strings.ToLower(getEnvVar(envLocation))}

	bucket := fmt.Sprintf("%s-%s-%s-%s", conf...)
	queue.OrderTopic = orderTopic

	// register concrete types for the gob encoder/decoder
	//gob.Register(types.LimitOrderType{})
	//gob.Register(types.MarketOrderType{})
	ps := handlers.NewGooglePubSub(projectID)

	kv, err := handlers.NewGoogleKVStore(&bucket)
	if err != nil {
		log.Fatal(err.Error())
	}
	GS = handlers.NewGoogleOrderBook(kv)

	jwt, err := handlers.NewJWTAuth(envIdentityURI)
	if err != nil {
		log.Fatal(err.Error())
	}

	rh, err := handlers.NewDefaultRouter(kv, ps, jwt)
	if err != nil {
		log.Fatal(err.Error())
	}

	Router = rh.Routes()
}

// RestAPI forwards all rest requests to the main API handler.
func RestAPI(w http.ResponseWriter, r *http.Request) {
	Router.ServeHTTP(w, r)
}

// PubSubMessage is the payload of a Pub/Sub event.
// See the documentation for more details:
// https://cloud.google.com/pubsub/docs/reference/rest/v1/PubsubMessage
type PubSubMessage struct {
	Data []byte `json:"data"`
}

// OrderPubSub consumes a Pub/Sub message.
func OrderPubSub(ctx context.Context, m PubSubMessage) error {

	req := &types.OrderRequest{}
	if err := req.UnmarshalJSON(m.Data); err != nil {
		return err
	}

	order := types.NewOrderFromRequest(*req)
	if err := GS.ExecuteOrInsertOrder(order); err != nil {
		return err
	}

	return nil
}

func getEnvVar(key string) string {
	keyEnv, ok := os.LookupEnv(key)
	if !ok {
		err := fmt.Errorf("%s environment variable not available", key)
		panic(err)
	}

	return strings.Trim(keyEnv, "\n")
}
