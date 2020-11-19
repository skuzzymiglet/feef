package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

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
	buf := make([]LinkedFeedItem, 0)         // For searching
	feeds := make([]LinkedFeed, len(urls)+1) // All feeds
	sema := make(chan struct{}, 6)
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
	if len(buf) != 0 {
		return buf, nil
	}
	return []LinkedFeedItem{}, errors.New("Feed item not found")
}

func Find(query string, v *[]LinkedFeedItem, urls []string) error {
	// TODO: cache
	var err error
	parts := strings.SplitN(query, delim, 2)
	switch len(parts) {
	// case 1:
	// 	v, err = FindFeeds(parts[0], urls)
	case 2:
		*v, err = FindItems(parts[0], parts[1], urls)
	default:
		err = errors.New("Invalid query")
	}
	return err
}
