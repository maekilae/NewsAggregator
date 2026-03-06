package feed

import (
	"crypto/sha256"
	"log/slog"
	"net/http"
	"newsaggregator/internal/classifier"
	"newsaggregator/internal/db"
	"newsaggregator/internal/item"
	"newsaggregator/internal/webscraper"
)

type FeedHandler struct {
	db         *db.DB
	Url        string
	Items      *item.ItemHandler
	Classifier *classifier.Classifier
	ws         *webscraper.WebScraper
	outFormat  []string
}

func NewFeedHandler(url string, outFormat []string, db *db.DB, keyEnv string, classifierUrl string) *FeedHandler {
	fh := FeedHandler{
		db:         db,
		Url:        url,
		Items:      item.NewItemHandler(db),
		Classifier: classifier.New(keyEnv, classifierUrl),
		ws:         webscraper.New([]string{url}),
		outFormat:  outFormat,
	}
	return &fh
}

func (fh *FeedHandler) UpdateFeed() error {

	slog.Debug("Updating feed", "url", fh.Url)
	resp, err := http.Get(fh.Url)
	slog.Debug("Received response", "status", resp.Status)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	slog.Debug("Parsing feed with URL", "url", fh.Url)
	fh.parseFeed(resp.Body)

	classArticles, err := fh.Classifier.Classify(fh.Items)
	if err != nil {
		return err
	}
	fh.Items = classArticles

	for _, article := range fh.Items {
		hash := sha256.Sum256([]byte(article.Url))
		article.UrlHash = hash[:]
		tn, err := fh.db.Get([]byte("url:" + string(article.UrlHash)))
		if err == nil {
			slog.Debug("Article already exists")
			article.Thumbnail = string(tn)
			continue
		}
		fh.ws.ScrapeMeta(&article)
		// Classify

		slog.Debug("Inserting article", "url", article.Url)
		err = fh.db.Insert([]byte("url:"+string(article.UrlHash)), []byte(article.Thumbnail))
		if err != nil {
			slog.Error("Failed to insert article", "error", err.Error())
		}
		err = fh.db.Insert([]byte("category:"+string(article.Tag)), []byte(article.Url))
		if err != nil {
			slog.Error("Failed to insert article category", "error", err.Error())
		}
	}
	return nil
}
