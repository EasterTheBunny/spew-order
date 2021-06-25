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

		fmt.Printf("status: %s; exists on book: %v\n", order.Status, result)
		if order.Status == persist.StatusCanceled && result {
			err = brepo.DeleteBookItem(ctx, &bitem)
			if err != nil {
				panic(err)
			}
		}
	}
}
