package types

import uuid "github.com/satori/go.uuid"

type ID uuid.UUID

func NewID() ID {
	return ID(uuid.NewV4())
}

func (id ID) String() string {
	return uuid.UUID(id).String()
}

func (id ID) MarshalText() ([]byte, error) {
	return uuid.UUID(id).MarshalText()
}

func (id *ID) UnmarshalText(b []byte) error {

	q := &uuid.UUID{}
	if err := q.UnmarshalText(b); err != nil {
		return err
	}

	g := ID(*q)
	id = &g

	return nil
}
