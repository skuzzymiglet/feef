package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/pkg/profile"
	log "github.com/sirupsen/logrus"
)

// func getFeed(url string, wg *sync.WaitGroup) (*gofeed.Feed, error) {
func getFeed(url string) (*gofeed.Feed, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	parser := gofeed.NewParser()
	b, err := ioutil.ReadAll(resp.Body)
	s := string(b)
	return parser.ParseString(s)
}

func main() {
	defer profile.Start().Stop()
	start := time.Now()
	urlsFile, err := os.Open("urls")
	if err != nil {
		log.Fatal(err)
	}
	urlNames, err := ioutil.ReadAll(urlsFile)
	urls := strings.Split(string(urlNames), "\n")
	urls = urls[0 : len(urls)-1]
	var wg sync.WaitGroup
	for i, v := range urls {
		wg.Add(1)
		go func(i int, v string, wg *sync.WaitGroup) {
			start := time.Now()
			_, err := getFeed(v)
			end := time.Now()
			if err != nil {
				log.Warnln(i, v, err)
			} else {
				log.Infof("Fetched feed %d (%s) in %s\n", i, v, end.Sub(start))
			}
			wg.Done()
		}(i, v, &wg)
	}
	wg.Wait()
	end := time.Now()
	log.Infoln("Fetched all in", end.Sub(start))
}
