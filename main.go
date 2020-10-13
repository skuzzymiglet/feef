package main

import (
	"bufio"
	"flag"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"git.sr.ht/~skuzzymiglet/feef/feeds"
	"git.sr.ht/~skuzzymiglet/feef/ui"

	"github.com/gizak/termui/v3"
	"github.com/kr/pretty"
	"github.com/pkg/profile"
	"github.com/sirupsen/logrus"
)

// NOTE: Main doesn't do much rn, because UI is not fully developed
func main() {
	log := logrus.New()

	var (
		urlsFile        string
		refreshInterval time.Duration
		logFileName     string
	)

	flag.StringVar(&urlsFile, "d", "/home/skuzzymiglet/.config/newsboat/urls", "file with URLs (one per line)")
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

	// Fetch!

	errChan := make(chan error, 0)

	f := feeds.InitFeeds()
	pretty.Println(f)

	go func() {
		wg := sync.WaitGroup{}
		// NOTE: This is basically a bad do-while
		// TODO: refactor to reduce code duplication
		for _, u := range urls {
			wg.Add(1)
			go func(u string, wg *sync.WaitGroup) {
				defer wg.Done()
				err := f.Fetch(u)
				if err != nil {
					errChan <- err
				}
			}(u, &wg)
		}
		wg.Wait()
		for range time.Tick(refreshInterval) {
			var wg sync.WaitGroup
			for _, u := range urls {
				wg.Add(1)
				go func(u string, wg *sync.WaitGroup) {
					defer wg.Done()
					err := f.Fetch(u)
					if err != nil {
						errChan <- err
					}
				}(u, &wg)
			}
			wg.Wait()
		}
	}()
	ui.RunUI(&ui.FeefUI{})
}
