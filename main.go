package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/dgraph-io/badger"
	"github.com/mmcdole/gofeed"
	"github.com/pkg/profile"

	"github.com/bartmeuris/progressio"
	log "github.com/sirupsen/logrus"
	"github.com/vmihailenco/msgpack/v5"
)

// Feeds holds RSS/Atom feeds and a database. It provides download/parse progress and a channel for feed updates

type Feeds struct {
	Feeds     map[string]gofeed.Feed
	DB        *badger.DB
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

// Fetch fetches urls
func (f *Feeds) Fetch(urls []string, progress chan multiProgress, errChan chan error) {
	parser := gofeed.NewParser()
	var wg sync.WaitGroup
	defer close(errChan)
	// defer close(progress)
	for _, v := range urls {
		wg.Add(1)
		go func(e chan error, url string, p chan multiProgress, wg *sync.WaitGroup) {
			defer wg.Done()
			defer log.Println(url, "done")
			resp, err := http.Get(url)
			if err != nil {
				e <- fmt.Errorf("http.Get error on %s: %w", url, err)
			}
			progressReader, tmp := progressio.NewProgressReader(resp.Body, resp.ContentLength)
			defer progressReader.Close()
			go func() {
				for v := range tmp {
					p <- multiProgress{url: url, v: v}
				}
			}()
			feed, err := parser.Parse(progressReader)
			if err != nil {
				e <- fmt.Errorf("parse error on %s: %w", url, err)
			}
			if feed != nil {
				log.Warn("url:", url)
				if f.Feeds == nil {
					f.Feeds = make(map[string]gofeed.Feed, 0)
				}
				f.Feeds[url] = *feed
				if f.DB != nil {
					a, err := msgpack.Marshal(feed)
					if err != nil {
						e <- fmt.Errorf("msgpack.Marshal error on %s: %w", url, err)
					}
					log.Warnf("msgpack encoded length of %s: %d\n", url, len(a))
					err = f.DB.Update(func(txn *badger.Txn) error {
						log.Infof("ADDING %s to cache\n", url)
						txn.Set([]byte(url), a)
						return nil
					})
				}
			}
		}(errChan, v, progress, &wg)
	}
	wg.Wait()
}

func main() {
	defer profile.Start().Stop()
	var f Feeds
	c := make(chan multiProgress)
	e := make(chan error)
	dbt, err := badger.Open(badger.DefaultOptions("test.db"))
	if err != nil {
		log.Fatal(err)
	}
	defer dbt.Close()
	f.DB = dbt
	go func() {
		for v := range c {
			log.Trace(v)
		}
		close(c)
	}()
	go f.Fetch([]string{
		"https://www.arp242.net/feed.xml",
		"https://reddit.com/golang.rss",
		"https://skuz.xyz/randomRSS/rss?items=100000?seed=000000",
		"https://golangcode.com/index.xml",
		"https://dave.cheney.net/feed/atom",
	}, c, e)
	for err := range e {
		log.Warn(err)
	}
}
