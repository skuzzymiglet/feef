package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
)

// Get fetches an URL into a LinkedFeed
func Get(url string) (LinkedFeed, error) {
	// TODO: reuse parser and http client
	// TODO: correct user-agent for reddit
	parser := gofeed.NewParser() // lol race
	resp, err := http.Get(url)
	if err != nil {
		return LinkedFeed{}, fmt.Errorf("error downloading %s: %w", url, err)
	}
	f, err := parser.Parse(resp.Body)
	resp.Body.Close()
	if err != nil {
		return LinkedFeed{}, fmt.Errorf("error parsing %s: %w", url, err)
	}
	return LinkFeed(f), nil
}

// TODO: use contexts for cancellation of stages in the pipeline

func GetAll(urls []string, threads int, out chan LinkedFeedItem, errChan chan error) {
	sema := make(chan struct{}, threads) // TODO: make number of downloader threads configurable
	// Send work down a channel
	var wg sync.WaitGroup
	for _, u := range urls {
		wg.Add(1)
		// TODO: match titles and stuff. But for that we need to fetch feed first ($$$)
		go func(u string) {
			defer wg.Done()
			sema <- struct{}{}
			log.Debugln("refreshing feed", u)
			lf, err := Get(u)
			if err != nil {
				errChan <- err
				return
			}
			if lf.Feed.FeedLink != u {
				log.Debugf("feed request url and self-reference url mismatch: requested %s, got %s", u, lf.Feed.FeedLink)
			}
			for _, i := range lf.Items {
				log.Debugf("found new item in %s: %s", u, i.GUID)
				out <- i
			}
			<-sema
		}(u)
	}
	wg.Wait()
}
