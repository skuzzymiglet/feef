package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
)

const delim = "~"

func Get(url string, feed *LinkedFeed) error {
	parser := gofeed.NewParser() // lol race
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error downloading %s: %w", url, err)
	}
	f, err := parser.Parse(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error parsing %s: %w", url, err)
	}
	*feed = LinkFeed(f)
	return nil
}

func FindItems(feed, item string, urls []string) ([]LinkedFeedItem, error) {
	// TODO: globs
	buf := make([]LinkedFeedItem, 0)         // Search results
	feeds := make([]LinkedFeed, len(urls)+1) // All feeds
	sema := make(chan struct{}, 10)
	// Fetch feed URL (exact)
	if feed == "*" {
		var wg sync.WaitGroup
		for i, u := range urls {
			wg.Add(1)
			go func(sema chan struct{}, wg *sync.WaitGroup, i int, u string) {
				defer wg.Done()
				sema <- struct{}{}
				err := Get(u, &feeds[i])
				if err != nil {
					log.Println(err)
				}
				<-sema
			}(sema, &wg, i, u)
		}
		wg.Wait()
	} else if _, err := url.ParseRequestURI(feed); err == nil {
		err := Get(feed, &feeds[0])
		if err != nil {
			return nil, err
		}
	}
	// Search GUIDs/links (exact)
	for _, feed := range feeds {
		if item == "*" {
			buf = append(buf, feed.Items...)
		} else {
			for _, i := range feed.Items {
				if i.GUID == item || i.Link == item {
					buf = append(buf, i)
				}

				var wordsMatched int
				words := strings.Split(item, " ")
				for _, word := range words {
					if strings.Contains(strings.ToLower(i.Title), strings.ToLower(word)) {
						wordsMatched++
					}
				}
				if wordsMatched >= len(words) {
					buf = append(buf, i)
				}
			}
		}
	}
	if len(buf) == 0 {
		return []LinkedFeedItem{}, errors.New("Feed item not found")
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
