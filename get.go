package main

import (
	"context"
	"sync"

	log "github.com/sirupsen/logrus"
)

func Get(ctx context.Context, p GetParam, out chan<- LinkedFeedItem, errChan chan<- error) {
	// Send work down a channel
	var wg sync.WaitGroup
	for _, u := range p.urls {
		wg.Add(1)
		// TODO: match titles and stuff. But for that we need to fetch feed first ($$$)
		go func(u string) {
			defer wg.Done()

			log.Infoln("Fetching feed", u) // TODO: nicer progress?
			lf, err := p.Fetch(ctx, u)
			if err != nil {
				errChan <- err
				return
			}
			for _, i := range lf.Items {
				out <- i
			}
		}(u)
	}
	wg.Wait()
}
