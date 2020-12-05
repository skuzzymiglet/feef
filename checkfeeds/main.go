package main

import (
	"bufio"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/mmcdole/gofeed"
)

func main() {
	log.SetLevel(log.WarnLevel)
	urls := make([]string, 0)
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		if !strings.HasPrefix(sc.Text(), "#") { // respect comments
			urls = append(urls, sc.Text())
		}
	}
	feeds := make(map[string]*gofeed.Feed, len(urls))
	parser := gofeed.NewParser()
	client := &http.Client{}
	guids := make(map[string]struct{})
	for _, u := range urls {
		resp, err := client.Get(u)
		if err != nil {
			log.Warnf("error fetching feed %s: %s", u, err)
			continue
		}
		log.Printf("Fetched %s, status code %d", u, resp.StatusCode)
		feeds[u], err = parser.Parse(resp.Body)
		if err != nil {
			log.Warnf("error parsing feed %s: %s", u, err)
			continue
		}
		log.Printf("Parsed %s, %d items", u, len(feeds[u].Items))
		for _, item := range feeds[u].Items {
			// Check presence of fields
			switch {
			case item.Title == "":
				log.Warnf("Empty title (%s)", u)
			case item.Description == "":
				log.Infof("Empty description (%s)", u)
			case item.Content == "":
				log.Infof("Empty content (%s)", u)
			case item.Link == "":
				log.Infof("Empty link (%s)", u)
			case item.UpdatedParsed == nil && item.PublishedParsed == nil:
				log.Warnf("No updated or published time (%s)", u)
			case item.GUID == "":
				log.Warnf("Empty GUID (%s)", u)
			}
			// Check GUID uniquity
			_, ok := guids[item.GUID]
			if ok {
				log.Warnf("Duplicate GUID %s", item.GUID)
			}
			guids[item.GUID] = struct{}{}
		}
	}
}
