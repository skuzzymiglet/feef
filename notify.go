package main

import (
	"context"
	"time"

	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
)

func Notify(ctx context.Context, n NotifyParam, out chan LinkedFeedItem, errChan chan error) {
	sema := make(chan struct{}, n.maxThreads)
	for _, u := range n.urls {
		// When only showing new items, fetch the initial feed
		// Othwerwise start with nothing
		initial := n.mode == newItems

		var last LinkedFeed
		// var lastTime time.Time // 0, initially download everything
		go func(u string) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					sema <- struct{}{}
					log.Debugln("refreshing", u)
					parser := gofeed.NewParser() // lol race
					resp, err := n.client.Get(u)
					if err != nil {
						errChan <- err
						return
					}
					f, err := parser.Parse(resp.Body)
					resp.Body.Close()
					if err != nil {
						errChan <- err
						return
					}
					lf := LinkFeed(f)
					if lf.Feed.FeedLink != u {
						log.Debugf("feed request url and self-reference url mismatch: requested %s, got %s", u, lf.Feed.FeedLink)

					}
					if initial { // Don't compare
						initial = false
					} else {
						for _, i := range lf.Items { // For each new...
							matched := false
							for _, j := range last.Items { // Compare each old
								if i.GUID == j.GUID {
									matched = true
								}
							}
							if !matched {
								log.Debugf("found new item %s", i.GUID)
								out <- i
							}
						}
					}
					last = lf
					<-sema
					time.Sleep(n.poll)
				}
			}
		}(u)
	}
}
