package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/mmcdole/gofeed"
	"github.com/pkg/profile"

	"github.com/bartmeuris/progressio"
	"github.com/sirupsen/logrus"
)

// Feeds holds RSS/Atom feeds
type Feeds struct {
	Feeds      map[string]gofeed.Feed
	updates    chan *gofeed.Feed
	httpClient *http.Client
}

type multiProgress struct {
	url string
	v   progressio.Progress
}

func (f *Feeds) Fetch(urls []string, progress chan multiProgress, errChan chan error) {
	var wg sync.WaitGroup
	defer close(errChan)
	// defer close(progress)
	for _, v := range urls {
		wg.Add(1)
		go func(e chan error, url string, p chan multiProgress, wg *sync.WaitGroup) {
			defer wg.Done()
			// f.gauges[url] = widgets.NewGauge()
			var feed *gofeed.Feed
			if f.httpClient != nil {
				resp, err := f.httpClient.Get(url)
				if err != nil {
					e <- fmt.Errorf("http.Get error on %s: %w", url, err)
				}
				progressReader, tmp := progressio.NewProgressReader(resp.Body, resp.ContentLength)
				defer progressReader.Close()
				go func() {
					for v := range tmp {
						p <- multiProgress{url: url, v: v}
						// f.gauges[url].Percent = int(v.Percent)
						// f.gauges[url].Title = url
					}
				}()
				parser := gofeed.NewParser()
				feed, err = parser.Parse(progressReader)
				if err != nil {
					e <- fmt.Errorf("parse error on %s: %w", url, err)
				}
			}
			if feed != nil {
				if f.Feeds == nil {
					f.Feeds = make(map[string]gofeed.Feed, 0)
				}
				f.Feeds[url] = *feed
			}
		}(errChan, v, progress, &wg)
	}
	wg.Wait()
}

func main() {
	defer profile.Start().Stop()

	// Logging
	log := logrus.New()
	log.SetLevel(logrus.FatalLevel)

	// Create Feeds
	var f Feeds
	progressChan := make(chan multiProgress)
	errChan := make(chan error)
	f.httpClient = &http.Client{Transport: &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: time.Second * 5,
	}}
	// URLs to test with
	urls := make([]string, 0)
	urlsFile, err := os.Open("urls")
	if err != nil {
		log.Fatal(err)
	}
	s := bufio.NewScanner(urlsFile)
	for s.Scan() {
		urls = append(urls, s.Text())
	}
	// Start
	go f.Fetch(urls, progressChan, errChan)
	// termui
	if err := termui.Init(); err != nil {
		log.Fatal(err)
	}
	defer termui.Close()
	barHeight := 3
	w, h := termui.TerminalDimensions()
	uiEvents := termui.PollEvents()

	// Log messages
	messages := widgets.NewParagraph()
	messages.Text = "message"

	tabs := []string{"new", "unread", "old", "queue", "jobs"}
	tabWidgets := make([]termui.Drawable, 5)

	tabWidgets[0] = widgets.NewParagraph()
	tabWidgets[0].(*widgets.Paragraph).Text = "memes"
	tabWidgets[0].SetRect(0, barHeight, w, h)
	tabWidgets[1] = tabWidgets[0]
	tabWidgets[2] = tabWidgets[0]
	tabWidgets[3] = tabWidgets[0]
	tabWidgets[4] = tabWidgets[0]

	tabpane := widgets.NewTabPane(tabs...)
	tabpane.SetRect(0, 0, w, barHeight)
	tabpane.Border = true
	tabpane.ActiveTabStyle = termui.Style{
		Fg:       15,
		Bg:       0,
		Modifier: termui.ModifierBold,
	}
	tabpane.InactiveTabStyle = termui.Style{
		Fg: 15,
		Bg: 0,
	}
	termui.Render(tabpane, tabWidgets[0])

	for {
		select {
		case ev := <-uiEvents:
			switch ev.ID {
			case "q":
				log.Println("Quitting")
				os.Exit(0)
			case "1", "2", "3", "4", "5":
				currentTab, err := strconv.Atoi(ev.ID)
				if err != nil {
					panic(err)

				}
				tabpane.ActiveTabIndex = currentTab - 1
				termui.Render(tabpane, tabWidgets[currentTab-1])
			}
		case e := <-errChan:
			if e != nil {
				log.Warn(err)
			}
		}
	}
}
