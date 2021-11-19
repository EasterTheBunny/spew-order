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
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/shopspring/decimal"
)

const (
	envIdentityURI       = "IDENTITY_PROVIDER" // identity provider as a URI path
	envProjectID         = "GOOGLE_CLOUD_PROJECT"
	envOrderTopic        = "ORDER_TOPIC"
	envRuntimeEnv        = "DEPLOYMENT_ENV"          // deployment environment; CI, QA, PROD
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

	// add funds to new accounts
	domain.FundNewAccounts = true
	domain.NewAccountFunds = decimal.NewFromInt(5000)

	pubKeySrc := strings.NewReader(`-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA9MsJBuXzFGIh/xkAA9Cy
QdZKRerV+apyOAWY7sEYV/AJg+AX/tW2SHeZj+3OilNYm5DlBi6ZzDboczmENrFn
mUXQsecsR5qjdDWb2qYqBkDkoZP02m9o9UmKObR8coKW4ZBw0hEf3fP9OEofG2s7
Z6PReWFyQffnnecwXJoN22qjjsUtNNKOOo7/l+IyGMVmdzJbMWQS4ybaU9r9Ax0J
4QUJSS/S4j4LP+3Z9i2DzIe4+PGa4Nf7fQWLwE45UUp5SmplxBfvEGwYNEsHvmRj
usIy2ZunSO2CjJ/xGGn9+/57W7/SNVzk/DlDWLaN27hUFLEINlWXeYLBPjw5GGWp
ieXGVcTaFSLBWX3JbOJ2o2L4MxinXjTtpiKjem9197QXSVZ/zF1DI8tRipsgZWT2
/UQMqsJoVRXHveY9q9VrCLe97FKAUiohLsskr0USrMCUYvLU9mMw15hwtzZlKY8T
dMH2Ugqv/CPBuYf1Bc7FAsKJwdC504e8kAUgomi4tKuUo25LPZJMTvMTs/9IsRJv
I7ibYmVR3xNsVEpupdFcTJYGzOQBo8orHKPFn1jj31DIIKociCwu6m8ICDgLuMHj
7bUHIlTzPPT7hRPyBQ1KdyvwxbguqpNhqp1hG2sghgMr0M6KMkUEz38JFElsVrpF
4z+EqsFcIZzjkSG16BjjjTkCAwEAAQ==
-----END PUBLIC KEY-----

date: 2014-07-09 13:37:00 UTC
version: 1`)

	pubKey := strings.NewReader(pubKeySrc)

	ky := getEnvVar(envCoinbaseAPIKey)
	sct := getEnvVar(envCoinbaseAPISecret)
	srcType := "COINBASE"
	if strings.ToUpper(getEnvVar(envRuntimeEnv)) != "PROD" {
		srcType = "MOCK"
	}
	f := handlers.NewFundingSource(srcType, &ky, &sct, log.Writer(), pubKey)
	air := handlers.NewFundingSource("CMTN", &ky, nil, nil, nil)
	ps := handlers.NewGooglePubSub(projectID)

	client, err := firestore.NewClient(context.Background(), projectID)
	if err != nil {
		panic(err)
	}

	GS = handlers.NewGoogleOrderBook(client, f, air)

	jwt, err := handlers.NewJWTAuth(getEnvVar(envIdentityURI))
	if err != nil {
		log.Fatal(err.Error())
	}

	rh, err := handlers.NewDefaultRouter(client, ps, jwt, f, air)
	if err != nil {
		log.Fatal(err.Error())
	}

	types.MakerFee = 0.0025
	types.TakerFee = 0.0050

	Router = rh.Routes()
	Webhooks = handlers.NewWebhookRouter(client, f, air).Routes()
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
