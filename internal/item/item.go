package item

import (
	"bytes"
	"encoding/gob"
	"encoding/xml"
	"newsaggregator/internal/db"
)

type Item struct {
	Hash        []byte
	Provider    string
	Title       string
	Url         string
	Description string
	PubDate     string
	Image       string
	Tags        []string
}

type GOBItem struct {
	Provider    string
	Title       string
	Description string
	PubDate     string
	Image       string
	Tags        []string
}

type JSONItem struct {
	Provider    string   `json:"provider"`
	Title       string   `json:"title"`
	Url         string   `json:"url"`
	Description string   `json:"description"`
	PubDate     string   `json:"pub_date"`
	Image       string   `json:"image"`
	Tags        []string `json:"tags"`
}

type RSSItem struct {
	XMLName     xml.Name `xml:"item"`
	Provider    string   `json:"provider"`
	Title       string   `json:"title"`
	Url         string   `json:"url"`
	Description string   `json:"description"`
	PubDate     string   `json:"pub_date"`
	Image       string   `json:"image"`
	Tags        []string `json:"tags"`
}

type ATOMItem struct {
	XMLName     xml.Name `xml:"entry"`
	Provider    string   `json:"provider"`
	Title       string   `json:"title"`
	Url         string   `json:"url"`
	Description string   `json:"description"`
	PubDate     string   `json:"pub_date"`
	Image       string   `json:"image"`
	Tags        []string `json:"tags"`
}

func (item *Item) MarshalGOB() ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	gobItem := &GOBItem{
		Provider:    item.Provider,
		Title:       item.Title,
		Description: item.Description,
		PubDate:     item.PubDate,
		Image:       item.Image,
		Tags:        item.Tags,
	}
	if err := enc.Encode(gobItem); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (item *Item) UnmarshalGOB(data []byte) error {
	dec := gob.NewDecoder(bytes.NewReader(data))
	gobItem := &GOBItem{}
	if err := dec.Decode(gobItem); err != nil {
		return err
	}
	item.Provider = gobItem.Provider
	item.Title = gobItem.Title
	item.Description = gobItem.Description
	item.PubDate = gobItem.PubDate
	item.Image = gobItem.Image
	item.Tags = gobItem.Tags
	return nil
}

func (item *Item) Insert(db *db.DB) error {
	hash := item.Hash
	if hash == nil {
		hash = item.Hash
	}
	gobItem, err := item.MarshalGOB()
	if err != nil {
		return err
	}
	if err := db.Insert(append([]byte("url:"), hash...), gobItem); err != nil {
		return err
	}
	// Tags[0] is the primary category for this item
	if err := db.Insert(append([]byte("category:"), []byte(item.Tags[0])...), hash); err != nil {
		return err
	}
	return nil
}
