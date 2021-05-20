package kv

import (
	"fmt"

	"github.com/easterthebunny/spew-order/internal/persist"
)

type AuthorizationRepository struct {
	kvstore persist.KVStore
}

func NewAuthorizationRepository(store persist.KVStore) *AuthorizationRepository {
	return &AuthorizationRepository{kvstore: store}
}

func (a *AuthorizationRepository) GetAuthorization(id persist.Key) (authz *persist.Authorization, err error) {

	k := authzKey(id)

	b, err := a.kvstore.Get(k)
	if err != nil {
		return nil, err
	}

	attr, err := a.kvstore.Attrs(k)
	if err != nil {
		return
	}

	authz = &persist.Authorization{}
	err = authz.Decode(b, encodingFromStr(attr.ContentEncoding))
	if err != nil {
		return
	}

	return
}

func (a *AuthorizationRepository) SetAuthorization(authz *persist.Authorization) error {

	if authz == nil {
		return fmt.Errorf("%w for authorization", persist.ErrCannotSaveNilValue)
	}

	enc := persist.JSON
	b, err := authz.Encode(enc)
	if err != nil {
		return err
	}

	attrs := persist.KVStoreObjectAttrsToUpdate{
		ContentEncoding: encodingToStr(enc),
		Metadata:        make(map[string]string),
	}

	a.kvstore.Set(authzKey(stringer(authz.ID)), b, &attrs)

	return nil
}

func (a *AuthorizationRepository) DeleteAuthorization(authz *persist.Authorization) error {

	if authz == nil {
		return fmt.Errorf("%w for authorization", persist.ErrCannotSaveNilValue)
	}

	return a.kvstore.Delete(authzKey(stringer(authz.ID)))
}
