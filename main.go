package main

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/pkg/profile"
	log "github.com/sirupsen/logrus"
)

// func getFeed(url string, wg *sync.WaitGroup) (*gofeed.Feed, error) {
func getFeed(url string, m *sync.Mutex) (*gofeed.Feed, error) {
	m.Lock()
	resp, err := http.Get(url)
	m.Unlock()
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
	urlsFile, err := os.Open("urls2")
	defer urlsFile.Close()
	if err != nil {
		log.Fatal(err)
	}
	// lines := bufio.NewScanner(urlsFile)
	var wg sync.WaitGroup
	var m sync.Mutex
	// for lines.Scan() {
	test, err := url.Parse("https://skuz.xyz")
	if err != nil {
		log.Fatal(err)
	}
	test.Path += "/randomRSS/rss"
	var i int64
	for i = 0; i <= 1_000_000; i++ {
		params := url.Values{}
		params.Add("seed", strconv.FormatInt(i, 10))
		params.Add("items", strconv.FormatInt(i, 10))
		test.RawQuery = params.Encode()
		wg.Add(1)
		go func(v string, wg *sync.WaitGroup, m *sync.Mutex) {

			_, err := getFeed(v, m)
			if err != nil {
				log.Fatalln(v, err)
			} else {
				log.Infof("Fetched feed %s", v)
			}
			wg.Done()
		}(test.String(), &wg, &m)
	}
	wg.Wait()
	end := time.Now()
	log.Infoln("Fetched all in", end.Sub(start))
	// log.Infof("%s per feed\n", end.Sub(start)/time.Duration(len(urls)))
}
