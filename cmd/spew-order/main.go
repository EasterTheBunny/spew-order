package main

import (
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/internal/persist/kv"
	"github.com/easterthebunny/spew-order/internal/queue"
	"github.com/easterthebunny/spew-order/pkg/domain"
	"github.com/easterthebunny/spew-order/pkg/handlers"
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/go-chi/chi"
)

func main() {

	log.Println("starting service")
	kvstore := persist.NewMockKVStore()
	book := domain.NewOrderBook(kv.NewBookRepository(kvstore))
	ps := queue.NewMockPubSub()
	jwt := &mockJWTAuth{}
	subscription := make(chan domain.OrderMessage)
	ps.Subscribe(queue.OrderTopic, subscription)

	rh, err := handlers.NewDefaultRouter(kvstore, ps, jwt)
	if err != nil {
		log.Fatal(err.Error())
	}

	wh := handlers.NewWebhookRouter(kvstore)

	wg := new(sync.WaitGroup)

	// start the api service
	go func() {
		defer wg.Done()
		wg.Add(1)

		host := "0.0.0.0:9999"
		log.Printf("starting api listener on %s", host)

		uni := func(api http.Handler, webhook http.Handler) http.Handler {
			r := chi.NewRouter()
			r.Mount("/api", api)
			r.Mount("/webhook", webhook)
			return r
		}

		l, _ := net.Listen("tcp", host)
		srv := &http.Server{Handler: uni(rh.Routes(), wh.Routes())}

		err := srv.Serve(l)
		if err != nil {
			return
		}
	}()

	// start the pubsub subscription handler
	go func() {
		defer wg.Done()
		wg.Add(1)
		log.Println("starting pubsub listener")
		for {
			m := <-subscription
			req := &types.OrderRequest{}
			if err := req.UnmarshalJSON(m.Data); err != nil {
				log.Printf("error: %s", err)
				continue
			}

			order := types.NewOrderFromRequest(*req)
			if err := book.ExecuteOrInsertOrder(order); err != nil {
				log.Printf("error: %s", err)
				continue
			}
		}
	}()

	<-time.Tick(time.Second)

	wg.Wait()
}

type mockJWTAuth struct{}

func (j *mockJWTAuth) Verifier() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
}

func (j *mockJWTAuth) Subject() string {
	return "test_subject"
}

func (j *mockJWTAuth) UpdateAuthz(a *persist.Authorization) {
	a.ID = j.Subject()
	a.Email = "test@email.com"
	a.Avatar = "picture/path"
	a.Name = "Test Person"
	a.Username = "username"
}
