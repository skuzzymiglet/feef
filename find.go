package main

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/gobwas/glob"
	"github.com/mmcdole/gofeed"
)

const delim = "~"

var ErrNotFound = errors.New("Feed item not found")

// Param holds query parameters
type Param struct {
	max     int
	urls    []string
	sort    bool
	item    glob.Glob
	feedURL glob.Glob
}

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
			lf, err := Get(u)
			if err != nil {
				errChan <- err
			}
			for _, i := range lf.Items {
				out <- i
			}
			<-sema
		}(u)
	}
	wg.Wait()
}

func Filter(p Param, in, out chan LinkedFeedItem, errChan chan error) {
	var buf []LinkedFeedItem
	for i := range in {
		matched := true
		switch {
		case p.item.Match(i.Link):
		case p.item.Match(i.Title):
		case (i.Author != nil && p.item.Match(i.Author.Name)):
		case (i.Author != nil && p.item.Match(i.Author.Email)):
		default:
			matched = false
		}
		if p.sort {
			buf = append(buf, i)
		} else if matched {
			out <- i
		}
	}
	if p.sort {
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
		if p.max == 0 || p.max >= len(buf) {
			for _, v := range buf {
				out <- v
			}
		} else if p.max <= len(buf) {
			for _, v := range buf[:p.max-1] {
				out <- v
			}
		}
	}
}
