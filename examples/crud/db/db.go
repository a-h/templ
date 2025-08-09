package db

import (
	"context"

	"github.com/a-h/kv"
)

func New(store kv.Store) *DB {
	return &DB{
		Store: store,
	}
}

type DB struct {
	Store kv.Store
}

func (db *DB) List(ctx context.Context) (contacts []Contact, err error) {
	records, err := db.Store.List(ctx, -1, -1)
	if err != nil {
		return nil, err
	}
	return kv.ValuesOf[Contact](records)
}

func (db *DB) Save(ctx context.Context, c Contact) error {
	return db.Store.Put(ctx, c.ID, -1, c)
}

func (db *DB) Get(ctx context.Context, id string) (c Contact, ok bool, err error) {
	_, ok, err = db.Store.Get(ctx, id, &c)
	if err != nil {
		return Contact{}, false, err
	}
	return c, ok, nil
}

func (db *DB) Delete(ctx context.Context, id string) error {
	if _, err := db.Store.Delete(ctx, id); err != nil {
		return err
	}
	return nil
}

func NewContact(id, name, email string) Contact {
	return Contact{
		ID:    id,
		Name:  name,
		Email: email,
	}
}

type Contact struct {
	ID    string
	Name  string
	Email string
}
