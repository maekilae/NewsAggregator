package item

import (
	"errors"
	"newsaggregator/internal/db"
)

type ItemHandler struct {
	db    *db.DB
	keys  []string
	items map[string]Item
}

func NewItemHandler(db *db.DB) *ItemHandler {
	return &ItemHandler{
		db:    db,
		items: make(map[string]Item),
	}
}

func (ih *ItemHandler) NewItemEntry() (*Item, error) {
	item := Item{}
	return &item, nil
}

// Insert all items currently in the Itemhandler.items hashmap into the database
// Skips items already present in the database
func (ih *ItemHandler) InsertAll() error {
	for _, key := range ih.keys {
		item := ih.items[key]
		if err := item.Insert(ih.db); err != nil {
			return err
		}
	}
	return nil
}

func (ih *ItemHandler) Add(item Item) error {
	if item.Hash == nil {
		return errors.New("item hash is nil")
	}
	if _, ok := ih.items[string(item.Hash)]; ok {
		return nil
	}

	ih.keys = append(ih.keys, string(item.Hash))
	ih.items[string(item.Hash)] = item
	return nil
}

func (ih *ItemHandler) InsertItem(item Item) error {
	if err := item.Insert(ih.db); err != nil {
		return err
	}
	return nil
}

func (ih *ItemHandler) GetItem(hash []byte) (*Item, error) {
	it, ok := ih.items[string(hash)]
	if ok {
		return &it, nil
	}
	v, err := ih.db.Get(hash)
	if err != nil {
		return nil, err
	}
	item := &Item{}
	if err := item.UnmarshalGOB(v); err != nil {
		return nil, err
	}

	return item, nil
}

func (ih *ItemHandler) GetItemsByCategory(category string) ([]*Item, error) {
	values, err := ih.db.GetByPrefix([]byte("category:" + category))
	if err != nil {
		return nil, err
	}
	items := make([]*Item, 0, len(values))
	for _, v := range values {
		item, ok := ih.items[string(v)]
		if ok {
			items = append(items, &item)
			continue
		}
		iv, err := ih.db.Get(append([]byte("url:"), v...))
		if err != nil {
			return nil, err
		}
		item = Item{}
		if err := item.UnmarshalGOB(iv); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	return items, nil
}
