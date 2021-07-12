package main

import (
	"context"
	"flag"
	"log"

	"cloud.google.com/go/firestore"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/internal/persist/firebase"
	"github.com/easterthebunny/spew-order/pkg/types"
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

	repo := firebase.NewBookRepository(client)
	var items []*persist.BookItem

	for i := 0; i < 10; i++ {
		order := types.NewOrder()
		order.OrderRequest = types.OrderRequest{
			Action: types.ActionTypeSell,
			Type:   &types.MarketOrderType{},
		}

		item := persist.NewBookItem(order)
		items = append(items, &item)

		for j := 0; j < 10; j++ {
			err := repo.SetBookItem(ctx, &item)
			if err != nil {
				log.Println(err)
			}
		}
	}

	order := types.NewOrder()
	order.OrderRequest = types.OrderRequest{
		Action: types.ActionTypeBuy,
		Type:   &types.MarketOrderType{},
	}
	item := persist.NewBookItem(order)

	batch, err := repo.GetHeadBatch(ctx, &item, 5)
	if err != nil {
		log.Println(err)
	}

	log.Printf("batch size %d; expected 5", len(batch))

	batch, err = repo.GetHeadBatch(ctx, &item, 10)
	if err != nil {
		log.Println(err)
	}

	log.Printf("batch size %d; expected 10", len(batch))

	for i := 0; i < 3; i++ {
		err := repo.DeleteBookItem(ctx, items[i])
		if err != nil {
			log.Println(err)
		}
	}

	batch, err = repo.GetHeadBatch(ctx, &item, 10)
	if err != nil {
		log.Println(err)
	}

	log.Printf("batch size %d; expected 7", len(batch))
}
