package feeds

import "github.com/mmcdole/gofeed"

type LinkedFeedItem struct {
	Feed *gofeed.Feed
	*gofeed.Item
}

func LinkFeed(f *gofeed.Feed) []*LinkedFeedItem {
	b := make([]*LinkedFeedItem, len(f.Items))
	for i, item := range f.Items {
		b[i] = &LinkedFeedItem{Feed: f, Item: item}
	}
	return b
}
