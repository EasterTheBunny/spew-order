package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/easterthebunny/render"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/api"
	"github.com/easterthebunny/spew-order/pkg/domain"
	"github.com/easterthebunny/spew-order/pkg/types"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

type FundingHandler struct {
	Balance *domain.BalanceManager
}

func NewFundingHandler(a persist.AccountRepository, l persist.LedgerRepository) *FundingHandler {
	return &FundingHandler{
		Balance: domain.NewBalanceManager(a, l)}
}

func (h *FundingHandler) PostFunding() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			render.Render(w, r, HTTPBadRequest(err))
			return
		}

		var callback api.FundingCallback
		err = json.Unmarshal(b, &callback)
		if err != nil {
			render.Render(w, r, HTTPBadRequest(err))
			return
		}

		var sym types.Symbol
		err = json.Unmarshal([]byte(fmt.Sprintf(`"%s"`, string(callback.Symbol))), &sym)
		if err != nil {
			render.Render(w, r, HTTPBadRequest(err))
			return
		}

		id, err := uuid.FromString(callback.Account)
		if err != nil {
			render.Render(w, r, HTTPBadRequest(err))
			return
		}

		amt, err := decimal.NewFromString(string(callback.Quantity))
		if err != nil {
			render.Render(w, r, HTTPBadRequest(err))
			return
		}

		err = h.Balance.FundAccountByID(id, sym, amt)
		if err != nil {
			render.Render(w, r, HTTPInternalServerError(err))
			return
		}

		render.Render(w, r, HTTPNewOKResponse(&api.CallbackResponse{Message: "ok"}))
	}
}
