package firebase

import (
	"context"
	"errors"

	"cloud.google.com/go/firestore"
	"github.com/easterthebunny/spew-order/internal/persist"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthorizationRepository struct {
	client *firestore.Client
}

func NewAuthorizationRepository(client *firestore.Client) *AuthorizationRepository {
	return &AuthorizationRepository{client: client}
}

// /root/authz/{authzid}
func (a *AuthorizationRepository) GetAuthorization(ctx context.Context, id persist.Key) (authz *persist.Authorization, err error) {
	dsnap, err := a.getClient(ctx).Collection("authz").Doc(id.String()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, persist.ErrObjectNotExist
		}

		return nil, err
	}

	var au persist.Authorization
	err = dsnap.DataTo(&au)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, persist.ErrObjectNotExist
		}

		return nil, err
	}

	authz = &au

	return
}

func (a *AuthorizationRepository) GetAuthorizations(ctx context.Context) (authz []*persist.Authorization, err error) {
	iter := a.getClient(ctx).Collection("authz").Documents(ctx)
	var doc *firestore.DocumentSnapshot
	for {
		doc, err = iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				err = nil
			}

			break
		}

		var item persist.Authorization
		if err := doc.DataTo(&item); err == nil {
			authz = append(authz, &item)
		}
	}
	return
}

func (a *AuthorizationRepository) SetAuthorization(ctx context.Context, authz *persist.Authorization) error {
	_, err := a.getClient(ctx).Collection("authz").Doc(authz.ID).Set(ctx, *authz)
	if err != nil {
		return err
	}
	return nil
}

func (a *AuthorizationRepository) DeleteAuthorization(ctx context.Context, authz *persist.Authorization) error {
	_, err := a.getClient(ctx).Collection("authz").Doc(authz.ID).Delete(ctx)
	return err
}

func (a *AuthorizationRepository) getClient(ctx context.Context) *firestore.Client {

	var client *firestore.Client
	if a.client == nil {
		client = clientFromContext(ctx)
	} else {
		client = a.client
	}
	return client
}
