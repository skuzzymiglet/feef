package main

import (
	"net/http"
	"sync"
	"testing"
	"time"
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

func TestFetch(t *testing.T) {
	client := &http.Client{}
	var wg sync.WaitGroup
	for i, u := range urls {
		wg.Add(1)
		go func(i int, u string, wg *sync.WaitGroup) {
			defer wg.Done()
			s := time.Now()
			_, err := Fetch(u, client)
			t.Logf("fetched feed %d in %s (%s)", i, time.Now().Sub(s), u)
			if err != nil {
				t.Fatal(err)
			}
		}(i, u, &wg)
	}
	wg.Wait()
}
