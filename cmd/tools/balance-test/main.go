package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/internal/persist/firebase"
	"github.com/easterthebunny/spew-order/pkg/domain"
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/shopspring/decimal"
)

var (
	projectID = flag.String("project", "", "Google project id.")
)

func main() {
	flag.Parse()

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, *projectID)
	if err != nil {
		log.Println(err)
		panic(err)
	}

	arepo := firebase.NewAccountRepository(client)
	acct, err := createAccount(ctx, arepo)
	if err != nil {
		log.Println(err)
		panic(err)
	}

	var count int
	var amt float64
	c := time.Tick(100 * time.Millisecond)
	for range c {
		count++
		if count > 20 {
			break
		}

		if count%2 == 0 {
			amt = 1.1
		} else {
			amt = -1.1
		}

		switch count {
		case 5:
			err = setHold(ctx, arepo, acct, 2.2)
			if err != nil {
				panic(err)
			}
		case 10:
			err = updateHold(ctx, arepo, acct, 2.2)
			if err != nil {
				panic(err)
			}
		case 15:
			err = deleteHold(ctx, arepo, acct)
			if err != nil {
				panic(err)
			}
		}

		err = updateBalance(ctx, arepo, acct, amt)
		if err != nil {
			panic(err)
		}

		if count%5 == 0 {
			err = printBalance(ctx, arepo, acct)
			if err != nil {
				panic(err)
			}
		}
	}

	<-time.After(2 * time.Second)

	fmt.Println("final balance:")
	err = printBalance(ctx, arepo, acct)
	if err != nil {
		panic(err)
	}
}

func createAccount(ctx context.Context, repo persist.AccountRepository) (*domain.Account, error) {
	a := domain.NewAccount()
	fmt.Printf("%s\n", a.ID)
	return a, repo.Save(ctx, &persist.Account{ID: a.ID.String()})
}

func updateBalance(ctx context.Context, repo persist.AccountRepository, acct *domain.Account, amt float64) error {
	brepo := repo.Balances(&persist.Account{ID: acct.ID.String()}, types.SymbolBitcoin)
	xrepo := repo.Balances(&persist.Account{ID: acct.ID.String()}, types.SymbolEthereum)

	err := brepo.AddToBalance(ctx, decimal.NewFromFloat(amt))
	if err != nil {
		return err
	}

	err = xrepo.AddToBalance(ctx, decimal.NewFromFloat(amt))
	if err != nil {
		return err
	}

	return nil
}

func printBalance(ctx context.Context, repo persist.AccountRepository, acct *domain.Account) error {
	brepo := repo.Balances(&persist.Account{ID: acct.ID.String()}, types.SymbolBitcoin)
	amt, err := brepo.GetBalance(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("BTC: %s\n", amt.StringFixedBank(8))
	holds, err := brepo.FindHolds(ctx)
	if err != nil {
		return err
	}
	for _, hold := range holds {
		amt = amt.Add(hold.Amount)
	}
	fmt.Printf("BTC (with holds): %s\n", amt.StringFixedBank(8))

	xrepo := repo.Balances(&persist.Account{ID: acct.ID.String()}, types.SymbolEthereum)
	amt, err = xrepo.GetBalance(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("ETH: %s\n", amt.StringFixedBank(8))
	holds, err = xrepo.FindHolds(ctx)
	if err != nil {
		return err
	}
	for _, hold := range holds {
		amt = amt.Add(hold.Amount)
	}
	fmt.Printf("ETH (with holds): %s\n", amt.StringFixedBank(8))

	return nil
}

func setHold(ctx context.Context, repo persist.AccountRepository, acct *domain.Account, amt float64) error {
	brepo := repo.Balances(&persist.Account{ID: acct.ID.String()}, types.SymbolBitcoin)
	xrepo := repo.Balances(&persist.Account{ID: acct.ID.String()}, types.SymbolEthereum)

	err := brepo.CreateHold(ctx, &persist.BalanceItem{ID: acct.ID.String(), Amount: decimal.NewFromFloat(amt)})
	if err != nil {
		return err
	}

	err = xrepo.CreateHold(ctx, &persist.BalanceItem{ID: acct.ID.String(), Amount: decimal.NewFromFloat(amt)})
	if err != nil {
		return err
	}

	return nil
}

func updateHold(ctx context.Context, repo persist.AccountRepository, acct *domain.Account, amt float64) error {
	brepo := repo.Balances(&persist.Account{ID: acct.ID.String()}, types.SymbolBitcoin)
	xrepo := repo.Balances(&persist.Account{ID: acct.ID.String()}, types.SymbolEthereum)

	err := brepo.UpdateHold(ctx, acct.ID, decimal.NewFromFloat(amt))
	if err != nil {
		return err
	}

	err = xrepo.UpdateHold(ctx, acct.ID, decimal.NewFromFloat(amt))
	if err != nil {
		return err
	}

	return nil
}

func deleteHold(ctx context.Context, repo persist.AccountRepository, acct *domain.Account) error {
	brepo := repo.Balances(&persist.Account{ID: acct.ID.String()}, types.SymbolBitcoin)
	xrepo := repo.Balances(&persist.Account{ID: acct.ID.String()}, types.SymbolEthereum)

	err := brepo.DeleteHold(ctx, acct.ID)
	if err != nil {
		return err
	}

	err = xrepo.DeleteHold(ctx, acct.ID)
	if err != nil {
		return err
	}

	return nil
}
