package main

import (
	"context"
	"sync"

	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
)

func GetAll(ctx context.Context, p GetParam, out chan LinkedFeedItem, errChan chan error) {
	sema := make(chan struct{}, p.maxThreads)
	// Send work down a channel
	var wg sync.WaitGroup
	for _, u := range p.urls {
		wg.Add(1)
		// TODO: match titles and stuff. But for that we need to fetch feed first ($$$)
		go func(u string) {
			defer wg.Done()
			sema <- struct{}{}
			log.Infoln("Downloading feed", u)
			parser := gofeed.NewParser() // lol race
			resp, err := p.client.Get(u)
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
			for _, i := range lf.Items {
				out <- i
			}
			<-sema
		}(u)
	}
	wg.Wait()
}
