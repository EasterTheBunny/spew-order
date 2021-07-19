package persist

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/googleapi"
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
	Updated         time.Time
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
	fmt.Printf("get key: %s\n", sKey)
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, storeTimeout)
	defer cancel()

	rc, err := gs.client.Bucket(gs.bucket).Object(sKey).NewReader(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			fmt.Println("--- key doesn't exist")
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
		Created:         obj.Created,
		Updated:         obj.Updated}

	return
}

// Set ...
func (gs *GoogleKVStore) Set(sKey string, data []byte, attrs *KVStoreObjectAttrsToUpdate) error {
	fmt.Printf("set key: %s\n", sKey)
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, storeTimeout)
	defer cancel()

	var doOp = func(b, k string) error {

		handle := gs.client.Bucket(b).Object(k)

		wc := handle.NewWriter(ctx)
		if _, err := io.Copy(wc, bytes.NewReader(data)); err != nil {
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

	min := 0
	max := 1000
	deadline := 5 * time.Second
	cnt := 1

	err := doOp(gs.bucket, sKey)
	for err != nil {
		fmt.Println("exponential backoff")

		if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
			// exponential backoff
			rand.Seed(time.Now().UnixNano())
			wait := rand.Intn(max-min+1) + min
			wait = int(math.Pow(float64(2), float64(cnt))) + wait
			<-time.After(time.Duration(wait))
			err = doOp(gs.bucket, sKey)

			cnt++
			if deadline <= time.Duration(wait) {
				return err
			}
		}

		if e, ok := err.(*googleapi.Error); ok {
			if e.Code == http.StatusTooManyRequests || e.Code >= http.StatusInternalServerError {
				// exponential backoff
				rand.Seed(time.Now().UnixNano())
				wait := rand.Intn(max-min+1) + min
				wait = int(math.Pow(float64(2), float64(cnt))) + wait

				// wait and execute again
				<-time.After(time.Duration(wait))
				err = doOp(gs.bucket, sKey)

				cnt++
				if deadline <= time.Duration(wait) {
					return err
				}
			} else {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

// RangeGet returns a set of size `limit`. Set `limit` to 0 for no limit.
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
				Created:         attrs.Created,
				Updated:         attrs.Updated}
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

	var doOp = func() error {
		return gs.client.Bucket(gs.bucket).Object(sKey).Delete(ctx)
	}

	min := 0
	max := 1000
	deadline := 5 * time.Second
	cnt := 1

	err := doOp()
	for err != nil {

		if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
			// exponential backoff
			rand.Seed(time.Now().UnixNano())
			wait := rand.Intn(max-min+1) + min
			wait = int(math.Pow(float64(2), float64(cnt))) + wait
			<-time.After(time.Duration(wait))
			err = doOp()

			cnt++
			if deadline <= time.Duration(wait) {
				return err
			}
		}

		if e, ok := err.(*googleapi.Error); ok {
			if e.Code == http.StatusTooManyRequests || e.Code >= http.StatusInternalServerError {
				// exponential backoff
				rand.Seed(time.Now().UnixNano())
				wait := rand.Intn(max-min+1) + min
				wait = int(math.Pow(float64(2), float64(cnt))) + wait
				<-time.After(time.Duration(wait))
				err = doOp()

				cnt++
				if deadline <= time.Duration(wait) {
					return err
				}
			} else {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

// NewMockKVStore ...
func NewMockKVStore() *MockKVStore {
	return &MockKVStore{
		data:     make(map[string][]byte),
		meta:     make(map[string]*KVStoreObjectAttrs),
		logLevel: 0}
}

// MockKVStore ...
type MockKVStore struct {
	key      []string
	data     map[string][]byte
	meta     map[string]*KVStoreObjectAttrs
	logLevel int
}

// Get ...
func (gsm *MockKVStore) Get(key string) ([]byte, error) {
	if gsm.logLevel > 0 {
		log.Printf("GET %s", key)
	}
	if _, ok := gsm.data[key]; !ok {
		return []byte{}, ErrObjectNotExist
	}
	return gsm.data[key], nil
}

func (gsm *MockKVStore) Attrs(key string) (a *KVStoreObjectAttrs, err error) {
	if gsm.logLevel > 0 {
		log.Printf("ATTRS %s", key)
	}
	a, ok := gsm.meta[key]
	if !ok {
		err = ErrObjectNotExist
	}
	return
}

// Set ...
func (gsm *MockKVStore) Set(key string, b []byte, attrs *KVStoreObjectAttrsToUpdate) error {
	if gsm.logLevel > 0 {
		log.Printf("SET %s", key)
	}
	sAttrs := &KVStoreObjectAttrs{
		Name:    key,
		Created: time.Now(),
		Updated: time.Now()}

	if attrs != nil {
		sAttrs.Metadata = attrs.Metadata
	}

	if m, ok := gsm.meta[key]; ok {
		sAttrs = gsm.meta[key]

		if attrs != nil {
			sAttrs.Metadata = attrs.Metadata
			sAttrs.Created = m.Created
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
	if gsm.logLevel > 0 {
		log.Printf("DELETE %s", key)
	}
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
	if gsm.logLevel > 0 {
		log.Printf("RANGE_GET %s", qry)
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
