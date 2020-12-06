package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
)

func Get(ctx context.Context, p GetParam, out chan<- LinkedFeedItem, errChan chan error) {
	sema := make(chan struct{}, p.maxThreads)
	// Send work down a channel
	var wg sync.WaitGroup
	for _, u := range p.urls {
		wg.Add(1)
		// TODO: match titles and stuff. But for that we need to fetch feed first ($$$)
		go func(u string) {
			defer wg.Done()
			sema <- struct{}{}
			log.Infoln("Fetching feed", u) // TODO: nicer progress?
			parser := gofeed.NewParser()   // lol race
			resp, err := p.client.Get(u)
			if err != nil {
				errChan <- fmt.Errorf("Error fetching %s : %w", u, err)
				return
			}
			f, err := parser.Parse(resp.Body)
			resp.Body.Close()
			if err != nil {
				errChan <- fmt.Errorf("Error parsing %s : %w", u, err)
				return
			}
			lf := LinkFeed(f)
			if lf.Feed.FeedLink != u {
				log.Debugf("feed request url and self-reference url mismatch: requested %s, got %s", u, lf.Feed.FeedLink)
			}
			for _, i := range lf.Items {
				out <- i
			}
			<-sema
		}(u)
	}
	wg.Wait()
}
