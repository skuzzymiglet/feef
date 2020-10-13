package feeds

import (
	"net/http"
	"sync"

	"github.com/mmcdole/gofeed" // TODO: fork gofeed and remove the bloat
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

// LinkFeedItems links a feed's items and returns them
func LinkFeedItems(f *gofeed.Feed) []LinkedFeedItem {
	b := make([]LinkedFeedItem, len(f.Items))
	for i, item := range f.Items {
		b[i] = LinkedFeedItem{Feed: f, Item: item}
	}
	return b
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

// InitFeeds initializes a Feeds struct with a map, a gofeed parser an an HTTP Client
func InitFeeds() *Feeds {
	return &Feeds{
		feeds:      make(map[string]LinkedFeed),
		parser:     gofeed.NewParser(),
		httpClient: &http.Client{},
	}
}

// Fetch fetches an URL and stores it
func (f *Feeds) Fetch(url string) error {
	// Maybe we should elminate this method and fetch lazily when the user asks for an URL
	resp, err := f.httpClient.Get(url)
	if err != nil {
		return err
	}
	// BUG: sometimes there is a cap-out-of-range error on parsing. Needs reproduction
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
