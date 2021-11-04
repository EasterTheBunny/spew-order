package funding

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestCallback(t *testing.T) {
	src := NewAirdropSource("secret")
	data := `{"amount":"50000","transaction_hash":"test","address":"hash"}`

	r := req(t, post(data))
	w := httptest.NewRecorder()

	handler := src.Callback()

	t.Run("InvalidSecretHeader", func(t *testing.T) {
		serve := &testHandler{
			TestCases: func(tr *Transaction, err *CallbackError) {
				assert.NotNil(t, err, "error not nil")
				if err != nil {
					assert.Equal(t, http.StatusNotAcceptable, err.Status)
				}

				assert.Nil(t, tr, "transaction nil")
			},
		}

		f := handler(serve)
		f.ServeHTTP(w, r.WithContext(context.Background()))
	})

	t.Run("Success", func(t *testing.T) {
		r.Header.Add("CMTN-SIGNATURE", "secret")
		serve := &testHandler{
			TestCases: func(tr *Transaction, err *CallbackError) {
				assert.NotNil(t, tr, "transaction not nil")
				if tr != nil {
					assert.Equal(t, types.SymbolCipherMtn, tr.Symbol)
					assert.Equal(t, "test", tr.TransactionHash)
					assert.Equal(t, "hash", tr.Address)
					assert.Equal(t, "50000", tr.Amount.StringFixed(0))
				}

				assert.Nil(t, err, "transaction nil")
			},
		}

		f := handler(serve)
		f.ServeHTTP(w, r.WithContext(context.Background()))
	})
}

type testHandler struct {
	TestCases func(*Transaction, *CallbackError)
}

func (h *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tr, err := TransactionFromContext(r.Context())

	h.TestCases(tr, err)
}

func post(cont string) string {
	post :=
		`POST / HTTP/1.1
Content-Type: application/json
User-Agent: mockagent
Content-Length: %d

%s`
	return fmt.Sprintf(post, len(cont), cont)
}

func req(t testing.TB, v string) *http.Request {
	req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(v)))
	if err != nil {
		t.Fatal(err)
	}
	return req
}
