package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/bartmeuris/progressio"
	"github.com/mmcdole/gofeed"
)

// Feeds holds RSS/Atom feeds
type Feeds struct {
	Feeds      map[string]gofeed.Feed
	updates    chan *gofeed.Feed
	httpClient *http.Client
}

type MultiProgress struct {
	url string
	v   progressio.Progress
}

// Fetch fetches the feeds from the URLs specified, and sends download progress and errors on channels
func (f *Feeds) Fetch(urls []string, progress chan MultiProgress, errChan chan error) {
	if f.Feeds == nil {
		f.Feeds = make(map[string]gofeed.Feed, 0)
	}
	var wg sync.WaitGroup
	for _, v := range urls {
		wg.Add(1)
		go func(e chan error, url string, p chan MultiProgress, wg *sync.WaitGroup) {
			defer wg.Done()
			var feed *gofeed.Feed
			if f.httpClient != nil {
				resp, err := f.httpClient.Get(url)
				if err != nil {
					e <- fmt.Errorf("http.Get error on %s: %w", url, err)
					return
				}
				progressReader, tmp := progressio.NewProgressReader(resp.Body, resp.ContentLength)
				defer progressReader.Close()
				go func() {
					for v := range tmp {
						p <- MultiProgress{url: url, v: v}
					}
				}()
				parser := gofeed.NewParser()
				feed, err = parser.Parse(progressReader)
				if err != nil {
					err := fmt.Errorf("parse error on %s: %w", url, err)
					e <- err
				}
			}
			if feed != nil {
				f.Feeds[url] = *feed
			}
		}(errChan, v, progress, &wg)
	}
	wg.Wait()
}
