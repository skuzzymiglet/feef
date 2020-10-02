package main

import (
	"fmt"
	"net/http"

	"github.com/mmcdole/gofeed"
	"github.com/sirupsen/logrus"
)

// Feeds holds RSS/Atom feeds
type Feeds struct {
	// TODO: Use a better feed abstraction
	// less bloat
	// items have reference to feed
	Feeds  map[string]gofeed.Feed
	Logger *logrus.Logger
}

// Fetch fetches the feeds from the URL specified
func Fetch(url string, httpClient *http.Client) (*gofeed.Feed, error) {
	var feed *gofeed.Feed
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http.Get error on %s: %w", url, err)
	}
	parser := gofeed.NewParser()

	feed, err = parser.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse error on %s: %w", url, err)
	}
	return feed, nil
}
