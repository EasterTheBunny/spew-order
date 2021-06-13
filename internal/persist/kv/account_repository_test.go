package kv

import (
	"context"
	"testing"

	"github.com/easterthebunny/spew-order/internal/persist"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestSaveFind(t *testing.T) {

	ctx := context.Background()

	s := persist.NewMockKVStore()
	r := &AccountRepository{kvstore: s}

	id := uuid.NewV4()

	expected := persist.Account{
		ID: id.String(),
	}

	var a *persist.Account
	var err error

	t.Run("NotFound", func(t *testing.T) {

		a, err = r.Find(ctx, id)
		if err == nil {
			assert.FailNowf(t, "Error expected when attempting to find account by id: %s", id.String())
		}

		if a != nil {
			assert.FailNowf(t, "Account found for id: %s when none was expected", id.String())
		}
	})

	t.Run("Save", func(t *testing.T) {

		err = r.Save(context.Background(), &expected)
		assert.NoError(t, err)
	})

	t.Run("Find", func(t *testing.T) {

		a, err = r.Find(ctx, id)
		assert.NoError(t, err)
		if err != nil {
			assert.FailNowf(t, "Error encountered: %s", err.Error())
		}

		if a == nil {
			assert.FailNowf(t, "Account for id: %s was nil", id.String())
		} else {
			assert.Equal(t, expected, *a)
		}
	})
}
