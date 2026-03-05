package feed

import (
	"crypto/sha256"
	"log/slog"
	"net/http"
	"newsaggregator/internal/classifier"
	"newsaggregator/internal/db"
	"newsaggregator/internal/webscraper"
)

type feedItems struct {
	Keys   []string
	Values map[string]string
}

func (f *feedItems) Add(key, value string) {
	f.Keys = append(f.Keys, key)
	f.Values[key] = value
}

func (f *feedItems) Get(key string) (string, bool) {
	value, ok := f.Values[key]
	return value, ok
}

func (f *feedItems) Delete(key string) {
	for i, k := range f.Keys {
		if k == key {
			f.Keys = append(f.Keys[:i], f.Keys[i+1:]...)
			delete(f.Values, key)
			return
		}
	}
}

func (f *feedItems) Get_Idx(key string) (int, bool) {
	for i, k := range f.Keys {
		if k == key {
			return i, true
		}
	}
	return -1, false
}

type FeedHandler struct {
	db         *db.DB
	Url        string
	Items      feedItems
	Classifier *classifier.Classifier
	ws         *webscraper.WebScraper
}

func NewFeedHandler(url string) *FeedHandler {
	return &FeedHandler{Url: url}
}

func (h *FeedHandler) GetFeed() {

	slog.Debug("Updating feed", "url", h.Url)
	resp, err := http.Get(h.Url)
	slog.Debug("Received response", "status", resp.Status)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	slog.Debug("Parsing feed with URL", "url", h.Url)
	h.parseFeed(resp.Body)

	classArticles, err := h.Classifier.Classify(h.Items)
	if err != nil {
		return err
	}
	h.Items = classArticles

	for _, article := range h.Items {
		hash := sha256.Sum256([]byte(article.Url))
		article.UrlHash = hash[:]
		tn, err := h.db.Get([]byte("url:" + string(article.UrlHash)))
		if err == nil {
			slog.Debug("Article already exists")
			article.Thumbnail = string(tn)
			continue
		}
		h.ws.ScrapeMeta(&article)
		// Classify

		slog.Debug("Inserting article", "url", article.Url)
		err = h.db.Insert([]byte("url:"+string(article.UrlHash)), []byte(article.Thumbnail))
		if err != nil {
			slog.Error("Failed to insert article", "error", err.Error())
		}
		err = h.db.Insert([]byte("category:"+string(article.Tag)), []byte(article.Url))
		if err != nil {
			slog.Error("Failed to insert article category", "error", err.Error())
		}
	}
	return nil
}
