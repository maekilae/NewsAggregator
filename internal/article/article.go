package article

import (
	"encoding/json"
	"encoding/xml"
)

type Article struct {
	XMLName     xml.Name `json:"-" xml:"item"`
	Url         string   `json:"url"`
	UrlHash     []byte   `json:"-" xml:"-"`
	Path        string   `json:"-" xml:"-"`
	Provider    string   `json:"provider" xml:"-"`
	Title       string   `json:"title" xml:"title"`
	Description string   `json:"description" xml:"description"`
	Thumbnail   string   `json:"thumbnail"`
	Tag         string   `json:"tag"`
}

func (a *Article) MarshalJSON() ([]byte, error) {
	return json.Marshal(a)

}
func (a *Article) MarshalRSS() ([]byte, error) {
	return xml.Marshal(a)
}
