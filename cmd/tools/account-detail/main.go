package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"cloud.google.com/go/firestore"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/internal/persist/firebase"
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/olekukonko/tablewriter"
	"github.com/shopspring/decimal"
)

var (
	projectID = flag.String("project", "", "Google project id.")
	account   = flag.String("account", "", "account uuid")
)

func main() {
	flag.Parse()

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, *projectID)
	if err != nil {
		panic(err)
	}

	arepo := firebase.NewAccountRepository(client)

	fmt.Println(" -------- Incomplete Orders -----------")
	fmt.Println("")
	printOpenOrders(ctx, arepo)
	fmt.Println("")
	fmt.Println(" --------- Complete Orders ------------")
	fmt.Println("")
	printClosedOrders(ctx, arepo)
	fmt.Println("")
	fmt.Println("")
	printTransactions(ctx, arepo)
	fmt.Println("")
	fmt.Println("")
}

func printOpenOrders(ctx context.Context, repo persist.AccountRepository) {
	a := &persist.Account{ID: *account}

	orders, err := repo.Orders(a).GetOrdersByStatus(ctx, persist.StatusOpen, persist.StatusPartial)
	if err != nil {
		log.Println(err)
	}

	printOrders(ctx, orders, repo, a)
}

func printClosedOrders(ctx context.Context, repo persist.AccountRepository) {
	a := &persist.Account{ID: *account}

	orders, err := repo.Orders(a).GetOrdersByStatus(ctx, persist.StatusCanceled, persist.StatusFilled)
	if err != nil {
		log.Println(err)
	}

	printOrders(ctx, orders, repo, a)
}

func printOrders(ctx context.Context, orders []*persist.Order, repo persist.AccountRepository, a *persist.Account) {

	data := [][]string{}

	for _, order := range orders {

		var symbol types.Symbol
		var price string
		var quantity string
		var holdAmt string

		switch t := order.Base.Type.(type) {
		case *types.MarketOrderType:
			symbol = t.Base
			quantity = t.Quantity.StringFixedBank(t.Base.RoundingPlace())
		case *types.LimitOrderType:
			symbol = t.Base
			price = t.Price.StringFixedBank(t.Base.RoundingPlace())
			quantity = t.Quantity.StringFixedBank(t.Base.RoundingPlace())
		}

		holdAmt = order.Base.HoldID

		bals := repo.Balances(a, symbol)
		holds, err := bals.FindHolds(ctx)
		if err != nil {
			panic(err)
		}

		for _, hold := range holds {
			if hold.ID == holdAmt {
				holdAmt = hold.Amount.StringFixedBank(symbol.RoundingPlace())
				break
			}
		}

		line := []string{
			order.Base.ID.String(),
			order.Status.String(),
			fmt.Sprintf("%s-%s", order.Base.Base, order.Base.Target),
			order.Base.Type.Name(),
			order.Base.Action.String(),
			holdAmt,
			symbol.String(),
			price,
			quantity,
			fmt.Sprintf("%d", len(order.Transactions)),
			fmt.Sprintf("%d", order.Base.Timestamp.UnixNano()),
		}

		data = append(data, line)
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i][9] < data[j][9]
	})

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Status", "Market", "Type", "Action", "Hold", "Symbol", "Price", "Quantity", "Parts", "Created"})
	//table.SetFooter([]string{"", "", "Total", "$146.93"}) // Add Footer
	table.SetBorder(false) // Set Border to false
	table.AppendBulk(data) // Add Bulk Data
	table.Render()
}

func printTransactions(ctx context.Context, repo persist.AccountRepository) {
	a := &persist.Account{ID: *account}

	balances := map[string]decimal.Decimal{
		"BTC": decimal.NewFromInt(0),
		"ETH": decimal.NewFromInt(0),
	}

	btcBal, _ := repo.Balances(a, types.SymbolBitcoin).GetBalance(ctx)
	ethBal, _ := repo.Balances(a, types.SymbolEthereum).GetBalance(ctx)

	transactions, err := repo.Transactions(a).GetTransactions(ctx)
	if err != nil {
		log.Println(err)
		return
	}

	data := [][]string{}

	for _, tr := range transactions {

		qty, _ := decimal.NewFromString(tr.Quantity)
		balances[tr.Symbol] = balances[tr.Symbol].Add(qty)

		line := []string{
			string(tr.Type),
			tr.Symbol,
			tr.Quantity,
			tr.Fee,
			tr.OrderID,
			tr.Timestamp.String(),
		}

		data = append(data, line)
	}

	sort.Slice(data, func(i, j int) bool {
		if data[i][1] == data[j][1] {
			return data[i][5] < data[j][5]
		}

		return data[i][1] == data[j][1]
	})

	data = append(data, []string{"balance", "BTC", btcBal.StringFixedBank(types.SymbolBitcoin.RoundingPlace()), "", "", ""})
	data = append(data, []string{"balance", "ETH", ethBal.StringFixedBank(types.SymbolEthereum.RoundingPlace()), "", "", ""})

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Type", "Symbol", "Quantity", "Fee", "Order ID", "Created"})
	table.SetFooter([]string{"", "BTC", balances["BTC"].StringFixedBank(types.SymbolBitcoin.RoundingPlace()), "ETH", balances["ETH"].StringFixedBank(types.SymbolEthereum.RoundingPlace()), ""}) // Add Footer
	table.SetBorder(false)                                                                                                                                                                       // Set Border to false
	table.AppendBulk(data)                                                                                                                                                                       // Add Bulk Data
	table.Render()
}
