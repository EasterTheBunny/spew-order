package main

import (
	"context"
	"flag"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/internal/persist/firebase"
)

var (
	account   = flag.String("account", "spew", "Environment naming prefix.")
	projectID = flag.String("project", "", "Google project id.")
	delete    = flag.Bool("delete", false, "delete book items")
)

func main() {
	flag.Parse()
	ctx := context.Background()

	client, err := firestore.NewClient(ctx, *projectID)
	if err != nil {
		panic(err)
	}

	arepo := firebase.NewAccountRepository(client)
	orepo := arepo.Orders(&persist.Account{ID: *account})
	brepo := firebase.NewBookRepository(client)

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
