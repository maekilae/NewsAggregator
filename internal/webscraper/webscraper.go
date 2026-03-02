package webscraper

import (
	"log/slog"
	"newsaggregator/internal/article"

	"github.com/gocolly/colly"
)

type WebScraper struct {
	c *colly.Collector
}

func InitWebScraper(urls []string) *WebScraper {
	c := colly.NewCollector()
	return &WebScraper{
		c,
	}
}

func (ws *WebScraper) ScrapeMeta(article *article.Article) (*article.Article, error) {
	selector := "meta[property='og:image']"
	attr := "content"
	ws.c.OnHTML(selector, func(e *colly.HTMLElement) {
		article.Thumbnail = e.Attr(attr)
	})

	slog.Debug("Visiting article", slog.String("url", article.Url))
	ws.c.Visit(article.Url)

	return article, nil
}
