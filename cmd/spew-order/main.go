package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/internal/queue"
	"github.com/easterthebunny/spew-order/pkg/domain"
	"github.com/easterthebunny/spew-order/pkg/handlers"
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/go-chi/chi"
)

func main() {

	log.Println("starting service")
	kvstore := persist.NewMockKVStore()
	f := handlers.NewFundingSource("MOCK", nil, nil, nil, nil)
	book := handlers.NewGoogleOrderBook(kvstore, f)
	ps := queue.NewMockPubSub()
	jwt := &mockJWTAuth{}
	subscription := make(chan domain.PubSubMessage)
	ps.Subscribe(queue.OrderTopic, subscription)

	rh, err := handlers.NewDefaultRouter(kvstore, ps, jwt, f)
	if err != nil {
		log.Fatal(err.Error())
	}

	wh := handlers.NewWebhookRouter(kvstore, f)

	wg := new(sync.WaitGroup)

	// start the api service
	go func() {
		defer wg.Done()
		wg.Add(1)

		host := "0.0.0.0:8080"
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
			var order types.Order
			if err := json.Unmarshal(m.Data, &order); err != nil {
				log.Printf("error: %s", err)
				continue
			}

			if err := book.ExecuteOrInsertOrder(order); err != nil {
				log.Printf("error: %s", err)
				continue
			}
		}
	}()

	<-time.Tick(time.Second)

	wg.Wait()
}

type mockJWTAuth struct {
	subject string
}

func (j *mockJWTAuth) Verifier() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			j.subject = tokenFromHeader(r)
			next.ServeHTTP(w, r)
		})
	}
}

func (j *mockJWTAuth) Subject() string {
	return j.subject
}

func (j *mockJWTAuth) UpdateAuthz(a *persist.Authorization) {
	a.ID = j.Subject()
	a.Email = "test@email.com"
	a.Avatar = "picture/path"
	a.Name = "Test Person"
	a.Username = "username"
}

func tokenFromHeader(r *http.Request) string {
	// Get token from authorization header.
	bearer := r.Header.Get("Authorization")
	if len(bearer) > 7 && strings.ToUpper(bearer[0:6]) == "BEARER" {
		return bearer[7:]
	}
	return ""
}
