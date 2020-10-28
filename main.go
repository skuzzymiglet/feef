package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/mmcdole/gofeed"
)

// LinkedFeedItem is an item linked to its parent feed
type LinkedFeedItem struct {
	Feed *gofeed.Feed
	*gofeed.Item
}

// LinkedFeed is a feed whose items link to itself
type LinkedFeed struct {
	*gofeed.Feed
	Items []LinkedFeedItem
}

// LinkFeed links a feed's items and returns a LinkedFeed containing them
func LinkFeed(f *gofeed.Feed) LinkedFeed {
	b := make([]LinkedFeedItem, len(f.Items))
	for i, item := range f.Items {
		b[i] = LinkedFeedItem{Feed: f, Item: item}
	}
	return LinkedFeed{
		Feed:  f,
		Items: b,
	}
}

// Format implements fmt.Formatter
func (f LinkedFeed) Format(s fmt.State, c rune) {
	switch c {
	case 'd':
		fmt.Fprint(s, f.Description)
	case 't':
		fmt.Fprint(s, f.Title)
	case 'u':
		fmt.Fprint(s, f.Link)
	case 'i':
		fmt.Fprint(s, f.Link)
	}
}

// Format implements fmt.Formatter
func (f LinkedFeedItem) Format(s fmt.State, c rune) {
	switch c {
	case 'd':
		fmt.Fprint(s, f.Description)
	case 't':
		fmt.Fprint(s, f.Title)
	case 'u':
		fmt.Fprint(s, f.Link)
	case 'i':
		if f.GUID != "" {
			fmt.Fprint(s, f.GUID)
		} else {
			fmt.Fprint(s, f.Link)
		}
	}
}

func (f LinkedFeedItem) String() string {
	if f.GUID != "" {
		return f.Feed.Link + "~" + f.GUID
	}
	if f.Link != "" {
		return f.Feed.Link + "~" + f.Link
	}
	return "!(NO LINK OR GUID)"
}
func (f LinkedFeed) String() string {
	return f.Feed.Link
}

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

func main() {
	urls := flag.String("u", "urls", "file with newline delimited URLs")
	query := flag.String("q", "", "query (feed~item)")
	format := flag.String("f", "%t %u", "format for printing")
	flag.Parse()
	switch os.Args[1] {
	case "list":
		urlsFile, err := os.Open(*urls)
		if err != nil {
			panic(err)
		}
		scanner := bufio.NewScanner(urlsFile)
		var wg sync.WaitGroup
		for scanner.Scan() {
			wg.Add(1)
			go func(u string, wg *sync.WaitGroup) {
				parser := gofeed.NewParser() // fucking race conditions
				defer wg.Done()
				resp, err := http.Get(u)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error downloading %s: %s\n", u, err)
					return
				}
				defer resp.Body.Close()
				feed, err := parser.Parse(resp.Body)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error parsing %s: %s\n", u, err)
					return
				}
				lf := LinkFeed(feed)
				fmt.Printf(*format, lf)
				fmt.Println()
			}(scanner.Text(), &wg)
		}
		wg.Wait()
	case "feeds":
	case "show":
		item, err := FindItem(*query)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf(*format, item)
		fmt.Println()
	}
}
