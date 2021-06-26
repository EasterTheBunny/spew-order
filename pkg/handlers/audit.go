package handlers

import (
	"context"
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

		response := AuditResponse{
			UserAccounts:   make(map[string]string),
			LedgerAccounts: make(map[string]map[string]string), // map[symbol][account]balance
			Balance:        make(map[string]string),
			Errors:         []string{},
		}

		ctx := r.Context()
		symbols := []types.Symbol{types.SymbolBitcoin, types.SymbolEthereum}
		accountBalances, err := h.getAccountBalances(ctx, symbols)
		if err != nil {
			render.Render(w, r, HTTPInternalServerError(err))
			return
		}

		aBal := make(map[types.Symbol]decimal.Decimal)
		lBal := make(map[types.Symbol]decimal.Decimal)

		transferAssets, err := h.ledger.GetAssetBalance(ctx, persist.Transfers)
		if err != nil {
			render.Render(w, r, HTTPInternalServerError(err))
			return
		}
		payableLiabilities, err := h.ledger.GetLiabilityBalance(ctx, persist.TransfersPayable)
		if err != nil {
			render.Render(w, r, HTTPInternalServerError(err))
			return
		}

		for k, x := range accountBalances {
			response.UserAccounts[k.String()] = x.StringFixedBank(k.RoundingPlace())

			y, ok := transferAssets[k]
			if !ok {
				y = decimal.NewFromInt(0)
			}

			response.LedgerAccounts[k.String()] = map[string]string{
				persist.Transfers.String(): y.StringFixedBank(k.RoundingPlace()),
			}

			if !x.Equal(y) {
				msg := fmt.Sprintf("incorrect %s balance check for account balance and transfer: %s", k, y.StringFixedBank(k.RoundingPlace()))
				response.Errors = append(response.Errors, msg)
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

			response.LedgerAccounts[k.String()][persist.TransfersPayable.String()] = z.StringFixedBank(k.RoundingPlace())

			if !x.Equal(z) {
				msg := fmt.Sprintf("incorrect balance check for account balance and payable: %s", k)
				response.Errors = append(response.Errors, msg)
			}

			l, ok := lBal[k]
			if !ok {
				lBal[k] = z
			} else {
				lBal[k] = l.Add(z)
			}
		}

		cashAssets, err := h.ledger.GetAssetBalance(ctx, persist.Cash)
		if err != nil {
			render.Render(w, r, HTTPInternalServerError(err))
			return
		}
		salesLiabilities, err := h.ledger.GetLiabilityBalance(ctx, persist.Sales)
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

			response.LedgerAccounts[k.String()][persist.Cash.String()] = x.StringFixedBank(k.RoundingPlace())

			z, ok := salesLiabilities[k]
			if !ok || !x.Equal(z) {
				response.Errors = append(response.Errors, fmt.Sprintf("sales/cash discrepancy: %s", k))
			}

			response.LedgerAccounts[k.String()][persist.Sales.String()] = z.StringFixedBank(k.RoundingPlace())

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

			response.Balance[k.String()] = l.Sub(a).StringFixedBank(k.RoundingPlace())

			if !l.Equal(a) {
				response.Errors = append(response.Errors, fmt.Sprintf("liability/asset discrepancy: %s", k))
			}
		}

		render.Render(w, r, HTTPNewOKResponse(&response))
	}
}

func (h *AuditHandler) getAccountBalances(ctx context.Context, symbols []types.Symbol) (map[types.Symbol]decimal.Decimal, error) {
	var err error
	var auths []*persist.Authorization

	accountBalances := make(map[types.Symbol]decimal.Decimal)
	for _, s := range symbols {
		accountBalances[s] = decimal.NewFromInt(0)
	}

	auths, err = h.auths.GetAuthorizations(ctx)
	if err != nil {
		return accountBalances, err
	}

	// calculate totals for all symbols for all accounts
	for _, auth := range auths {
		for _, acc := range auth.Accounts {
			a := &persist.Account{ID: acc}

			for _, s := range symbols {
				br := h.accounts.Balances(a, s)

				var bal decimal.Decimal
				bal, err = br.GetBalance(ctx)
				if err != nil {
					return accountBalances, err
				}

				accountBalances[s] = accountBalances[s].Add(bal)

				var bi []*persist.BalanceItem
				bi, err = br.FindPosts(ctx)
				if err != nil {
					return accountBalances, err
				}

				for _, b := range bi {
					accountBalances[s] = accountBalances[s].Add(b.Amount)
				}
			}
		}
	}

	return accountBalances, nil
}

type AuditResponse struct {
	UserAccounts   map[string]string            `json:"user_accounts"`
	LedgerAccounts map[string]map[string]string `json:"ledger_accounts"`
	Balance        map[string]string            `json:"balance"`
	Errors         []string                     `json:"errors"`
}

// Render implements the render.Renderer interface for use with chi-router
func (a *AuditResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
