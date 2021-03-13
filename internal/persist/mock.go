package persist

import (
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/storage"
)

// NewGoogleStorageMock ...
func NewGoogleStorageMock() *GoogleStorageMock {
	return &GoogleStorageMock{
		data: make(map[string][]byte),
		meta: make(map[string]*storage.ObjectAttrs)}
}

// GoogleStorageMock ...
type GoogleStorageMock struct {
	key  []string
	data map[string][]byte
	meta map[string]*storage.ObjectAttrs
}

// Get ...
func (gsm *GoogleStorageMock) Get(key string) ([]byte, error) {
	if _, ok := gsm.data[key]; !ok {
		return []byte{}, storage.ErrObjectNotExist
	}
	return gsm.data[key], nil
}

// Set ...
func (gsm *GoogleStorageMock) Set(key string, b []byte, attrs *storage.ObjectAttrsToUpdate) error {

	sAttrs := &storage.ObjectAttrs{
		Name:     key,
		Metadata: attrs.Metadata,
		Created:  time.Now()}

	if _, ok := gsm.meta[key]; ok {
		sAttrs = gsm.meta[key]
		sAttrs.Metadata = attrs.Metadata
	} else {
		gsm.key = append(gsm.key, key)
		sort.Sort(sort.StringSlice(gsm.key))
	}

	gsm.data[key] = b
	gsm.meta[key] = sAttrs
	return nil
}

// Delete ...
func (gsm *GoogleStorageMock) Delete(key string) error {
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
func (gsm *GoogleStorageMock) RangeGet(q *storage.Query, limit int) (attrs []*storage.ObjectAttrs, err error) {
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
func (gsm *GoogleStorageMock) Len() int {
	return len(gsm.key)
}
