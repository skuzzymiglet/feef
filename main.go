package main

import (
	"net/http"
	"sync"

	"github.com/dgraph-io/badger"
	"github.com/mmcdole/gofeed"
	"github.com/pkg/profile"

	"github.com/bartmeuris/progressio"
	log "github.com/sirupsen/logrus"
)

// Feeds holds RSS/Atom feeds and a database. It provides download/parse progress and a channel for feed updates
type Feeds struct {
	Feeds     []gofeed.Feed
	db        *badger.DB
	updates   chan *gofeed.Feed
	httpMutex *sync.Mutex // Currently unused
}

type _ interface {
	NotifyUpdates() <-chan *gofeed.Feed
	Fetch(progressChans map[string]chan progressio.Progress) error
}

type multiProgress struct {
	url string
	v   progressio.Progress
}

// Fetch fetches the keys of progressChans and sends progress to their value
func (f *Feeds) Fetch(urls []string, progress chan multiProgress) error {
	parser := gofeed.NewParser()
	for _, v := range urls {
		var e error
		go func(e *error, url string, p chan multiProgress) {
			resp, err := http.Get(url)
			if err != nil {
				e = &err
				return
			}
			progressReader, tmp := progressio.NewProgressReader(resp.Body, resp.ContentLength)
			defer progressReader.Close()
			go func() {
				p <- multiProgress{url: url, v: <-tmp}
			}()
			f, err := parser.Parse(progressReader)
			if err != nil {
				e = &err
				return
			}
			log.Infoln(f)
			return
		}(&e, v, progress)
		if e != nil {
			return e
		}
	}
	return nil
}

func main() {
	defer profile.Start().Stop()
	var f Feeds
	c := make(chan multiProgress, 0)
	go func() {
		for v := range c {
			log.Info(v)
		}
	}()
	log.Warn(f.Fetch([]string{
		// "https://www.arp242.net/feed.xml",
		"https://reddit.com/golang.rss",
	}, c))
	select {}
}
