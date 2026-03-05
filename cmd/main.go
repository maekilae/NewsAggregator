package main

import (
	"log/slog"
	"newsaggregator/internal/auth"
	"newsaggregator/internal/classifier"
	"newsaggregator/internal/db"
	"newsaggregator/internal/logger"
	"newsaggregator/internal/rss_handler"
)

func main() {
	logger.Init("log")
	d, err := db.InitDB()
	if err != nil {
		slog.Error("Failed to initialize database", slog.String("error", err.Error()))
		return
	}
	defer d.Close()
	c := classifier.New("MISTRAL_API_KEY", "https://api.mistral.ai/v1/chat/completions")

	providers := []string{"https://www.svt.se/rss.xml"}

	for _, p := range providers {
		_ = rss_handler.NewRSSHandler(p, d, c)
		// feed.UpdateFeed()
	}
	auth.NewAccessToken(d)

	d.Iterate()
}
