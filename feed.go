package main

import (
	"fmt"

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
