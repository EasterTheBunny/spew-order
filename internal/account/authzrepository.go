package account

import (
	"encoding/json"
	"errors"

	"github.com/easterthebunny/spew-order/internal/auth"
	"github.com/easterthebunny/spew-order/internal/key"
	"github.com/easterthebunny/spew-order/internal/persist"
)

type KVAuthzRepository struct {
	kvstore persist.KVStore
}

func NewKVAuthzRepository(store persist.KVStore) *KVAuthzRepository {
	return &KVAuthzRepository{kvstore: store}
}

func (a *KVAuthzRepository) GetAuthorization(id string) (*auth.Authorization, error) {

	k := gsAuthz.Pack(key.Tuple{id})
	b, err := a.kvstore.Get(k.String())
	if err != nil {
		return nil, err
	}

	var c auth.Authorization
	err = json.Unmarshal(b, &a)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (a *KVAuthzRepository) SetAuthorization(au *auth.Authorization) error {

	if au == nil {
		return errors.New("no authorization available for nil value")
	}

	k := gsAuthz.Pack(key.Tuple{au.ID})

	b, err := json.Marshal(*a)
	if err != nil {
		return err
	}

	attrs := persist.KVStoreObjectAttrsToUpdate{
		Metadata: make(map[string]string),
	}

	a.kvstore.Set(k.String(), b, &attrs)

	return nil
}

func (a *KVAuthzRepository) DeleteAuthorization(au *auth.Authorization) error {

	if au == nil {
		return errors.New("no authorization available for nil value")
	}

	k := gsAuthz.Pack(key.Tuple{au.ID})

	return a.kvstore.Delete(k.String())
}
