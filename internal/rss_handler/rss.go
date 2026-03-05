package rss_handler

import (
	"crypto/sha256"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"newsaggregator/internal/article"
	"newsaggregator/internal/classifier"
	"newsaggregator/internal/db"
	"newsaggregator/internal/webscraper"

	xpp "github.com/mmcdole/goxpp"
)

type RSSHandler struct {
	Name       string
	Url        string
	Articles   map[string]article.Article
	ws         *webscraper.WebScraper
	db         *db.DB
	Classifier *classifier.Classifier
}

func NewRSSHandler(url string, db *db.DB, c *classifier.Classifier) *RSSHandler {

	ws := webscraper.InitWebScraper([]string{})
	rh := &RSSHandler{
		Url:        url,
		Articles:   make(map[string]article.Article),
		ws:         ws,
		db:         db,
		Classifier: c,
	}
	err := rh.UpdateFeed()
	if err != nil {
		slog.Error("Failed to update feed", "url", rh.Url, "error", err)
	}
	return rh
}

func (rh *RSSHandler) UpdateFeed() error {
	slog.Debug("Updating feed", "url", rh.Url)
	resp, err := http.Get(rh.Url)
	slog.Debug("Received response", "status", resp.Status)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	slog.Debug("Parsing feed with URL", "url", rh.Url)
	rh.parseFeed(resp.Body)

	classArticles, err := rh.Classifier.Classify(rh.Articles)
	if err != nil {
		return err
	}
	rh.Articles = classArticles

	for _, article := range rh.Articles {
		hash := sha256.Sum256([]byte(article.Url))
		article.UrlHash = hash[:]
		tn, err := rh.db.Get([]byte("url:" + string(article.UrlHash)))
		if err == nil {
			slog.Debug("Article already exists")
			article.Thumbnail = string(tn)
			continue
		}
		rh.ws.ScrapeMeta(&article)
		// Classify

		slog.Debug("Inserting article", "url", article.Url)
		err = rh.db.Insert([]byte("url:"+string(article.UrlHash)), []byte(article.Thumbnail))
		if err != nil {
			slog.Error("Failed to insert article", "error", err.Error())
		}
		err = rh.db.Insert([]byte("category:"+string(article.Tag)), []byte(article.Url))
		if err != nil {
			slog.Error("Failed to insert article category", "error", err.Error())
		}
	}
	return nil
}

func (rh *RSSHandler) parseFeed(body io.Reader) error {
	p := xpp.NewXMLPullParser(body, false, nil)
	for tok, err := p.NextTag(); tok != xpp.EndDocument; tok, err = p.NextTag() {
		if err != nil {
			return err
		}
		if tok == xpp.StartTag && p.Name == "channel" {
			for tok, err = p.NextTag(); tok != xpp.EndTag; tok, err = p.NextTag() {
				if err != nil {
					return err
				}
				if tok == xpp.StartTag {
					switch p.Name {
					case "title":
						rh.Name, _ = p.NextText()
					case "item":
						p.NextTag()
						title, _ := p.NextText()
						p.NextTag()
						linkRaw, _ := p.NextText()
						link, err := url.Parse(linkRaw)
						if err != nil {
							return err
						}
						p.NextTag()
						desc, _ := p.NextText()
						if _, ok := rh.Articles[linkRaw]; !ok {
							rh.Articles[linkRaw] = article.Article{
								Url:         linkRaw,
								Path:        link.Path,
								Provider:    link.Host,
								Title:       title,
								Description: desc,
							}
						}
						p.Skip()
					default:
						p.Skip()
					}
				}
			}
			break
		}
	}

	return nil

}
