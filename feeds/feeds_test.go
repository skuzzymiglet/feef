package feeds

import (
	"net/http"
	"testing"

	"github.com/mmcdole/gofeed"
)

var urls []string = []string{
	"https://github.com/terminal-discord/weechat-discord/commits/master.atom",
	"https://github.com/qutebrowser/qutebrowser/releases.atom",
	"https://www.youtube.com/feeds/videos.xml?channel_id=UCJZTjBlrnDHYmf0F-eYXA3Q",
	"https://www.youtube.com/feeds/videos.xml?channel_id=UCr_Q-bPpcw5fJ-Oow1BW1NQ",
	"https://blog.golang.org/feed.atom?format=xml",
	"https://dave.cheney.net/feed/atom",
	"https://www.arp242.net/feed.xml",
	"https://buttondown.email/cryptography-dispatches/rss",
	"https://drewdevault.com/feed.xml",
	"https://www.youtube.com/feeds/videos.xml?channel_id=UCVQCQJyZQcIioTDQ4SACvZQ",
	"https://www.youtube.com/feeds/videos.xml?channel_id=UC8EQAfueDGNeqb1ALm0LjHA",
	"https://www.youtube.com/feeds/videos.xml?channel_id=UC8R8FRt1KcPiR-rtAflXmeg",
	"https://www.youtube.com/feeds/videos.xml?channel_id=UC7YOGHUfC1Tb6E4pudI9STA",
	"https://www.youtube.com/feeds/videos.xml?channel_id=UCsnGwSIHyoYN0kiINAGUKxg",
	"https://www.youtube.com/feeds/videos.xml?channel_id=UCtM5z2gkrGRuWd0JQMx76qA",
	"https://www.youtube.com/feeds/videos.xml?channel_id=UCS4FAVeYW_IaZqAbqhlvxlA",
	"https://www.youtube.com/feeds/videos.xml?channel_id=UCS0N5baNlQWJCUrhCEo8WlA",
	"https://www.youtube.com/feeds/videos.xml?channel_id=UCnkp4xDOwqqJD7sSM3xdUiQ",
	"https://www.youtube.com/feeds/videos.xml?channel_id=UCVQCQJyZQcIioTDQ4SACvZQ",
	"https://www.youtube.com/feeds/videos.xml?channel_id=UCf7D886oBahxSSwBRVIib0A",
	"https://www.youtube.com/feeds/videos.xml?channel_id=UCXkNod_JcH7PleOjwK_8rYQ",
	"https://www.youtube.com/feeds/videos.xml?channel_id=UCMk_WSPy3EE16aK5HLzCJzw",
	"https://www.youtube.com/feeds/videos.xml?channel_id=UCeh-pJYRZTBJDXMNZeWSUVA",
	"https://www.youtube.com/feeds/videos.xml?channel_id=UCSl5Uxu2LyaoAoMMGp6oTJA",
}

func TestLinkedFeed(t *testing.T) {
	parser := gofeed.NewParser()
	resp, err := http.Get("https://arp242.net/feed.xml")
	if err != nil {
		t.Fatal(err)
	}
	f, err := parser.Parse(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	li := LinkFeedItems(f)
	if li[0].Feed != f {
		t.Fatalf("LinkedFeed's parent is not correct: %#v", li[0].Feed)
	}
	if li[0].Item != f.Items[0] {
		t.Fatalf("LinkedFeed's item 0 is not same as parent feed's item 0: %#v", f.Items[0])
	}
	l := LinkFeed(f)
	if l.Feed != f {
		t.Fatalf("LinkedFeed's parent is not correct: %#v", l.Feed)
	}
	if l.Items[0].Item != f.Items[0] {
		t.Fatalf("LinkedFeed's item 0 is not same as parent feed's item 0: %#v", f.Items[0])
	}
}

func TestFetch(t *testing.T) {
	f := InitFeeds()
	for _, u := range urls {
		t.Log("fetching", u)
		err := f.Fetch(u)
		if err != nil {
			t.Fatal(err)
		}
	}
}
