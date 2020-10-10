package feeds

import (
	"net/http"
	"testing"

	"github.com/mmcdole/gofeed"
)

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
	l := LinkFeed(f)
	if l[0].Feed != f {
		t.Fatalf("LinkedFeed's parent is not correct: %#v", l[0].Feed)
	}
	if l[0].Item != f.Items[0] {
		t.Fatalf("LinkedFeed's item 0 is not same as parent feed's item 0: %#v", f.Items[0])
	}
}
