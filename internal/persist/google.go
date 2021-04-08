package persist

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"time"

	"github.com/easterthebunny/spew-order/internal/key"
	"google.golang.org/api/iterator"

	"cloud.google.com/go/storage"
)

var (
	gsRoot = key.FromBytes([]byte{0xFE})
	gsBook = gsRoot.Sub("book")
)

var (
	// ErrStoreConnectionError ...
	ErrStoreConnectionError = errors.New("connection issue to data store")
)

// IGoogleStorage ...
type IGoogleStorage interface {
	Get(string) ([]byte, error)
	Set(string, []byte, *storage.ObjectAttrsToUpdate) error
	Delete(string) error
	RangeGet(*storage.Query, int) ([]*storage.ObjectAttrs, error)
}

// NewGoogleStorage ...
func NewGoogleStorage(s IGoogleStorage) *GoogleStorage {
	return &GoogleStorage{store: s}
}

// GoogleStorage ...
type GoogleStorage struct {
	store IGoogleStorage
}

// NewGoogleStorageAPI ...
func NewGoogleStorageAPI(c *storage.Client) *GoogleStorageAPI {
	return &GoogleStorageAPI{Client: c}
}

// GoogleStorageAPI ...
type GoogleStorageAPI struct {
	Client *storage.Client
}

// Get ...
func (gs *GoogleStorageAPI) Get(sKey string) ([]byte, error) {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, storeTimeout)
	defer cancel()

	rc, err := gs.Client.Bucket(StorageBucket).Object(sKey).NewReader(ctx)
	if err != nil {
		return []byte{}, err
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return []byte{}, err
	}

	return data, nil
}

// Set ...
func (gs *GoogleStorageAPI) Set(sKey string, data []byte, attrs *storage.ObjectAttrsToUpdate) error {
	ctx := context.Background()
	value := bytes.NewReader(data)
	ctx, cancel := context.WithTimeout(ctx, storeTimeout)
	defer cancel()

	handle := gs.Client.Bucket(StorageBucket).Object(sKey)

	wc := handle.NewWriter(ctx)
	if _, err := io.Copy(wc, value); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}

	if attrs != nil {
		handle.Update(ctx, *attrs)
	}

	return nil
}

// RangeGet ...
func (gs *GoogleStorageAPI) RangeGet(q *storage.Query, limit int) ([]*storage.ObjectAttrs, error) {
	// bucket := "bucket-name"
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	it := gs.Client.Bucket(StorageBucket).Objects(ctx, q)
	attr := []*storage.ObjectAttrs{}
	var cnt int
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return attr, err
		}

		attr = append(attr, attrs)

		cnt++
		if limit > 0 && cnt >= limit {
			break
		}
	}
	return attr, nil
}

// Delete ...
func (gs *GoogleStorageAPI) Delete(sKey string) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, storeTimeout)
	defer cancel()
	return gs.Client.Bucket(StorageBucket).Object(sKey).Delete(ctx)
}
