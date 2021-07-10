package main

import (
	"context"
	"flag"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/internal/persist/firebase"
	"github.com/easterthebunny/spew-order/pkg/types"
)

var (
	projectID = flag.String("project", "", "Google project id.")
	account   = flag.String("account", "", "account uuid")
	delete    = flag.Bool("delete", false, "delete book items")
	print     = flag.Bool("print", false, "print order book")
)

func main() {
	flag.Parse()
	ctx := context.Background()

	client, err := firestore.NewClient(ctx, *projectID)
	if err != nil {
		panic(err)
	}

	arepo := firebase.NewAccountRepository(client)
	brepo := firebase.NewBookRepository(client)

	if *print {
		item := persist.BookItem{
			Order: types.Order{
				OrderRequest: types.OrderRequest{
					Base:   types.SymbolBitcoin,
					Target: types.SymbolEthereum,
				},
			},
			ActionType: types.ActionTypeBuy,
		}

		fmt.Printf("%s\n", item.ActionType.String())
		head, _ := brepo.GetHeadBatch(ctx, &item, 50)
		for _, h := range head {
			switch j := h.Order.Type.(type) {
			case *types.MarketOrderType:
				fmt.Printf("MARKET:%s %30s %s\n", j.Base, j.Quantity, h.Order.ID)
			case *types.LimitOrderType:
				fmt.Printf("%-10s %30s %s\n", j.Price, j.Quantity, h.Order.ID)
			}
		}

		item.ActionType = types.ActionTypeSell

		fmt.Printf("%s\n", item.ActionType.String())
		head, _ = brepo.GetHeadBatch(ctx, &item, 50)
		for _, h := range head {
			switch j := h.Order.Type.(type) {
			case *types.MarketOrderType:
				fmt.Printf("MARKET:%s %30s %s\n", j.Base, j.Quantity, h.Order.ID)
			case *types.LimitOrderType:
				fmt.Printf("%-10s %30s %s\n", j.Price, j.Quantity, h.Order.ID)
			}
		}
	}

	if *account != "" {
		orepo := arepo.Orders(&persist.Account{ID: *account})

		orders, err := orepo.GetOrdersByStatus(ctx, persist.StatusOpen, persist.StatusPartial, persist.StatusCanceled)
		if err != nil {
			panic(err)
		}

		for _, order := range orders {
			bitem := persist.NewBookItem(order.Base)
			result, err := brepo.BookItemExists(ctx, &bitem)
			if err != nil {
				panic(err)
			}

			fmt.Printf("action %s; status: %s; exists on book: %v\n", order.Base.Action, order.Status, result)
			if order.Status == persist.StatusCanceled && result && *delete {
				err = brepo.DeleteBookItem(ctx, &bitem)
				if err != nil {
					panic(err)
				}
			}

			if order.Status == persist.StatusOpen || order.Status == persist.StatusPartial {
				items, err := brepo.GetHeadBatch(ctx, &bitem, 10)
				if err != nil {
					panic(err)
				}

				fmt.Printf("%d items found\n", len(items))
				for _, item := range items {
					fmt.Printf("%v\n", item)
				}
			}
		}
	}
}
