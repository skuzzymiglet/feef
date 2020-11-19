package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/mmcdole/gofeed"
)

const delim = "~"

func FindItems(feed, item string, urls []string) ([]LinkedFeedItem, error) {
	buf := make([]LinkedFeedItem, 0)
	parser := gofeed.NewParser()
	// Fetch feed URL (exact)
	if _, err := url.ParseRequestURI(feed); err == nil {
		resp, err := http.Get(feed)
		if err != nil {
			return nil, fmt.Errorf("FindItem: error downloading %s: %w", feed, err)
		}
		f, err := parser.Parse(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("FindItem: error parsing %s: %w", feed, err)
		}
		lf := LinkFeed(f)
		// Search GUIDs/links (exact)
		for _, i := range lf.Items {
			if i.GUID == item || i.Link == item {
				buf = append(buf, i)
			}
		}
		if len(buf) != 0 {
			return buf, nil
		}
		// Fuzzy search titles
		for _, i := range lf.Items {
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
		if len(buf) != 0 {
			return buf, nil
		}
	} else {
		// Search urls (this is expensive as hell)
		// I'm not implementing this until we have a cache
		for range urls {
		}
	}
	return []LinkedFeedItem{}, errors.New("Feed item not found")
}

// func FindFeeds(feed string, urls []string) ([]gofeed.Feed, error) {
// 	return nil, nil
// }

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
