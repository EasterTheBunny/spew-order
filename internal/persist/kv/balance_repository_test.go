package kv

import (
	"testing"
	"time"

	"github.com/easterthebunny/spew-order/internal/persist"
	"github.com/easterthebunny/spew-order/pkg/types"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestBalances(t *testing.T) {

	s := persist.NewMockKVStore()
	m := types.SymbolBitcoin

	id := uuid.NewV4()

	a := persist.Account{
		ID: id.String(),
	}

	br := &BalanceRepository{kvstore: s, account: &a, symbol: m}
	assert.NotNil(t, br)

	t.Run("StartingZeroBalance", func(t *testing.T) {

		v, err := br.GetBalance()
		if err != nil {
			assert.FailNowf(t, "Error encountered getting balance: %s", err.Error())
		}

		expected := decimal.NewFromInt(0)
		if !expected.Equal(v) {
			assert.FailNowf(t, "Balance expected to be zero: %s", v.StringFixedBank(2))
		}
	})

	t.Run("UpdateBalance", func(t *testing.T) {

		bal := decimal.NewFromFloat(10.288942)
		err := br.UpdateBalance(bal)
		assert.NoError(t, err)

		v, err := br.GetBalance()
		if err != nil {
			assert.FailNowf(t, "Error encountered getting balance: %s", err.Error())
		}

		assert.Equal(t, bal.StringFixedBank(6), v.StringFixedBank(6), "Saved balance must match expected")
	})

	t.Run("StartingEmptyHolds", func(t *testing.T) {

		holds, err := br.FindHolds()
		if err != nil {
			assert.FailNowf(t, "Error encountered getting holds: %s", err.Error())
		}

		assert.Len(t, holds, 0)
	})

	t.Run("InsertAndDeleteHolds", func(t *testing.T) {

		var expected []*persist.BalanceItem

		for x := 1; x < 4; x++ {
			amt := decimal.NewFromInt(int64(x))
			hold := persist.BalanceItem{
				ID:        uuid.NewV4().String(),
				Timestamp: persist.NanoTime(time.Now()),
				Amount:    amt,
			}
			expected = append(expected, &hold)

			err := br.CreateHold(&hold)
			if err != nil {
				assert.FailNowf(t, "Error encountered saving hold: %s", err.Error())
			}
		}

		holds, err := br.FindHolds()
		if err != nil {
			assert.FailNowf(t, "Error encountered getting holds: %s", err.Error())
		}

		assert.Len(t, holds, len(expected))

		err = br.DeleteHold(expected[0])
		if err != nil {
			assert.FailNowf(t, "Error encountered deleting hold: %s", err.Error())
		}

		expected = expected[1:]

		holds, err = br.FindHolds()
		if err != nil {
			assert.FailNowf(t, "Error encountered getting holds: %s", err.Error())
		}

		assert.Len(t, holds, len(expected))
	})

	t.Run("InsertAndDeletePosts", func(t *testing.T) {

		var expected []*persist.BalanceItem

		for x := 1; x < 4; x++ {
			amt := decimal.NewFromInt(int64(x))
			post := persist.BalanceItem{
				ID:        uuid.NewV4().String(),
				Timestamp: persist.NanoTime(time.Now()),
				Amount:    amt,
			}
			expected = append(expected, &post)

			err := br.CreatePost(&post)
			if err != nil {
				assert.FailNowf(t, "Error encountered saving post: %s", err.Error())
			}
		}

		posts, err := br.FindPosts()
		if err != nil {
			assert.FailNowf(t, "Error encountered getting posts: %s", err.Error())
		}

		assert.Len(t, posts, len(expected))

		err = br.DeletePost(expected[0])
		if err != nil {
			assert.FailNowf(t, "Error encountered deleting post: %s", err.Error())
		}

		expected = expected[1:]

		posts, err = br.FindPosts()
		if err != nil {
			assert.FailNowf(t, "Error encountered getting posts: %s", err.Error())
		}

		assert.Len(t, posts, len(expected))
	})

}
