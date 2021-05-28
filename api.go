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

	"github.com/easterthebunny/spew-order/internal/queue"
	"github.com/easterthebunny/spew-order/pkg/domain"
	"github.com/easterthebunny/spew-order/pkg/handlers"
	"github.com/easterthebunny/spew-order/pkg/types"
)

const (
	envIdentityURI       = "IDENTITY_PROVIDER" // identity provider as a URI path
	envProjectID         = "GOOGLE_CLOUD_PROJECT"
	envOrderTopic        = "ORDER_TOPIC"
	envAppName           = "APP_NAME"                // application name used as prefix for named resources
	envRuntimeEnv        = "DEPLOYMENT_ENV"          // deployment environment; CI, QA, PROD
	envLocation          = "LOCATION"                // resources location for this function instanc
	envCoinbasePubKey    = "COINBASE_RSA_PUBLIC_KEY" // rsa public key for verifying signature of coinbase notifications
	envCoinbaseAPIKey    = "COINBASE_API_KEY"        // api key for coinbase
	envCoinbaseAPISecret = "COINBASE_API_SECRET"     // api secret for coinbase
)

var (
	// client is a global Pub/Sub client, initialized once per instance.
	orderTopic = getEnvVar(envOrderTopic)

	GS       *domain.OrderBook
	Router   http.Handler
	Webhooks http.Handler
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

	pubKey := strings.NewReader(getEnvVar(envCoinbasePubKey))
	ky := getEnvVar(envCoinbaseAPIKey)
	sct := getEnvVar(envCoinbaseAPISecret)
	f := handlers.NewFundingSource("COINBASE", &ky, &sct, log.Writer(), pubKey)
	ps := handlers.NewGooglePubSub(projectID)

	kv, err := handlers.NewGoogleKVStore(&bucket)
	if err != nil {
		log.Fatal(err.Error())
	}
	GS = handlers.NewGoogleOrderBook(kv, f)

	jwt, err := handlers.NewJWTAuth(envIdentityURI)
	if err != nil {
		log.Fatal(err.Error())
	}

	rh, err := handlers.NewDefaultRouter(kv, ps, jwt, f)
	if err != nil {
		log.Fatal(err.Error())
	}

	Router = rh.Routes()
	Webhooks = handlers.NewWebhookRouter(kv, f).Routes()
}

// RestAPI forwards all rest requests to the main API handler.
func RestAPI(w http.ResponseWriter, r *http.Request) {
	Router.ServeHTTP(w, r)
}

// FundingWebhooks includes webhook endpoints to handle funding
// for notices for all assets
func FundingWebhooks(w http.ResponseWriter, r *http.Request) {
	Webhooks.ServeHTTP(w, r)
}

// OrderPubSub consumes a Pub/Sub message.
func OrderPubSub(ctx context.Context, m domain.OrderMessage) error {

	var order types.Order
	if err := json.Unmarshal(m.Data, &order); err != nil {
		return err
	}

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
