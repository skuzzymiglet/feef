package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/mmcdole/gofeed"
)

// FindItem finds a feed item
// q should be a in the format `feed~item`
// If fuzzy searches if feed or item is not a URL
func FindItem(q string) (LinkedFeedItem, error) {
	// TODO: cache
	var feed, item string
	parts := strings.SplitN(q, "~", 2)
	if len(parts) != 2 {
		return LinkedFeedItem{}, fmt.Errorf("Invalid query: \"%s\"", q)
	}
	feed = parts[0]
	item = parts[1]
	parser := gofeed.NewParser()
	// Fetch feed URL (exact)
	if _, err := url.ParseRequestURI(feed); err == nil {
		resp, err := http.Get(feed)
		if err != nil {
			return LinkedFeedItem{}, fmt.Errorf("FindItem: error downloading %s: %w", feed, err)
		}
		defer resp.Body.Close()
		f, err := parser.Parse(resp.Body)
		if err != nil {
			return LinkedFeedItem{}, fmt.Errorf("FindItem: error parsing %s: %w", feed, err)
		}
		lf := LinkFeed(f)
		// Search GUIDs/links (exact)
		for _, i := range lf.Items {
			if i.GUID == item || i.Link == item {
				return i, nil
			}
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
				return i, nil
			}
		}
	} else {
		// Search urls (this is expensive as hell)
		// I'm not implementing this until we have a cache
		// for _, u := range urls {
		// }
	}
	return LinkedFeedItem{}, errors.New("Feed item not found")
}

// func FindFeed(q string) (LinkedFeedItem, error) {
// 	return LinkedFeedItem{}, errors.New("Feed item not found")
// }
