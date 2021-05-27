package handlers

import (
	"log"
	"net/http"

	"github.com/easterthebunny/render"
	"github.com/easterthebunny/spew-order/internal/funding"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/domain"
)

type FundingHandler struct {
	Balance *domain.BalanceManager
	Source  funding.Source
}

func NewFundingHandler(a persist.AccountRepository, l persist.LedgerRepository, s funding.Source) *FundingHandler {
	return &FundingHandler{
		Balance: domain.NewBalanceManager(a, l),
		Source:  s}
}

func (h *FundingHandler) PostFunding() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		tr, cerr := funding.TransactionFromContext(r.Context())
		if cerr != nil {
			log.Println(cerr.Err)
			render.Render(w, r, HTTPStatusError(cerr.Status, cerr.Err))
			return
		}

		err := h.Balance.FundAccountByAddress(tr.Address, tr.Symbol, tr.Amount)
		if err != nil {
			log.Println(err)
			render.Render(w, r, HTTPInternalServerError(err))
			return
		}

		w.WriteHeader(h.Source.OKResponse())
	}
}
