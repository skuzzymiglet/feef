package main

import (
	"fmt"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
)

func Filter(p FilterParam, in, out chan LinkedFeedItem, errChan chan error) {
	var buf []LinkedFeedItem
	var sent int
	for i := range in { // TODO: more specific queries
		matched := true
		switch {
		case p.item.Match(i.Link):
			log.Debugf("item %s's link matched", i.GUID)
		case p.item.Match(i.Title):
			log.Debugf("item %s's title matched", i.GUID)
		case (i.Author != nil && p.item.Match(i.Author.Name)):
			log.Debugf("item %s's author name matched", i.GUID)
		case (i.Author != nil && p.item.Match(i.Author.Email)):
			log.Debugf("item %s's email matched", i.GUID)
		default:
			matched = false
		}
		if p.max != 0 && sent >= p.max {
			return
		}
		if p.sort {
			buf = append(buf, i)
		} else if matched {
			out <- i
			sent++
		}
	}
	if p.sort {
		log.Debugf("sorting %d items...", len(buf))
		sort.Slice(buf, func(i, j int) bool {
			var it, jt time.Time
			if buf[i].PublishedParsed != nil {
				it = *buf[i].PublishedParsed
			} else if buf[i].UpdatedParsed != nil {
				it = *buf[i].UpdatedParsed
			} else {
				panic(fmt.Sprintf("error sorting feed (from %s): item does not include an update or published time", buf[i].Feed.FeedLink))
			}
			if buf[j].PublishedParsed != nil {
				jt = *buf[j].PublishedParsed
			} else if buf[j].UpdatedParsed != nil {
				jt = *buf[j].UpdatedParsed
			} else {
				panic(fmt.Sprintf("error sorting feed (from %s): item does not include an update or published time", buf[j].Feed.FeedLink))
			}
			return jt.Before(it) // Newest first, so comparator is upside-down
		})
		if p.max == 0 || p.max >= len(buf) {
			for _, v := range buf {
				out <- v
			}
		} else if p.max <= len(buf) {
			for _, v := range buf[:p.max-1] {
				out <- v
			}
		}
	}
}
