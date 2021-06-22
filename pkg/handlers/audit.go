package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/easterthebunny/render"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/shopspring/decimal"
)

type AuditHandler struct {
	accounts persist.AccountRepository
	auths    persist.AuthorizationRepository
	ledger   persist.LedgerRepository
}

func NewAuditHandler(a persist.AccountRepository, t persist.AuthorizationRepository, l persist.LedgerRepository) *AuditHandler {
	return &AuditHandler{
		accounts: a,
		auths:    t,
		ledger:   l}
}

func (h *AuditHandler) AuditBalances() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var auths []*persist.Authorization

		symbols := []types.Symbol{types.SymbolBitcoin, types.SymbolEthereum}
		accountBalances := make(map[types.Symbol]decimal.Decimal)
		for _, s := range symbols {
			accountBalances[s] = decimal.NewFromInt(0)
		}

		auths, err = h.auths.GetAuthorizations()
		if err != nil {
			render.Render(w, r, HTTPInternalServerError(err))
			return
		}

		for _, auth := range auths {
			for _, acc := range auth.Accounts {
				a := &persist.Account{ID: acc}

				for _, s := range symbols {
					br := h.accounts.Balances(a, s)

					var bal decimal.Decimal
					bal, err = br.GetBalance()
					if err != nil {
						render.Render(w, r, HTTPInternalServerError(err))
						return
					}

					accountBalances[s] = accountBalances[s].Add(bal)

					var bi []*persist.BalanceItem
					bi, err = br.FindPosts()
					if err != nil {
						render.Render(w, r, HTTPInternalServerError(err))
						return
					}

					for _, b := range bi {
						accountBalances[s] = accountBalances[s].Add(b.Amount)
					}
				}
			}
		}

		aBal := make(map[types.Symbol]decimal.Decimal)
		lBal := make(map[types.Symbol]decimal.Decimal)

		transferAssets, err := h.ledger.GetAssetBalance(persist.Transfers)
		if err != nil {
			render.Render(w, r, HTTPInternalServerError(err))
			return
		}
		payableLiabilities, err := h.ledger.GetLiabilityBalance(persist.TransfersPayable)
		if err != nil {
			render.Render(w, r, HTTPInternalServerError(err))
			return
		}

		for k, x := range accountBalances {
			y, ok := transferAssets[k]
			if !ok {
				y = decimal.NewFromInt(0)
			}

			if !x.Equal(y) {
				err = fmt.Errorf("incorrect %s balance check for account balance and transfer: %s", k, y.StringFixedBank(k.RoundingPlace()))
				render.Render(w, r, HTTPInternalServerError(err))
				return
			}

			a, ok := aBal[k]
			if !ok {
				aBal[k] = y
			} else {
				aBal[k] = a.Add(y)
			}

			z, ok := payableLiabilities[k]
			if !ok {
				z = decimal.NewFromInt(0)
			}

			if !x.Equal(z) {
				err = fmt.Errorf("incorrect balance check for account balance and payable: %s", k)
				render.Render(w, r, HTTPInternalServerError(err))
				return
			}

			l, ok := lBal[k]
			if !ok {
				lBal[k] = z
			} else {
				lBal[k] = l.Add(z)
			}
		}

		cashAssets, err := h.ledger.GetAssetBalance(persist.Cash)
		if err != nil {
			render.Render(w, r, HTTPInternalServerError(err))
			return
		}
		salesLiabilities, err := h.ledger.GetLiabilityBalance(persist.Sales)
		if err != nil {
			render.Render(w, r, HTTPInternalServerError(err))
			return
		}

		for k, x := range cashAssets {
			a, ok := aBal[k]
			if !ok {
				aBal[k] = x
			} else {
				aBal[k] = a.Add(x)
			}

			z, ok := salesLiabilities[k]
			if !ok || !x.Equal(z) {
				render.Render(w, r, HTTPInternalServerError(errors.New("incorrect balance")))
				return
			}

			l, ok := lBal[k]
			if !ok {
				lBal[k] = z
			} else {
				lBal[k] = l.Add(z)
			}
		}

		for k, a := range aBal {
			l, ok := lBal[k]
			if !ok {
				render.Render(w, r, HTTPInternalServerError(errors.New("missing balance")))
				return
			}

			if !l.Equal(a) {
				render.Render(w, r, HTTPInternalServerError(errors.New("incorrect balance")))
				return
			}
		}

		render.Render(w, r, HTTPNoContentResponse())
	}
}
