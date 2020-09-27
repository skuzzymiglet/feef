package main

import (
	"os"
	"testing"
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
	f := InitFeeds()
	f.Logger = NewLogger(nil, os.Stdout)
	progressChan := make(chan MultiProgress)
	defer close(progressChan)
	errChan := make(chan error)
	defer close(errChan)
	done := make(chan bool)
	go func() {
		f.Fetch(urls, progressChan, errChan)
		done <- true
	}()
	for {
		select {
		case p := <-progressChan:
			t.Logf("%s: %0.f%% transferred\n", p.url, p.v.Percent)
		case e := <-errChan:
			t.Log(e)
		case <-done:
            os.Exit(0)
		}
	}
}
