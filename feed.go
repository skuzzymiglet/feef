package main

import (
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
