package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/dgraph-io/badger"
	"github.com/mmcdole/gofeed"
	"github.com/pkg/profile"

	"github.com/bartmeuris/progressio"
)

type Feeds struct {
	Feeds     []gofeed.Feed
	db        *badger.DB
	updates   chan *gofeed.Feed
	httpMutex *sync.Mutex
}

type _ interface {
	SetURLs([]string) error
	NotifyUpdates() <-chan *gofeed.Feed
	Fetch(string) (<-chan progressio.Progress, error)
	FetchAll() (<-chan struct {
		url string
		p   progressio.Progress
	}, error)
}

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

func (f *Feeds) SetURLS([]string) error {
	return nil
}

func main() {
	defer profile.Start().Stop()
	var f Feeds
	urlsFile, err := os.Open("urls2")
	defer urlsFile.Close()
	if err != nil {
		log.Fatal(err)
	}
	b, err := ioutil.ReadAll(urlsFile)
	strs := strings.Split(string(b), "\n")
	err = f.SetURLS(strs[:len(strs)-2])
	if err != nil {
		log.Fatal(err)
	}
}
