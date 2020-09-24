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
// It is the caller's responsibility to provide unique URLs, to avoid a race condition on maps
func (f *Feeds) Fetch(urls []string, progress chan MultiProgress, errChan chan error) {
	var wg sync.WaitGroup
	// defer close(errChan)
	// defer close(progress)
	// panic(pretty.Sprint(urls))
	for _, v := range urls {
		wg.Add(1)
		go func(e chan error, url string, p chan MultiProgress, wg *sync.WaitGroup) {
			defer wg.Done()
			// f.gauges[url] = widgets.NewGauge()
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
						// f.gauges[url].Percent = int(v.Percent)
						// f.gauges[url].Title = url
					}
				}()
				parser := gofeed.NewParser()
				feed, err = parser.Parse(progressReader)
				if err != nil {
					e <- fmt.Errorf("parse error on %s: %w", url, err)
				}
			}
			if feed != nil {
				if f.Feeds == nil {
					f.Feeds = make(map[string]gofeed.Feed, 0)
				}
				f.Feeds[url] = *feed
			}
		}(errChan, v, progress, &wg)
	}
	wg.Wait()
}
