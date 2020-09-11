package main

import (
	"bufio"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/mmcdole/gofeed"

	"github.com/bartmeuris/progressio"
	log "github.com/sirupsen/logrus"
)

// Feeds holds RSS/Atom feeds and a database (in the future). It provides download/parse progress and a channel for feed updates
type Feeds struct {
	Feeds map[string]gofeed.Feed
	// DB         *badger.DB // Currently unused
	updates    chan *gofeed.Feed
	httpClient *http.Client
	// gauges     map[string]*widgets.Gauge
}

type multiProgress struct {
	url string
	v   progressio.Progress
}

// Fetch fetches urls
func (f *Feeds) Fetch(urls []string, progress chan multiProgress, errChan chan error) {
	var wg sync.WaitGroup
	defer close(errChan)
	// defer close(progress)
	for _, v := range urls {
		wg.Add(1)
		go func(e chan error, url string, p chan multiProgress, wg *sync.WaitGroup) {
			defer wg.Done()
			defer log.Println(url, "done")
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
	log.SetLevel(log.FatalLevel)
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
	gauges := make(map[string]*widgets.Gauge, len(urls))
	// Boxes for the gauges
	boxes := make(map[string][4]int, len(urls))
	tabpaneHeight := 3
	var cx, cy int
	cy = tabpaneHeight
	w, h := termui.TerminalDimensions()
	gaugeHeight := int(math.Floor(float64(h) / float64(len(urls))))
	for _, v := range urls {
		// x1, y1, x2, y2
		boxes[v] = [4]int{cx, cy, cx + w, cy + gaugeHeight}
		cy += gaugeHeight
	}
	uiEvents := termui.PollEvents()
	tabpane := widgets.NewTabPane("unread", "progress")
	tabpane.SetRect(0, 0, w, tabpaneHeight)
	tabpane.Border = true
	tabpane.ActiveTabStyle = termui.Style{
		Fg: 1,
		Bg: 23,
	}
	tabpane.InactiveTabStyle = termui.Style{
		Bg: 1,
		Fg: 23,
	}
	termui.Render(tabpane)

	currentTab := 2
	for {
		select {
		case ev := <-uiEvents:
			switch ev.ID {
			case "q":
				log.Println("Quitting")
				os.Exit(0)
			case "1", "2":
				currentTab, err := strconv.Atoi(ev.ID)
				if err != nil {
					log.Warn(err)
				}
				tabpane.ActiveTabIndex = currentTab
				log.Fatal("rendering tabpane")
				termui.Render(tabpane)
				termui.Clear()
				if currentTab == 2 {
					for _, v := range gauges {
						termui.Render(v)
					}
				}
			}
		case v := <-progressChan:
			if currentTab == 2 {
				gauges[v.url] = widgets.NewGauge()
				gauges[v.url].Title = v.url
				gauges[v.url].Percent = int(v.v.Percent)
				gauges[v.url].SetRect(boxes[v.url][0], boxes[v.url][1], boxes[v.url][2], boxes[v.url][3])
				termui.Render(gauges[v.url])
			}
		case e := <-errChan:
			log.Warn(e)
		}
	}
}
