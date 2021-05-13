package handlers

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/easterthebunny/render"
	"github.com/easterthebunny/spew-order/internal/contexts"
	"github.com/easterthebunny/spew-order/internal/queue"
	"github.com/easterthebunny/spew-order/pkg/api"
)

var (
	ErrNoAccountIDFound = errors.New("no account identifier found in request")
)

type OrderHandler struct {
	queue *queue.OrderQueue
}

func NewOrderHandler(q *queue.OrderQueue) *OrderHandler {
	return &OrderHandler{queue: q}
}

// PostOrder publishes a message to Pub/Sub. PublishMessage only works
// with topics that already exist.
func (h *OrderHandler) PostOrder() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		acct := contexts.GetAccount(ctx)
		if acct == nil {
			render.Render(w, r, HTTPInternalServerError(errors.New("incorrect route structure")))
			return
		}
		ctx = contexts.AttachAccountID(ctx, acct.ID.String())

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			render.Render(w, r, HTTPBadRequest(err))
			return
		}

		or, err := api.OrderRequestFromBytes(b)
		if err != nil {
			render.Render(w, r, HTTPBadRequest(err))
			return
		}

		// TODO: validate order request

		id, err := h.queue.PublishOrderRequest(ctx, or)
		if err != nil {
			log.Printf("PostOrder.PublistOrderRequest: %v", err)
			render.Render(w, r, HTTPInternalServerError(err))
			return
		}

		render.Render(w, r, HTTPNewOKResponse(&api.BookOrder{Guid: id}))
	}
}
