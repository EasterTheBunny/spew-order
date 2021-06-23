package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/easterthebunny/spew-order/internal/contexts"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/internal/persist/firebase"
	"github.com/easterthebunny/spew-order/pkg/api"
	"github.com/easterthebunny/spew-order/pkg/domain"
	"github.com/easterthebunny/spew-order/pkg/handlers"
	"github.com/go-chi/chi"
)

func main() {

	arepo := firebase.NewAccountRepository(nil)
	lrepo := firebase.NewLedgerRepository(nil)
	h := handlers.NewAccountHandler(arepo)
	f := handlers.NewFundingSource("MOCK", nil, nil, nil, nil)
	bm := domain.NewBalanceManager(arepo, lrepo, f)

	router := chi.NewRouter()
	router.Use(h.AccountCtx(bm, paramFunc))
	router.Get(fmt.Sprintf("/{%s}", api.AccountPathParamName), h.GetAccount())

	start := time.Now()
	wg := new(sync.WaitGroup)

	for x := 0; x < 10; x++ {

		acct := domain.NewAccount()
		authz := persist.Authorization{
			Accounts: []string{acct.ID.String()},
		}
		r := NewGet(fmt.Sprintf("/%s", acct.ID))
		ctx := r.Context()

		client, err := firestore.NewClient(ctx, "centering-rex-274623")
		if err != nil {
			panic(err)
		}

		ctx = context.WithValue(ctx, ctxKey{api.AccountPathParamName}, acct.ID.String())
		ctx = context.WithValue(ctx, firebase.ClientContextKey, client)
		r = r.WithContext(contexts.AttachAuthorization(ctx, authz))

		// create a response recorder for later inspection of the response
		w := httptest.NewRecorder()

		wg.Add(1)
		go func() {
			defer wg.Done()

			router.ServeHTTP(w, r)

			body, err := ioutil.ReadAll(w.Body)
			if err != nil {
				panic(err)
			}
			fmt.Printf("response code: %d\n", w.Code)
			fmt.Printf("body: %s\n", string(body))
		}()
	}

	wg.Wait()

	end := time.Now()

	fmt.Printf("time taken: %d\n", end.Unix()-start.Unix())
}

func NewGet(path string) *http.Request {
	fmt.Println(path)
	request, err := http.NewRequest(http.MethodGet, path, nil)
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("User-Agent", "mockagent")
	if err != nil {
		panic(err)
	}
	return request
}

func paramFunc(r *http.Request, name string) string {
	ctx := r.Context()
	val := ctx.Value(ctxKey{name}).(string)
	return val
}

type ctxKey struct {
	name string
}

func (k ctxKey) String() string {
	return "context value " + k.name
}
