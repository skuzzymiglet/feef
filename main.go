package main

import (
	"bufio"
	"flag"
	"net/http"
	"os"
	"strconv"
	"sync"
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

	uiEvents := termui.PollEvents()

	// Fetch!

	errChan := make(chan error, 0)
	client := &http.Client{}
	var wg sync.WaitGroup

	go func() {
		for i, u := range urls {
			wg.Add(1)
			go func(i int, u string, wg *sync.WaitGroup) {
				defer wg.Done()
				_, err := Fetch(u, client)
				if err != nil {
					errChan <- err
				}
			}(i, u, &wg)
		}
	}()

	for {
		select {
		case ev := <-uiEvents:
			switch ev.Type {
			case termui.KeyboardEvent:
				switch ev.ID {
				case "q":
					return
				case "1", "2", "3", "4", "5":
					currentTab, err := strconv.Atoi(ev.ID)
					if err != nil {
						panic(err)
					}
					tabs.Go(currentTab - 1)
				}
			case termui.ResizeEvent:
				// TODO: Have a function to redraw multiple Drawables
				// Have a list of Drawables
				// Call SetRect and render on each
				// TODO: Use another TUI library which handles resizes
				// maybe fork termui
				tabs.Refresh()
			}
		case e := <-errChan:
			if e != nil {
				log.Warn(e)
			}
		}
	}
}
