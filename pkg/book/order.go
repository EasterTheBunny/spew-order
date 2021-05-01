package book

import (
	"context"
	"log"

	"cloud.google.com/go/storage"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
)

type OrderBook interface {
	ExecuteOrInsertOrder(order types.Order) error
}

func NewGoogleOrderBook(bucket string) OrderBook {

	// set the primary storage bucket
	persist.StorageBucket = bucket

	storageClient, err := storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("storage.NewClient: %v", err)
	}

	kvStore := persist.NewGoogleStorageAPI(storageClient)

	return persist.NewGoogleStorage(kvStore)
}

func NewMockOrderBook() OrderBook {
	return persist.NewGoogleStorage(persist.NewGoogleStorageMock())
}
