package handlers

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	iaccount "github.com/easterthebunny/spew-order/internal/account"
	"github.com/easterthebunny/spew-order/internal/persist"
	iqueue "github.com/easterthebunny/spew-order/internal/queue"
	"github.com/easterthebunny/spew-order/pkg/account"
	"github.com/easterthebunny/spew-order/pkg/queue"
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestPostOrder(t *testing.T) {
	data := `{"base":"BTC","target":"ETH","action":"BUY","type":%s}`
	limitType := `{"base":"BTC","name":"LIMIT","price":"0.0234","quantity":"0.0000042"}`

	// set up a buffer to log to
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)

	// set up the mocked pub sub and establish a subscription to the topic
	subscription := make(chan []byte)
	mps := iqueue.NewMockPubSub()
	mps.Subscribe(queue.OrderTopic, subscription)

	acct := types.NewAccount()
	repo := iaccount.NewKVAccountRepository(persist.NewMockKVStore())
	err := repo.Save(&acct)
	if err != nil {
		t.FailNow()
	}
	svc := account.NewBalanceService(repo)
	svc.PostToBalance(&acct, types.SymbolBitcoin, decimal.NewFromFloat(5.5))

	oq := queue.NewOrderQueue(mps, svc)

	// create handler to test
	handler := NewRESTHandler(oq)

	t.Run("SuccessPath", func(t *testing.T) {
		// create a response recorder for later inspection of the response
		w := httptest.NewRecorder()

		r := req(t, NewPost(fmt.Sprintf(data, limitType)))
		r.Header.Set("Account", acct.ID.String())

		handler.PostOrder()(w, r)

		assert.Equal(t, 200, w.Code, "response code is a 200 success")

		if len(logBuf.Bytes()) != 0 {
			t.Errorf("unexpected log output: %s", &logBuf)
		}

		// the new order should be published to the order queue within the
		// handler. wait for the posting and fail if not found
		select {
		case <-time.After(100 * time.Millisecond):
			t.Errorf("no data found on the queue subscription")
		case <-subscription:
			// happy case
			return
		}
	})

	// the handler requires an account id to be in the header, query, or cookie
	t.Run("MissingAccount", func(t *testing.T) {
		// create a response recorder for later inspection of the response
		w := httptest.NewRecorder()

		r := req(t, NewPost(fmt.Sprintf(data, limitType)))

		handler.PostOrder()(w, r)

		assert.Equal(t, 400, w.Code, "response code is a 400 Bad Request")

		// the new order should NOT be posted
		select {
		case <-time.After(100 * time.Millisecond):
			return
		case <-subscription:
			t.Errorf("data found on the queue subscription")
		}
	})
}

func NewPost(cont string) string {
	post :=
		`POST / HTTP/1.1
Content-Type: application/json
User-Agent: mockagent
Content-Length: %d

%s`
	return fmt.Sprintf(post, len(cont), cont)
}

func req(t testing.TB, v string) *http.Request {
	req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(v)))
	if err != nil {
		t.Fatal(err)
	}
	return req
}
