package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gizak/termui/v3"
	"github.com/pkg/profile"
)

func main() {
	defer profile.Start().Stop()
	if err := termui.Init(); err != nil {
		log.Fatal(err)
	}
	defer termui.Close()
	log := NewLogger()
	tabs := InitTabs()
	tabs.Render(0)

	// Create Feeds
	var f Feeds
	progressChan := make(chan MultiProgress)
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
		// Is it unique?
		if func() bool {
			for _, v := range urls {
				if s.Text() == v {
					return false
				}
			}
			return true
		}() {
			urls = append(urls, s.Text())
		}
	}
	// Start fetch
	doneFetching := make(chan bool)
	go func() {
		f.Fetch(urls, progressChan, errChan)
		doneFetching <- true
	}()
	uiEvents := termui.PollEvents()
	go tabs.GaugeLoop(progressChan)

	for {
		select {
		case ev := <-uiEvents:
			log.Println("ui event")
			switch ev.Type {
			case termui.KeyboardEvent:
				switch ev.ID {
				case "q":
					panic(nil)
				case "1", "2", "3", "4", "5":
					currentTab, err := strconv.Atoi(ev.ID)
					if err != nil {
						panic(err)
					}
					tabs.Go(currentTab - 1)
				}
			case termui.ResizeEvent:
				tabs.Refresh()
			}
		// case <-doneFetching:
		// 	tabWidgets[4] = []termui.Drawable{}
		// 	termui.Render(tabWidgets[tabpane.ActiveTabIndex]...)
		case e := <-errChan:
			if e != nil {
				log.Warn(e)
			}
		}
	}
}
