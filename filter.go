package main

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/antonmedv/expr"
	log "github.com/sirupsen/logrus"
)

func Filter(p FilterParam, in <-chan LinkedFeedItem, out chan<- LinkedFeedItem, errChan chan error) {
	var buf []LinkedFeedItem
	var sent int
	for i := range in { // TODO: more specific queries
		output, err := expr.Run(p.matcher, i)
		if err != nil {
			errChan <- err
			return
		}
		matched, ok := output.(bool)
		if !ok {
			errChan <- errors.New("Expression didn't return a bool")
			return
		}
		if p.max != 0 && sent >= p.max {
			return
		}
		if p.sort {
			buf = append(buf, i)
		} else if matched {
			out <- i // NOTE: not recieving from here causes Filter to block!
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
