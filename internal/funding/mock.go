package funding

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/easterthebunny/spew-order/pkg/types"
	uuid "github.com/satori/go.uuid"
)

func NewMockSource() Source {
	return &mockSource{}
}

type mockSource struct {
}

func (s *mockSource) Name() string {
	return "MOCK"
}

func (s *mockSource) Supports(types.Symbol) bool {
	return true
}

func (s *mockSource) Callback() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			b, _ := ioutil.ReadAll(r.Body)
			var tr Transaction
			var cerr *CallbackError
			err := json.Unmarshal(b, &tr)
			if err != nil {
				fmt.Println(err)
				cerr = &CallbackError{
					Status: http.StatusBadRequest,
					Err:    err,
				}
			}

			ctx = attachToContext(ctx, tr, cerr)
			next.ServeHTTP(w, r.WithContext(ctx))

		})
	}
}

func (s *mockSource) CreateAddress(types.Symbol) (*Address, error) {
	id := uuid.NewV4()

	hash := hmac.New(sha256.New, []byte("secret"))
	_, err := io.WriteString(hash, id.String())
	if err != nil {
		return nil, err
	}
	encoded := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	return &Address{ID: id.String(), Hash: encoded}, nil
}

func (s *mockSource) Withdraw(*Transaction) (string, error) {

	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	randSeq := func(n int) string {
		b := make([]rune, n)
		for i := range b {
			b[i] = letters[rand.Intn(len(letters))]
		}
		return string(b)
	}

	rand.Seed(time.Now().UnixNano())

	hash := hmac.New(sha256.New, []byte("secret"))
	_, err := io.WriteString(hash, randSeq(30))
	if err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	return encoded, nil
}

func (s *mockSource) OKResponse() int {
	return http.StatusOK
}
