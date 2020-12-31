package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
)

func findNewItems(oldFeed, newFeed LinkedFeed) []LinkedFeedItem {
	var buf []LinkedFeedItem
	if len(newFeed.Items) > len(oldFeed.Items) {
		tmp := make(map[LinkedFeedItem]struct{}, len(newFeed.Items))

		for _, i := range oldFeed.Items {
			tmp[i] = struct{}{}
		}
		for _, i := range newFeed.Items { // For each new...
			if _, found := tmp[i]; !found {
				buf = append(buf, i)
			}
		}
	}
	return buf
}

func Notify(ctx context.Context, n NotifyParam, out chan<- LinkedFeedItem, errChan chan error) {
	sema := make(chan struct{}, n.maxThreads)
	var wg sync.WaitGroup
	for _, u := range n.urls {
		// TODO: don't download feeds if they weren't modified.
		// When only showing new items, fetch the initial feed
		// Othwerwise start with nothing
		initial := n.mode == newItems

		var last LinkedFeed
		wg.Add(1)
		go func(u string) {
			parser := gofeed.NewParser() // lol race
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:

					log.Debugln("refreshing", u)
					var body bytes.Reader
					req, err := http.NewRequestWithContext(ctx, "GET", u, &body)
					if err != nil {
						errChan <- fmt.Errorf("Error creating request for %s : %w", u, err)
						return
					}
					sema <- struct{}{}
					resp, err := n.client.Do(req)
					if err != nil {
						errChan <- fmt.Errorf("Error fetching %s : %w", u, err)
						return
					}

					feed, err := parser.Parse(resp.Body)
					resp.Body.Close()
					<-sema
					if err != nil {
						errChan <- fmt.Errorf("error parsing %s: %w", u, err)
						return
					}

					lf := LinkFeed(feed)
					if lf.Feed.FeedLink != u {
						log.Debugf("feed request url and self-reference url mismatch: requested %s, got %s", u, lf.Feed.FeedLink)
					}

					if initial { // Don't compare
						initial = false
					} else {
						newItems := findNewItems(last, lf)
						for _, item := range newItems {
							out <- item
						}
						last = lf

						time.Sleep(n.poll)
					}
				}
			}
		}(u)
	}
	wg.Wait()
}
