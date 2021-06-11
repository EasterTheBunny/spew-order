package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/easterthebunny/render"
	"github.com/easterthebunny/spew-order/internal/contexts"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/internal/queue"
	"github.com/easterthebunny/spew-order/pkg/api"
	"github.com/easterthebunny/spew-order/pkg/domain"
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/shopspring/decimal"
)

var (
	ErrNoAccountIDFound = errors.New("no account identifier found in request")
)

type OrderHandler struct {
	queue *queue.OrderQueue
}

type patchType string

const (
	patchTypeReplace = "replace"
)

type statusPatch struct {
	Operation patchType       `json:"op"`
	Path      string          `json:"path"`
	Value     api.OrderStatus `json:"value"`
}

func NewOrderHandler(q *queue.OrderQueue) *OrderHandler {
	return &OrderHandler{queue: q}
}

func (h *OrderHandler) CancelOrder() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			render.Render(w, r, HTTPBadRequest(fmt.Errorf("%s method not allowed", r.Method)))
			return
		}

		var patches []statusPatch
		err := json.NewDecoder(r.Body).Decode(&patches)
		if err != nil {
			render.Render(w, r, HTTPBadRequest(err))
			return
		}

		if len(patches) != 1 {
			render.Render(w, r, HTTPBadRequest(errors.New("only status value allowed to be patched")))
			return
		}

		patch := patches[0]
		if patch.Path != "/status" || patch.Operation != patchTypeReplace || api.OrderStatusValue(patch.Value) != persist.StatusCanceled {
			render.Render(w, r, HTTPBadRequest(errors.New("only status value allowed to be patched")))
			return
		}

		order := contexts.GetOrder(r.Context())

		if order.Status == persist.StatusCanceled {
			render.Render(w, r, HTTPBadRequest(errors.New("order already cancelled")))
			return
		}

		err = h.queue.CancelOrder(r.Context(), order.Base)
		if err != nil {
			render.Render(w, r, HTTPInternalServerError(err))
			return
		}

		o := api.BookOrder{
			Guid:   order.Base.ID.String(),
			Order:  api.BuildOrderRequest(order.Base.OrderRequest),
			Status: api.StringOrderStatus(persist.StatusCanceled),
		}
		render.Render(w, r, HTTPNewOKResponse(&o))
	}
}

// PostOrder publishes a message to Pub/Sub. PublishMessage only works
// with topics that already exist.
func (h *OrderHandler) PostOrder() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		acct := contexts.GetAccount(ctx)
		authz := contexts.GetAuthorization(ctx)
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

		or.Account = acct.ID
		or.Owner = authz.ID

		if !validPair(or.Base, or.Target) {
			render.Render(w, r, HTTPBadRequest(errors.New("invalid trade pair")))
			return
		}

		switch t := or.Type.(type) {
		case *types.MarketOrderType:
			if (or.Action == types.ActionTypeBuy && t.Base != or.Base) || (or.Action == types.ActionTypeSell && t.Base != or.Target) {
				render.Render(w, r, HTTPBadRequest(errors.New("quantity based market orders not supported")))
				return
			}

			if t.Quantity.LessThanOrEqual(decimal.NewFromInt(0)) {
				render.Render(w, r, HTTPBadRequest(errors.New("quantity must be greater than 0")))
				return
			}
		case *types.LimitOrderType:
			if t.Base != or.Base {
				render.Render(w, r, HTTPBadRequest(errors.New("incorrect base value for limit order")))
				return
			}

			if t.Price.LessThanOrEqual(decimal.NewFromInt(0)) {
				render.Render(w, r, HTTPBadRequest(errors.New("price must be greater than 0")))
				return
			}

			if t.Quantity.LessThanOrEqual(decimal.NewFromInt(0)) {
				render.Render(w, r, HTTPBadRequest(errors.New("quantity must be greater than 0")))
				return
			}
		default:
			render.Render(w, r, HTTPBadRequest(errors.New("incorrect order type")))
			return
		}

		order, err := h.queue.PublishOrderRequest(ctx, or)
		if err != nil {
			if errors.Is(domain.ErrInsufficientBalanceForHold, err) {
				render.Render(w, r, HTTPConflict(err))
				return
			}

			render.Render(w, r, HTTPInternalServerError(err))
			return
		}

		o := api.BookOrder{
			Guid:   order.ID.String(),
			Order:  api.BuildOrderRequest(order.OrderRequest),
			Status: api.StringOrderStatus(persist.StatusOpen),
		}
		render.Render(w, r, HTTPNewOKResponse(&o))
	}
}

func validPair(a, b types.Symbol) bool {

	pair := fmt.Sprintf("%s%s", a, b)
	for _, p := range types.ValidPairs {
		if p == pair {
			return true
		}
	}

	return false
}
