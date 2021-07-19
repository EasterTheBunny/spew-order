package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/internal/persist/firebase"
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/olekukonko/tablewriter"
)

var (
	projectID = flag.String("project", "", "Google project id.")
	account   = flag.String("account", "", "account id")
	orderid   = flag.String("order", "", "order id")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	client, err := firestore.NewClient(ctx, *projectID)
	if err != nil {
		panic(err)
	}

	var order *persist.Order

	switch {
	case *account != "" && *orderid != "":
		log.Printf("searching by account %s", *account)
		arepo := firebase.NewAccountRepository(client)
		orepo := arepo.Orders(&persist.Account{ID: *account})

		order, err = orepo.GetOrder(ctx, ky(*orderid))
		if err != nil {
			log.Println(err)
			return
		}
	case *orderid != "":
		arepo := firebase.NewAuthorizationRepository(client)

		var auths []*persist.Authorization
		auths, err = arepo.GetAuthorizations(ctx)
		if err != nil {
			log.Println(err)
			return
		}

		// calculate totals for all symbols for all accounts
	AuthLoop:
		for _, auth := range auths {
			for _, acc := range auth.Accounts {
				a := &persist.Account{ID: acc}

				repo := firebase.NewAccountRepository(client).Orders(a)

				var orders []*persist.Order
				orders, err = repo.GetOrdersByStatus(ctx, persist.StatusCanceled, persist.StatusFilled, persist.StatusOpen, persist.StatusPartial)
				if err != nil {
					log.Println(err)
					return
				}

				for _, o := range orders {
					if o.Base.ID.String() == *orderid {
						order = o
						account = &acc
						break AuthLoop
					}
				}
			}
		}
	}

	if order != nil {
		printOrderDetail(*account, order)
		printOrderTransactions(order.Transactions)
	}
}

func printOrderDetail(accountid string, order *persist.Order) {

	var symbol types.Symbol
	var price string
	var quantity string

	switch t := order.Base.Type.(type) {
	case *types.MarketOrderType:
		symbol = t.Base
		quantity = t.Quantity.StringFixedBank(t.Base.RoundingPlace())
	case *types.LimitOrderType:
		symbol = t.Base
		price = t.Price.StringFixedBank(t.Base.RoundingPlace())
		quantity = t.Quantity.StringFixedBank(t.Base.RoundingPlace())
	}

	fmt.Println("")
	fmt.Printf("----------- order for %s -----------", accountid)
	fmt.Println("")

	data := [][]string{
		{"Status", order.Status.String()},
		{"Market", fmt.Sprintf("%s-%s", order.Base.Base, order.Base.Target)},
		{"Type", order.Base.Type.Name()},
		{"Action", order.Base.Action.String()},
		{"Hold", order.Base.HoldID},
		{"Symbol", symbol.String()},
		{"Price", price},
		{"Quantity", quantity},
		{"Parts", fmt.Sprintf("%d", len(order.Transactions))},
		{"Created", fmt.Sprintf("%d", order.Base.Timestamp.UnixNano())},
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader([]string{"Property", "Value"})
	table.SetBorder(false)
	table.AppendBulk(data)
	table.Render()

	fmt.Println("")
}

func printOrderTransactions(data [][]string) {

	fmt.Println("")
	fmt.Println("----------- transactions -----------")
	fmt.Println("")

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader([]string{"Timestamp", "Fee", "Add Amount", "Sub Amount"})
	table.SetBorder(false)
	table.AppendBulk(data)
	table.Render()

	fmt.Println("")
}

type ky string

func (k ky) String() string {
	return string(k)
}
