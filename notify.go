package main

import (
	"context"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

func findNewItems(oldFeed, newFeed LinkedFeed) []LinkedFeedItem {
	var buf []LinkedFeedItem
	tmp := make(map[string]struct{}, len(newFeed.Items))

	for _, i := range oldFeed.Items {
		tmp[i.GUID] = struct{}{}
	}
	for _, i := range newFeed.Items { // For each new...
		if _, found := tmp[i.GUID]; !found {
			buf = append(buf, i)
		}
	}
	return buf
}

func Notify(ctx context.Context, n NotifyParam, out chan<- LinkedFeedItem, errChan chan error) {
	var wg sync.WaitGroup
	for _, u := range n.urls {
		// When only showing new items, fetch the initial feed
		// Othwerwise start with nothing
		initial := true
		var last LinkedFeed
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			// gofeed.Parser is not thread-safe
			for {
				select {
				case <-ctx.Done():
					return
				default:
					lf, err := n.Fetcher.Fetch(ctx, u)
					if err != nil {
						errChan <- err
						return
					}
					if lf.Feed.FeedLink != u {
						log.Debugf("feed request url and self-reference url mismatch: requested %s, got %s", u, lf.Feed.FeedLink)
					}
					if initial && n.mode == newItems {
						// immediately move on in "newItems" mode
						initial = false
					} else {
						newItems := findNewItems(last, lf)
						for _, item := range newItems {
							out <- item
						}
						time.Sleep(n.poll)
					}
					last = lf

				}
			}
		}(u)
	}
	wg.Wait()
}
