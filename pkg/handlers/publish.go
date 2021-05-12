package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/easterthebunny/spew-order/internal/contexts"
	"github.com/easterthebunny/spew-order/pkg/api"
	"github.com/easterthebunny/spew-order/pkg/queue"
)

var (
	ErrNoAccountIDFound = errors.New("no account identifier found in request")
)

type RESTHandler struct {
	queue *queue.OrderQueue
}

func NewRESTHandler(q *queue.OrderQueue) *RESTHandler {
	return &RESTHandler{queue: q}
}

// PostOrder publishes a message to Pub/Sub. PublishMessage only works
// with topics that already exist.
func (h *RESTHandler) PostOrder() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		aid, err := getAccountIDFromRequest(r, accountFromCookie, accountFromHeader, accountFromQuery)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx := contexts.AttachAccountID(r.Context(), aid)

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		or, err := api.OrderRequestFromBytes(b)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// TODO: validate order request

		id, err := h.queue.PublishOrderRequest(ctx, or)
		if err != nil {
			log.Printf("topic(%s).Publish.Get: %v", queue.OrderTopic, err)
			http.Error(w, "Error publishing message", http.StatusInternalServerError)
			return
		}

		res := api.BookOrder(id)
		out, err := json.Marshal(res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, out)
	}
}

func getAccountIDFromRequest(r *http.Request, findIDFns ...func(r *http.Request) string) (string, error) {
	var idString string

	for _, fn := range findIDFns {
		idString = fn(r)
		if idString != "" {
			break
		}
	}

	if idString == "" {
		return "", ErrNoAccountIDFound
	}

	return idString, nil
}

func accountFromCookie(r *http.Request) string {
	cookie, err := r.Cookie("account")
	if err != nil {
		return ""
	}
	return cookie.Value
}

func accountFromHeader(r *http.Request) string {
	return r.Header.Get("Account")
}

func accountFromQuery(r *http.Request) string {
	return r.URL.Query().Get("account")
}
