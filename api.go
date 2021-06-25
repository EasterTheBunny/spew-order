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

	"cloud.google.com/go/firestore"
	"github.com/easterthebunny/spew-order/internal/queue"
	"github.com/easterthebunny/spew-order/pkg/domain"
	"github.com/easterthebunny/spew-order/pkg/handlers"
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
	Audit    http.Handler
)

func init() {
	// GOOGLE_CLOUD_PROJECT is a user-set environment variable.
	var projectID = getEnvVar(envProjectID)

	queue.OrderTopic = orderTopic

	pubKey := strings.NewReader(getEnvVar(envCoinbasePubKey))
	ky := getEnvVar(envCoinbaseAPIKey)
	sct := getEnvVar(envCoinbaseAPISecret)
	srcType := "COINBASE"
	if strings.ToUpper(getEnvVar(envRuntimeEnv)) != "PROD" {
		srcType = "MOCK"
	}
	f := handlers.NewFundingSource(srcType, &ky, &sct, log.Writer(), pubKey)
	ps := handlers.NewGooglePubSub(projectID)

	client, err := firestore.NewClient(context.Background(), projectID)
	if err != nil {
		panic(err)
	}

	GS = handlers.NewGoogleOrderBook(client, f)

	jwt, err := handlers.NewJWTAuth(getEnvVar(envIdentityURI))
	if err != nil {
		log.Fatal(err.Error())
	}

	rh, err := handlers.NewDefaultRouter(client, ps, jwt, f)
	if err != nil {
		log.Fatal(err.Error())
	}

	Router = rh.Routes()
	Webhooks = handlers.NewWebhookRouter(client, f).Routes()
	Audit = handlers.NewAuditRouter(client).Routes()
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

// AuditAPI ...
func AuditAPI(w http.ResponseWriter, r *http.Request) {
	Audit.ServeHTTP(w, r)
}

// OrderPubSub consumes a Pub/Sub message.
func OrderPubSub(ctx context.Context, m domain.PubSubMessage) error {

	var msg domain.OrderMessage
	if err := json.Unmarshal(m.Data, &msg); err != nil {
		return err
	}

	if msg.Action == domain.CancelOrderMessageType {
		return GS.CancelOrder(ctx, msg.Order)
	} else if msg.Action == domain.OpenOrderMessageType {
		return GS.ExecuteOrInsertOrder(ctx, msg.Order)
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
