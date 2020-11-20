package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"

	"github.com/gobwas/glob"
	"github.com/mmcdole/gofeed"
)

const delim = "~"

func Get(url string) (LinkedFeed, error) {
	// TODO: reuse parser and http client
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

func FindItems(feed, item string, urls []string) ([]LinkedFeedItem, error) {
	// TODO: globs

	work := make(chan LinkedFeed)
	errChan := make(chan error)
	go func() {
		sema := make(chan struct{}, 10) // TODO: make number of downloader threads configurable
		// Send work down a channel
		if t, err := url.ParseRequestURI(feed); err == nil && t.IsAbs() {
			lf, err := Get(feed)
			if err != nil {
				errChan <- err
				return
			}
			work <- lf
		} else {
			glob, err := glob.Compile(feed)
			if err != nil {
				errChan <- fmt.Errorf("error compiling feed glob \"%s\": %w", feed, err)
				return
			}
			var wg sync.WaitGroup
			for i, u := range urls {
				// TODO: match titles and stuff. But for that we need to fetch feed first ($$$)
				if glob.Match(u) {
					wg.Add(1)
					go func(sema chan struct{}, wg *sync.WaitGroup, i int, u string) {
						defer wg.Done()
						sema <- struct{}{}
						lf, err := Get(u)
						if err != nil {
							errChan <- err
						}
						work <- lf
						<-sema
					}(sema, &wg, i, u)
				}
			}
			wg.Wait()
		}
		close(work)
	}()

	glob, err := glob.Compile(item)
	if err != nil {
		return nil, fmt.Errorf("error compiling item glob \"%s\": %w", feed, err)
	}
	buf := make([]LinkedFeedItem, 0)
	for {
		select {
		case err := <-errChan:
			log.Println(err)
		case feed, more := <-work:
			if !more {
				if len(buf) == 0 {
					return []LinkedFeedItem{}, errors.New("Feed item not found")
					// TODO: be nicer to the user when they don't specify an urls file and nothing's found!
				}
				sort.Slice(buf, func(i, j int) bool {
					var it, jt time.Time
					if buf[i].PublishedParsed != nil {
						it = *buf[i].PublishedParsed
					} else if buf[i].UpdatedParsed != nil {
						it = *buf[i].UpdatedParsed
					} else {
						panic("error sorting feed: item does not include an update or published time")
					}
					if buf[j].PublishedParsed != nil {
						jt = *buf[j].PublishedParsed
					} else if buf[j].UpdatedParsed != nil {
						jt = *buf[j].UpdatedParsed
					} else {
						panic("error sorting feed: item does not include an update or published time")
					}
					return jt.Before(it) // Newest first, so comparator is upside-down
				})
				return buf, nil
			}
			for _, i := range feed.Items {
				matched := true
				switch {
				// case glob.Match(i.GUID):
				// 	log.Printf("%s matched GUID %s", item, i.GUID)
				case glob.Match(i.Link):
					// log.Printf("\"%s\" matched link \"%s\"", item, i.Link)
				case glob.Match(i.Title):
					// log.Printf("\"%s\" matched title \"%s\"", item, i.Title)
				case (i.Author != nil && glob.Match(i.Author.Name)):
					// log.Printf("\"%s\" matched author name \"%s\"", item, i.Author.Name)
				case (i.Author != nil && glob.Match(i.Author.Email)):
					// log.Printf("\"%s\" matched author email \"%s\"", item, i.Author.Email)
				default:
					matched = false
				}
				if matched {
					buf = append(buf, i)
				}
			}
		}
	}
}
