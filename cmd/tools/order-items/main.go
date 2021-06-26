package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/internal/persist/firebase"
)

var (
	projectID = flag.String("project", "", "Google project id.")
	account   = flag.String("account", "spew", "Environment naming prefix.")
	orderid   = flag.String("order", "", "")
)

func main() {
	flag.Parse()
	fmt.Printf("account: %s\n", *account)

	ctx := context.Background()

	client, err := firestore.NewClient(ctx, *projectID)
	if err != nil {
		panic(err)
	}

	arepo := firebase.NewAccountRepository(client)
	orepo := arepo.Orders(&persist.Account{ID: *account})

	order, err := orepo.GetOrder(ctx, ky(*orderid))
	if err != nil {
		panic(err)
	}

	err = orepo.UpdateOrderStatus(ctx, ky(*orderid), persist.StatusPartial, []string{"a", "b", "c", "d"})
	if err != nil {
		panic(err)
	}

	b, err := json.MarshalIndent(order, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", string(b))
}

type ky string

func (k ky) String() string {
	return string(k)
}
