package feeds

import (
	"net/http"
	"sync"

	"github.com/mmcdole/gofeed"
)

type LinkedFeedItem struct {
	Feed *gofeed.Feed
	*gofeed.Item
}

type LinkedFeed struct {
	*gofeed.Feed
	Items []LinkedFeedItem
}

func LinkFeedItems(f *gofeed.Feed) []LinkedFeedItem {
	b := make([]LinkedFeedItem, len(f.Items))
	for i, item := range f.Items {
		b[i] = LinkedFeedItem{Feed: f, Item: item}
	}
	return b
}

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

// Feeds holds feeds. It should only be accessed by calling its methods
type Feeds struct {
	feeds      map[string]LinkedFeed
	mu         sync.RWMutex
	parser     *gofeed.Parser
	httpClient *http.Client
}

type _ interface {
	ItemsSince()
	Feeds() []LinkedFeed
	Fetch(url string) error
}

func InitFeeds() *Feeds {
	return &Feeds{
		feeds:      make(map[string]LinkedFeed),
		parser:     gofeed.NewParser(),
		httpClient: &http.Client{},
	}
}

func (f *Feeds) Fetch(url string) error {
	resp, err := f.httpClient.Get(url)
	if err != nil {
		return err
	}
	feed, err := f.parser.Parse(resp.Body)
	if err != nil {
		return err
	}
	f.mu.Lock()
	lf := LinkFeed(feed)
	f.feeds[url] = lf
	f.mu.Unlock()
	return nil
}
