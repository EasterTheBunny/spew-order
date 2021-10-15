package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/easterthebunny/render"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/shopspring/decimal"
)

type AuditHandler struct {
	accounts persist.AccountRepository
	auths    persist.AuthorizationRepository
	ledger   persist.LedgerRepository
	book     persist.BookRepository
}

func NewAuditHandler(a persist.AccountRepository, t persist.AuthorizationRepository, l persist.LedgerRepository, b persist.BookRepository) *AuditHandler {
	return &AuditHandler{
		accounts: a,
		auths:    t,
		ledger:   l,
		book:     b}
}

func (h *AuditHandler) AuditBalances() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		response := AuditResponse{
			UserAccounts:   make(map[string]string),
			LedgerAccounts: make(map[string]map[string]string), // map[symbol][account]balance
			Balance:        make(map[string]string),
			Throughput:     make(map[string]string),
			Errors:         []string{},
		}

		ctx := r.Context()
		symbols := []types.Symbol{
			types.SymbolBitcoin,
			types.SymbolEthereum,
			types.SymbolCipherMtn}
		accountBalances, orderBalances, msgs, err := h.getAccountBalances(ctx, symbols)
		if err != nil {
			render.Render(w, r, HTTPInternalServerError(err))
			return
		}
		response.Errors = append(response.Errors, msgs...)

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

		for k, x := range orderBalances {
			response.Throughput[k.String()] = x.StringFixedBank(k.RoundingPlace())
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

func (h *AuditHandler) getAccountBalances(ctx context.Context, symbols []types.Symbol) (map[types.Symbol]decimal.Decimal, map[types.Symbol]decimal.Decimal, []string, error) {
	var err error
	var msgs []string
	var auths []*persist.Authorization

	accountBalances := make(map[types.Symbol]decimal.Decimal)
	orderBalance := make(map[types.Symbol]decimal.Decimal)
	for _, s := range symbols {
		accountBalances[s] = decimal.NewFromInt(0)
		orderBalance[s] = decimal.NewFromInt(0)
	}

	auths, err = h.auths.GetAuthorizations(ctx)
	if err != nil {
		return accountBalances, orderBalance, msgs, err
	}

	// calculate totals for all symbols for all accounts
	for _, auth := range auths {
		for _, acc := range auth.Accounts {
			a := &persist.Account{ID: acc}

			trbals := map[types.Symbol]decimal.Decimal{
				types.SymbolBitcoin:   decimal.NewFromInt(0),
				types.SymbolEthereum:  decimal.NewFromInt(0),
				types.SymbolCipherMtn: decimal.NewFromInt(0),
			}

			trepo := h.accounts.Transactions(a)
			transactions, err := trepo.GetTransactions(ctx)
			if err != nil {
				return accountBalances, orderBalance, msgs, err
			}

			sort.Slice(transactions, func(i, j int) bool {
				return transactions[i].Timestamp.Value() < transactions[j].Timestamp.Value()
			})

			for _, tr := range transactions {
				var sym types.Symbol
				switch tr.Symbol {
				case "BTC":
					sym = types.SymbolBitcoin
				case "ETH":
					sym = types.SymbolEthereum
				case "CMTN":
					sym = types.SymbolCipherMtn
				}

				qty, _ := decimal.NewFromString(tr.Quantity)
				trbals[sym] = trbals[sym].Add(qty)
			}

			orepo := h.accounts.Orders(a)
			orders, err := orepo.GetOrdersByStatus(ctx, persist.StatusFilled, persist.StatusOpen, persist.StatusPartial, persist.StatusCanceled)
			if err != nil {
				return accountBalances, orderBalance, msgs, err
			}

			for _, order := range orders {
				bi := persist.NewBookItem(order.Base)
				exists, err := h.book.BookItemExists(ctx, &bi)

				s, a := order.Base.Type.HoldAmount(order.Base.Action, order.Base.Base, order.Base.Target)
				orderBalance[s] = orderBalance[s].Add(a)

				key := fmt.Sprintf("%s.%d", bi.Order.Type.KeyString(bi.Order.Action), bi.Order.Timestamp.UnixNano())

				if err != nil {
					return accountBalances, orderBalance, msgs, err
				}
				switch order.Status {
				case persist.StatusCanceled, persist.StatusFilled:
					// order should not be on the book
					if exists {
						msgs = append(msgs, fmt.Sprintf("account %s order %s exists on order book with key %s", acc, order.Base.ID, key))
					}
				case persist.StatusOpen, persist.StatusPartial:
					// order should be on the book
					if !exists {
						msgs = append(msgs, fmt.Sprintf("account %s order %s does not exist on order book", acc, order.Base.ID))
					}
				}

				/* TODO: market orders are allowed on the order book if there are no matching orders on the opposite book
				switch order.Base.Type.(type) {
				case *types.MarketOrderType:
					// market orders should not appear on the book
					if exists {
						msgs = append(msgs, fmt.Sprintf("account %s order %s is a market order and is on the order book with key %s", acc, order.Base.ID, key))
					}
				}
				*/
			}

			for _, s := range symbols {
				br := h.accounts.Balances(a, s)

				var bal decimal.Decimal
				bal, err = br.GetBalance(ctx)
				if err != nil {
					return accountBalances, orderBalance, msgs, err
				}

				accountBalances[s] = accountBalances[s].Add(bal)

				if !bal.Equal(trbals[s]) {
					msgs = append(msgs,
						fmt.Sprintf(`account balance '%s' does not match transaction balance '%s' for account %s`,
							bal.StringFixedBank(s.RoundingPlace()),
							trbals[s].StringFixedBank(s.RoundingPlace()),
							acc))
				}
			}
		}
	}

	return accountBalances, orderBalance, msgs, nil
}

type AuditResponse struct {
	UserAccounts   map[string]string            `json:"user_accounts"`
	LedgerAccounts map[string]map[string]string `json:"ledger_accounts"`
	Balance        map[string]string            `json:"balance"`
	Throughput     map[string]string            `json:"order_throughput"`
	Errors         []string                     `json:"errors"`
}

// Render implements the render.Renderer interface for use with chi-router
func (a *AuditResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
