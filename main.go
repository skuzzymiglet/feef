package main

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gizak/termui/v3"
	"github.com/pkg/profile"
	"github.com/sirupsen/logrus"
)

var urls = []string{
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

func main() {
	// TODO: parse flags

	defer profile.Start().Stop()
	if err := termui.Init(); err != nil {
		logrus.Fatal(err)
	}
	defer termui.Close()
	log := NewLogger()
	tabs := InitTabs()
	tabs.Render(0)

	// Create Feeds
	var f Feeds
	progressChan := make(chan MultiProgress)
	errChan := make(chan error)
	f.httpClient = &http.Client{}
	// Start fetch
	fetching := make(chan time.Duration)
	// time.Time = time to fetch
	// 0 = just started
	// TODO: more elegant solution, I guess
	go func() {
		s := time.Now()
		fetching <- 0
		f.Fetch(urls, progressChan, errChan)
		fetching <- time.Now().Sub(s)
		for range time.Tick(time.Second * 5) {
			s = time.Now()
			fetching <- 0
			f.Fetch(urls, progressChan, errChan)
			fetching <- time.Now().Sub(s)
		}
	}()
	uiEvents := termui.PollEvents()

	for {
		select {
		case ev := <-uiEvents:
			switch ev.Type {
			case termui.KeyboardEvent:
				switch ev.ID {
				case "q":
					os.Exit(0)
					// TODO: show cursor again when exiting
				case "1", "2", "3", "4", "5":
					currentTab, err := strconv.Atoi(ev.ID)
					if err != nil {
						panic(err)
					}
					tabs.Go(currentTab - 1)
				}
			case termui.ResizeEvent:
				// TODO: resize hook
				// Have a list of Drawables
				// Call SetRect and render on each
				tabs.Refresh()
			}
		case e := <-errChan:
			if e != nil {
				log.Warn(e)
			}
		case t := <-fetching:
			if t == 0 {
				log.Println("fetching", len(urls), "feeds...")
			} else {
				log.Println("fetched", len(urls), "feeds in", t)
			}
		}
	}
}
