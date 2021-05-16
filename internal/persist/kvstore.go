package persist

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

var ErrObjectNotExist = fmt.Errorf("object not found for key")

// KVStore ...
type KVStore interface {
	Get(string) ([]byte, error)
	Attrs(string) (*KVStoreObjectAttrs, error)
	Set(string, []byte, *KVStoreObjectAttrsToUpdate) error
	Delete(string) error
	RangeGet(*KVStoreQuery, int) ([]*KVStoreObjectAttrs, error)
}

type KVStoreQuery struct {
	Prefix      string
	StartOffset string
}

type KVStoreObjectAttrs struct {
	Name            string
	Metadata        map[string]string
	ContentEncoding string
	Created         time.Time
}

type KVStoreObjectAttrsToUpdate struct {
	ContentEncoding string
	Metadata        map[string]string
}

func NewGoogleKVStore(bucket *string) (*GoogleKVStore, error) {

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	if bucket == nil {
		bucket = &StorageBucket
	}

	return &GoogleKVStore{client: client, bucket: *bucket}, nil
}

type GoogleKVStore struct {
	client *storage.Client
	bucket string
}

// Get ...
func (gs *GoogleKVStore) Get(sKey string) ([]byte, error) {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, storeTimeout)
	defer cancel()

	rc, err := gs.client.Bucket(gs.bucket).Object(sKey).NewReader(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			err = ErrObjectNotExist
		}
		return []byte{}, err
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return []byte{}, err
	}

	return data, nil
}

// Get ...
func (gs *GoogleKVStore) Attrs(sKey string) (attrs *KVStoreObjectAttrs, err error) {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, storeTimeout)
	defer cancel()

	var obj *storage.ObjectAttrs
	obj, err = gs.client.Bucket(gs.bucket).Object(sKey).Attrs(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			err = ErrObjectNotExist
		}
		return
	}

	attrs = &KVStoreObjectAttrs{
		Name:            obj.Name,
		Metadata:        obj.Metadata,
		ContentEncoding: obj.ContentEncoding,
		Created:         obj.Created}

	return
}

// Set ...
func (gs *GoogleKVStore) Set(sKey string, data []byte, attrs *KVStoreObjectAttrsToUpdate) error {
	ctx := context.Background()
	value := bytes.NewReader(data)
	ctx, cancel := context.WithTimeout(ctx, storeTimeout)
	defer cancel()

	handle := gs.client.Bucket(gs.bucket).Object(sKey)

	wc := handle.NewWriter(ctx)
	if _, err := io.Copy(wc, value); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}

	if attrs != nil {
		at := &storage.ObjectAttrsToUpdate{
			ContentEncoding: attrs.ContentEncoding,
			Metadata:        attrs.Metadata}
		handle.Update(ctx, *at)
	}

	return nil
}

// RangeGet ...
func (gs *GoogleKVStore) RangeGet(q *KVStoreQuery, limit int) ([]*KVStoreObjectAttrs, error) {
	// bucket := "bucket-name"
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	var qu *storage.Query
	if q != nil {
		qu = &storage.Query{
			StartOffset: q.StartOffset}
	}
	qu.SetAttrSelection([]string{"Name", "MetaData", "Created", "ContentEncoding"})

	it := gs.client.Bucket(gs.bucket).Objects(ctx, qu)
	attr := []*KVStoreObjectAttrs{}
	var cnt int
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return attr, err
		}

		var a *KVStoreObjectAttrs
		if attrs != nil {
			a = &KVStoreObjectAttrs{
				Name:            attrs.Name,
				Metadata:        attrs.Metadata,
				ContentEncoding: attrs.ContentEncoding,
				Created:         attrs.Created}
		}

		attr = append(attr, a)

		cnt++
		if limit > 0 && cnt >= limit {
			break
		}
	}
	return attr, nil
}

// Delete ...
func (gs *GoogleKVStore) Delete(sKey string) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, storeTimeout)
	defer cancel()
	return gs.client.Bucket(gs.bucket).Object(sKey).Delete(ctx)
}

// NewMockKVStore ...
func NewMockKVStore() *MockKVStore {
	return &MockKVStore{
		data: make(map[string][]byte),
		meta: make(map[string]*KVStoreObjectAttrs)}
}

// MockKVStore ...
type MockKVStore struct {
	key  []string
	data map[string][]byte
	meta map[string]*KVStoreObjectAttrs
}

// Get ...
func (gsm *MockKVStore) Get(key string) ([]byte, error) {
	if _, ok := gsm.data[key]; !ok {
		return []byte{}, ErrObjectNotExist
	}
	return gsm.data[key], nil
}

func (gsm *MockKVStore) Attrs(key string) (a *KVStoreObjectAttrs, err error) {
	a, ok := gsm.meta[key]
	if !ok {
		err = ErrObjectNotExist
	}
	return
}

// Set ...
func (gsm *MockKVStore) Set(key string, b []byte, attrs *KVStoreObjectAttrsToUpdate) error {

	sAttrs := &KVStoreObjectAttrs{
		Name:    key,
		Created: time.Now()}

	if attrs != nil {
		sAttrs.Metadata = attrs.Metadata
	}

	if _, ok := gsm.meta[key]; ok {
		sAttrs = gsm.meta[key]

		if attrs != nil {
			sAttrs.Metadata = attrs.Metadata
		}
	} else {
		gsm.key = append(gsm.key, key)
		sort.Strings(gsm.key)
	}

	gsm.data[key] = b
	gsm.meta[key] = sAttrs
	return nil
}

// Delete ...
func (gsm *MockKVStore) Delete(key string) error {
	for i, k := range gsm.key {
		if k == key {
			gsm.key = append(gsm.key[:i], gsm.key[i+1:]...)
			break
		}
	}
	delete(gsm.data, key)
	delete(gsm.meta, key)
	return nil
}

// RangeGet ...
func (gsm *MockKVStore) RangeGet(q *KVStoreQuery, limit int) (attrs []*KVStoreObjectAttrs, err error) {
	var cnt int
	var qry string

	if q.Prefix != "" {
		qry = q.Prefix
	}

	if q.StartOffset != "" {
		qry = q.StartOffset
	}

	for _, k := range gsm.key {
		if strings.HasPrefix(k, qry) {
			attrs = append(attrs, gsm.meta[k])

			cnt++
			if limit > 0 && cnt >= limit {
				break
			}
		}
	}
	return
}

// Len ...
func (gsm *MockKVStore) Len() int {
	return len(gsm.key)
}
