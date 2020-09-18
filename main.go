package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
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
					return
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

type BarMessageHook struct {
	b *widgets.Paragraph
}

func (b *BarMessageHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
	}
}

func (b *BarMessageHook) Fire(l *logrus.Entry) error {
	var style termui.Style
	switch l.Level {
	case logrus.ErrorLevel:
		style.Fg = 9
		style.Modifier = termui.ModifierBold
	case logrus.WarnLevel:
		style.Fg = 11
		style.Modifier = termui.ModifierBold
	case logrus.InfoLevel:
		style.Fg = 15
		style.Modifier = termui.ModifierReverse
	}
	b.b.TextStyle = style
	b.b.Text = l.Message
	termui.Render(b.b)
	return nil
}

func main() {
	defer profile.Start().Stop()

	// Logging
	log := logrus.New()

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
	// Start fetch
	var doneFetching bool
	go func() {
		f.Fetch(urls, progressChan, errChan)
		doneFetching = true
	}()
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
	messages.SetRect(0, 0, w, barHeight)

	log.AddHook(&BarMessageHook{
		b: messages,
	})
	log.Out = ioutil.Discard

	tabs := []string{"new", "unread", "old", "queue", "jobs"}
	tabWidgets := make([][]termui.Drawable, 5)

	tabWidgets[0] = []termui.Drawable{widgets.NewParagraph()}
	tabWidgets[0][0].(*widgets.Paragraph).Text = "memes"
	tabWidgets[0][0].SetRect(0, barHeight, w, h)
	tabWidgets[1] = tabWidgets[0]
	tabWidgets[2] = tabWidgets[0]
	tabWidgets[3] = tabWidgets[0]

	// progress
	tabWidgets[4] = []termui.Drawable{
		widgets.NewGauge(),
	}
	tabWidgets[4][0].(*widgets.Gauge).Label = "fetch feeds"
	tabWidgets[4][0].(*widgets.Gauge).SetRect(0, barHeight, w, h)

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
	termui.Render(tabpane, tabWidgets[0][0])

	var totalPercent float64 = float64(100) * float64(len(urls))
	progressData := make(map[string]float64)

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
				termui.Render(tabpane)
				termui.Render(tabWidgets[currentTab-1]...)
			}
		case p := <-progressChan:
			if !doneFetching {
				progressData[p.url] = p.v.Percent
				var sum float64
				for _, v := range progressData {
					sum += v
				}
				tabWidgets[4][0].(*widgets.Gauge).Percent = int(100 * sum / totalPercent)
				if tabpane.ActiveTabIndex == 4 {
					termui.Render(tabWidgets[4][0])
				}
			} else {
				tabWidgets[4][0].(*widgets.Gauge).Percent = 0
			}
		case e, more := <-errChan:
			if more == true {
				log.Warn(e)
			}
		}
	}
}
