package kv

import (
	"testing"

	"github.com/easterthebunny/spew-order/internal/persist"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuthorizationPersistence(t *testing.T) {

	s := persist.NewMockKVStore()
	r := NewAuthorizationRepository(s)

	id := uuid.NewV4()

	expected := persist.Authorization{
		ID:       id.String(),
		Username: "test3",
		Email:    "test4",
		Name:     "test5",
		Avatar:   "test6",
		Accounts: []string{"test7", "test8"},
	}

	t.Run("Set/Get", func(t *testing.T) {
		err := r.SetAuthorization(&expected)
		assert.NoError(t, err)

		x, err := r.GetAuthorization(id)
		assert.NoError(t, err)

		if x != nil {
			assert.Equal(t, expected, *x)
		} else {
			t.Error("nil result found")
		}
	})

	t.Run("Update", func(t *testing.T) {
		expected.Username = "test9"

		err := r.SetAuthorization(&expected)
		assert.NoError(t, err)

		x, err := r.GetAuthorization(id)
		assert.NoError(t, err)

		if x != nil {
			assert.Equal(t, expected, *x)
		} else {
			t.Error("nil result found")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		err := r.DeleteAuthorization(&expected)
		assert.NoError(t, err)

		x, err := r.GetAuthorization(id)
		assert.Nil(t, x)
		assert.ErrorIs(t, err, persist.ErrObjectNotExist)
	})
}
