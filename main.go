package main

import (
	"bufio"
	"flag"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gizak/termui/v3"
	"github.com/pkg/profile"
)

func main() {
	log := NewLogger(os.Stdout)

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
	if urlsFile == "" {
		flag.Usage()
		os.Exit(2)
	}
	uf, err := os.Open(urlsFile)
	if err != nil {
		log.Fatal(err)
	}
	s := bufio.NewScanner(uf)
	for s.Scan() {
		urls = append(urls, s.Text())
	}
	uf.Close()

	defer profile.Start().Stop()

	if err := termui.Init(); err != nil {
		log.Fatal(err)
	}
	defer termui.Close()

	tabs := InitTabs()
	var logFile io.Writer
	if logFileName != "" {
		logFile, err = os.Create(logFileName)
		if err != nil {
			log.Fatal(err)
		}
		defer logFile.(*os.File).Close()
	} else {
		logFile = ioutil.Discard
	}
	log.SetOutput(logFile)

	log.AddHook(&BarMessageHook{
		b: tabs.messageBox,
	})
	tabs.Go(0)

	uiEvents := termui.PollEvents()

	// Fetch!

	errChan := make(chan error, 0)
	client := &http.Client{}

	fetchStatus := make(chan time.Duration)
	go func() {
		s := time.Now()
		wg := sync.WaitGroup{}
		for _, u := range urls {
			wg.Add(1)
			go func(u string, wg *sync.WaitGroup) {
				defer wg.Done()
				_, err := Fetch(u, client)
				if err != nil {
					errChan <- err
				}
			}(u, &wg)
		}
		wg.Wait()
		fetchStatus <- time.Now().Sub(s)
		for range time.Tick(refreshInterval) {
			s := time.Now()
			var wg sync.WaitGroup
			for _, u := range urls {
				wg.Add(1)
				go func(u string, wg *sync.WaitGroup) {
					defer wg.Done()
					_, err := Fetch(u, client)
					if err != nil {
						errChan <- err
					}
				}(u, &wg)
			}
			wg.Wait()
			fetchStatus <- time.Now().Sub(s)
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
				// TODO: sometimes only shows "parse error" and not url
				log.Warn(e)
			}
		case d := <-fetchStatus:
			log.Printf("fetched %d feeds in %s", len(urls), d)
		}
	}
}
