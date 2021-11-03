package funding

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/easterthebunny/spew-order/pkg/types"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrNotImplemented = errors.New("function not implemented")
)

func NewAirdropSource(key string) Source {
	return &airdropSource{
		secretKey: key,
	}
}

type airdropSource struct {
	secretKey string
}

func (s *airdropSource) Name() string {
	return "CMTN"
}

func (s *airdropSource) Supports(symbol types.Symbol) bool {
	switch symbol {
	case types.SymbolCipherMtn:
		return true
	default:
		return false
	}
}

func (s *airdropSource) Callback() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			var err error
			ok := true

			for ok {
				if r.Method != "POST" {
					err = errors.New("incorrect callback method")
					ctx = attachToContext(ctx, nil, &CallbackError{Status: http.StatusNotAcceptable, Err: err})
					break
				}

				signature := r.Header.Get("CMTN-SIGNATURE")
				if signature != s.secretKey {
					err = errors.New("not authorized")
					ctx = attachToContext(ctx, nil, &CallbackError{Status: http.StatusNotAcceptable, Err: err})
					break
				}

				var posting fundingStruct
				err := json.NewDecoder(r.Body).Decode(&posting)
				if err != nil {
					err = errors.New("request parse error")
					ctx = attachToContext(ctx, nil, &CallbackError{Status: http.StatusNotAcceptable, Err: err})
					break
				}

				amt, err := strconv.Atoi(posting.FundingAmount)
				if err != nil {
					err = errors.New("invalid amount")
					ctx = attachToContext(ctx, nil, &CallbackError{Status: http.StatusNotAcceptable, Err: err})
					break
				}

				tr := Transaction{
					Symbol:          types.SymbolCipherMtn,
					TransactionHash: posting.TransactionHash,
					Address:         posting.Address,
					Amount:          decimal.NewFromInt(int64(amt)),
				}

				ctx = attachToContext(ctx, tr, nil)

				ok = false
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (s *airdropSource) CreateAddress(symbol types.Symbol) (*Address, error) {

	id := uuid.NewV4()

	hash := hmac.New(sha256.New, uuid.NewV4().Bytes())
	_, err := io.WriteString(hash, id.String())
	if err != nil {
		return nil, err
	}
	encoded := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	return &Address{ID: id.String(), Hash: encoded}, nil
}

func (s *airdropSource) Withdraw(tr *Transaction) (string, error) {
	return "", ErrNotImplemented
}

func (s *airdropSource) OKResponse() int {
	return http.StatusOK
}

type fundingStruct struct {
	Address         string `json:"address"`
	FundingAmount   string `json:"amount"`
	TransactionHash string `json:"transaction_hash"`
}
