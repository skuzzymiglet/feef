package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
)

type Fetcher struct {
	maxThreads int
	client     http.Client
	sema       chan struct{}
}

func (f *Fetcher) Fetch(ctx context.Context, url string, out chan<- LinkedFeedItem) error {
	parser := gofeed.NewParser() // lol race
	var body bytes.Reader
	req, err := http.NewRequestWithContext(ctx, "GET", url, &body)
	if err != nil {
		return fmt.Errorf("Error creating request for %s : %w", url, err)
	}
	f.sema <- struct{}{}
	resp, err := f.client.Do(req)
	if err != nil {
		return fmt.Errorf("Error fetching %s : %w", url, err)
	}
	feed, err := parser.Parse(resp.Body)
	resp.Body.Close()
	<-f.sema
	if err != nil {
		return fmt.Errorf("Error parsing %s : %w", url, err)
	}
	linkedFeed := LinkFeed(feed)
	if linkedFeed.Feed.FeedLink != url {
		log.Debugf("feed request url and self-reference url mismatch: requested %s, got %s", url, linkedFeed.Feed.FeedLink)
	}
	for _, i := range linkedFeed.Items {
		out <- i
	}

	return nil
}
