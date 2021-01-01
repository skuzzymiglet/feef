package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
)

func Get(ctx context.Context, p GetParam, out chan<- LinkedFeedItem, errChan chan error) {
	// TODO: significant code duplication between get.go and notify.go. Needs to be cleaned up for maintainability
	sema := make(chan struct{}, p.maxThreads)
	// Send work down a channel
	var wg sync.WaitGroup
	for _, u := range p.urls {
		wg.Add(1)
		// TODO: match titles and stuff. But for that we need to fetch feed first ($$$)
		go func(u string) {
			defer wg.Done()

			log.Infoln("Fetching feed", u) // TODO: nicer progress?
			parser := gofeed.NewParser()   // lol race
			var body bytes.Reader
			req, err := http.NewRequestWithContext(ctx, "GET", u, &body)
			if err != nil {
				errChan <- fmt.Errorf("Error creating request for %s : %w", u, err)
				return
			}
			sema <- struct{}{}
			resp, err := p.client.Do(req)
			if err != nil {
				errChan <- fmt.Errorf("Error fetching %s : %w", u, err)
				return
			}
			f, err := parser.Parse(resp.Body)
			resp.Body.Close()
			<-sema
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

		}(u)
	}
	wg.Wait()
}
