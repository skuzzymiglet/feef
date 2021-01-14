package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mmcdole/gofeed"
	"github.com/peterbourgon/diskv"
	log "github.com/sirupsen/logrus"

	"github.com/gosimple/slug"
)

type Fetcher struct {
	client    *http.Client
	sema      chan struct{}
	itemCache *diskv.Diskv
}

var encodeFunc func(string) string = slug.Make

func (f *Fetcher) Fetch(ctx context.Context, url string) (LinkedFeed, error) {
	// HEAD, see if from cache
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return LinkedFeed{}, fmt.Errorf("Error creating request for %s : %w", url, err)
	}
	f.sema <- struct{}{}
	resp, err := f.client.Do(req)
	if err != nil {
		<-f.sema
		return LinkedFeed{}, fmt.Errorf("Error fetching %s : %w", url, err)
	}
	<-f.sema
	if len(resp.Header["X-From-Cache"]) == 1 && resp.Header["X-From-Cache"][0] == "1" {
		// lookup in itemCache
		keyName := encodeFunc(url)
		if f.itemCache.Has(keyName) {
			b, err := f.itemCache.Read(keyName)
			if err != nil {
				return LinkedFeed{}, err
			}
			var buf *gofeed.Feed
			err = json.Unmarshal(b, &buf)
			if err != nil {
				return LinkedFeed{}, err
			}
			return LinkFeed(buf), err
		}
	}
	log.Infof("get on %s", url)
	req, err = http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return LinkedFeed{}, fmt.Errorf("Error creating request for %s : %w", url, err)
	}
	f.sema <- struct{}{}
	resp, err = f.client.Do(req)
	if err != nil {
		return LinkedFeed{}, fmt.Errorf("Error fetching %s : %w", url, err)
	}
	parser := gofeed.NewParser() // lol race
	feed, err := parser.Parse(resp.Body)
	resp.Body.Close()
	<-f.sema
	if err != nil {
		return LinkedFeed{}, fmt.Errorf("Error parsing %s : %w", url, err)
	}
	jsonBuf, err := json.Marshal(feed)
	if err != nil {
		return LinkedFeed{}, fmt.Errorf("Error marshalling %s : %w", url, err)
	}
	err = f.itemCache.Write(encodeFunc(url), jsonBuf)
	if err != nil {
		return LinkedFeed{}, fmt.Errorf("Error writing to item cache %s : %w", url, err)
	}

	linkedFeed := LinkFeed(feed)
	if linkedFeed.Feed.FeedLink != url {
		log.Debugf("feed request url and self-reference url mismatch: requested %s, got %s", url, linkedFeed.Feed.FeedLink)
	}
	return linkedFeed, nil
}
