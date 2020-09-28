package main

import (
	"bufio"
	"flag"
	"os"
	"strconv"
	"time"

	"github.com/gizak/termui/v3"
	"github.com/pkg/profile"
	"github.com/sirupsen/logrus"
)

func main() {
	var (
		urlsFile        string
		refreshInterval time.Duration
		logFileName     string
	)

	flag.StringVar(&urlsFile, "d", "", "file with URLs (one per line)")
	flag.StringVar(&logFileName, "l", "", "log file")
	flag.DurationVar(&refreshInterval, "r", time.Second*60, "time between refreshes")
	flag.Parse()

	urls := make([]string, 0)
	uf, err := os.Open(urlsFile)
	if err != nil {
		logrus.Fatal(err)
	}
	s := bufio.NewScanner(uf)
	for s.Scan() {
		urls = append(urls, s.Text())
	}
	uf.Close()

	defer profile.Start().Stop()

	if err := termui.Init(); err != nil {
		logrus.Fatal(err)
	}
	defer termui.Close()

	tabs := InitTabs()
	logFile, err := os.Create("feef.log")
	if err != nil {
		logrus.Fatal(err)
	}
	defer logFile.Close()
	log := NewLogger(tabs.messageBox, logFile)
	tabs.Go(0)

	// Create Feeds
	f := InitFeeds()
	f.Logger = log

	progressChan := make(chan MultiProgress)
	errChan := make(chan error)

	// Start fetch
	fetching := make(chan time.Duration)
	// time.Time = time to fetch
	// 0 = just started
	// TODO: make Fetch() itself return a FetchStatus struct
	go func() {
		s := time.Now()
		fetching <- 0
		f.Fetch(urls, progressChan, errChan)
		fetching <- time.Now().Sub(s)
		for range time.Tick(refreshInterval) {
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
					goto end
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
		case p := <-progressChan:
			log.WithFields(logrus.Fields{
				"url":     p.url,
				"percent": p.v.Percent,
			}).Debug()
		case t := <-fetching:
			if t == 0 {
				log.Println("fetching", len(urls), "feeds...")
			} else {
				log.Println("fetched", len(urls), "feeds in", t)
			}
		}
	}
end:
}
