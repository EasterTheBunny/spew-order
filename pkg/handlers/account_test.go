package handlers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http/httptest"
	"testing"

	iaccount "github.com/easterthebunny/spew-order/internal/account"
	"github.com/easterthebunny/spew-order/internal/contexts"
	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/api"
	"github.com/easterthebunny/spew-order/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestGetAccount(t *testing.T) {

	// set up a buffer to log to
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)

	acct := types.NewAccount()
	repo := iaccount.NewKVAccountRepository(persist.NewMockKVStore())

	r := NewGet(t, "/")
	r = r.WithContext(contexts.AttachAccount(r.Context(), acct))

	// create a response recorder for later inspection of the response
	w := httptest.NewRecorder()

	h := NewAccountHandler(repo).GetAccount()

	h(w, r)

	assert.Equal(t, 200, w.Code, "response code is a 200 success")

	var res map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&res)
	assert.NoError(t, err)

	data, ok := res["data"]
	assert.Equal(t, true, ok)

	bts, _ := json.Marshal(data)

	var responseAccount api.Account
	err = json.Unmarshal(bts, &responseAccount)
	assert.NoError(t, err)

	assert.Equal(t, acct.ID.String(), responseAccount.Id)
}
