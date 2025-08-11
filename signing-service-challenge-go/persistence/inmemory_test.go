package persistence_test

import (
	"errors"
	"testing"

	"github.com/fiskaly/coding-challenges/signing-service-challenge/persistence"
	"golang.org/x/exp/slices"
)

func TestStoreItem(t *testing.T) {
	store := persistence.NewInMemoryStore[int]()

	if err := store.Store("some id", 1); err != nil {
		t.Errorf("Storing item in store returned unexpected error: %s", err.Error())
	}
}

func TestStoreDuplicitItem(t *testing.T) {
	store := persistence.NewInMemoryStore[int]()

	if err := store.Store("some id", 1); err != nil {
		t.Errorf("Storing item in store returned unexpected error: %s", err.Error())
	}

	if err := store.Store("some id", 1); !errors.Is(err, persistence.ErrItemExists) {
		t.Error("Storing item in store twice did not return expected error")
	}
}

func TestGetNonExistingItem(t *testing.T) {
	store := persistence.NewInMemoryStore[int]()

	_, err := store.Get("invalid id")

	if !errors.Is(err, persistence.ErrItemNotFound) {
		t.Error("Getting non existent item from store did not return expected error")
	}
}

func TestGetItem(t *testing.T) {
	store := persistence.NewInMemoryStore[int]()

	_, err := store.Get("invalid id")

	if !errors.Is(err, persistence.ErrItemNotFound) {
		t.Error("Getting non existent item from store did not return expected error")
	}
}

func TestListItems(t *testing.T) {
	store := persistence.NewInMemoryStore[int]()

	items, err := store.List()
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	if len(items) != 0 {
		t.Errorf("Unexpected size of the item list: %d, expected 0", len(items))
	}

	store.Store("some id", 1)
	store.Store("some other id", 2)

	items, err = store.List()
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	if len(items) != 2 {
		t.Errorf("Unexpected size of the item list: %d, expected 2", len(items))
	}

	if !slices.Contains(items, 1) {
		t.Error("Item '1' not found in the item list")
	}

	if !slices.Contains(items, 2) {
		t.Error("Item '2' not found in the item list")
	}
}
